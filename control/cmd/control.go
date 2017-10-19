package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/felixge/pidctrl"
	"github.com/golang/protobuf/proto"
	"github.com/hatstand/shinywaffle/control"
	"github.com/hatstand/shinywaffle/wirelesstag"
)

var client = flag.String("client", "", "OAuth client id")
var secret = flag.String("secret", "", "OAuth client secret")
var config = flag.String("config", "config.textproto", "Path to config proto")
var dryRun = flag.Bool("n", false, "Disables radiator commands")

const (
	kP = 1
	kI = .5
	kD = .0
)

var (
	statusHtml = template.Must(template.New("status.html").Funcs(template.FuncMap{
		"convertColour": convertColour,
	}).ParseFiles("status.html"))
)

type radiatorController interface {
	TurnOn([]byte)
	TurnOff([]byte)
}

type Room struct {
	Pid      *pidctrl.PIDController
	config   *control.Room
	schedule *control.IntervalTree
	LastTemp float64
}

type Controller struct {
	Config     map[string]*Room
	controller radiatorController
}

// convertColour converts a temperature in degrees Celsius into a hue value in the HSV space.
func convertColour(temp float64) int {
	clamped := math.Min(30, math.Max(0, temp)) * 4
	return int(240 + clamped)
}

func NewController(path string, controller radiatorController) *Controller {
	configText, err := ioutil.ReadFile(*config)
	if err != nil {
		log.Fatalf("Failed to read config file: %s %v", *config, err)
	}
	var config control.Config
	err = proto.UnmarshalText(string(configText), &config)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	m := make(map[string]*Room)
	for _, room := range config.Room {
		log.Printf("Configuring controller for: %s", *room.Name)
		ctrl := pidctrl.NewPIDController(kP, kI, kD)
		ctrl.SetOutputLimits(0, 100)
		ctrl.Set(float64(*room.TargetTemperature))
		m[*room.Name] = &Room{
			Pid:      ctrl,
			config:   room,
			schedule: control.NewSchedule(room.Schedule),
		}
	}
	return &Controller{
		Config:     m,
		controller: controller,
	}
}

func (c *Controller) checkSchedule(room *Room) control.Schedule_Interval_State {
	return room.schedule.QueryTime(time.Now())
}

func (c *Controller) GetNextState(room *Room) control.Schedule_Interval_State {
	scheduled := c.checkSchedule(room)
	switch scheduled {
	case control.Schedule_Interval_OFF, control.Schedule_Interval_UNKNOWN:
		return control.Schedule_Interval_OFF
	case control.Schedule_Interval_ON:
		lastUpdated := time.Now()
		value := room.Pid.UpdateDuration(room.LastTemp, time.Since(lastUpdated))
		log.Printf("Room: %s Temperature: %.1f Target: %d PID: %.1f\n", room.config.GetName(), room.LastTemp, room.config.GetTargetTemperature(), value)
		if value < 50.0 {
			return control.Schedule_Interval_OFF
		} else {
			return control.Schedule_Interval_ON
		}
	}
	return control.Schedule_Interval_UNKNOWN
}

func (c *Controller) ControlRadiators(ctx context.Context) {
	ch := time.Tick(15 * time.Second)
	for {
		select {
		case <-ch:
			tags, err := wirelesstag.GetTags(*client, *secret)
			if err != nil {
				log.Printf("Failed to fetch tag data: %v", err)
			}
			for _, t := range tags {
				room := c.Config[t.Name]
				room.LastTemp = t.Temperature
				if room != nil {
					nextState := c.GetNextState(room)
					switch nextState {
					case control.Schedule_Interval_OFF, control.Schedule_Interval_UNKNOWN:
						c.controller.TurnOff(room.config.Radiator.Address)
					case control.Schedule_Interval_ON:
						c.controller.TurnOn(room.config.Radiator.Address)
					}
				} else {
					log.Printf("No config for room: %s", t.Name)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

type stubRadiatorController struct {
}

func (*stubRadiatorController) TurnOn(addr []byte) {
	log.Printf("Turning on radiator: %v\n", addr)
}

func (*stubRadiatorController) TurnOff(addr []byte) {
	log.Printf("Turning off radiator: %v\n", addr)
}

func createRadiatorController() radiatorController {
	if *dryRun {
		return &stubRadiatorController{}
	} else {
		return control.NewController()
	}
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	controller := NewController(*config, createRadiatorController())
	go controller.ControlRadiators(ctx)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, world!")
	})
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Title string
			Ctrl  *Controller
		}{
			"foobar",
			controller,
		}
		err := statusHtml.Execute(w, data)
		if err != nil {
			panic(err)
		}
	})

	srv := &http.Server{Addr: ":8080"}
	go func() {
		log.Println("Listening...")
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down...")
			timeout, httpCancel := context.WithTimeout(ctx, 5*time.Second)
			defer httpCancel()
			srv.Shutdown(timeout)
			return
		case <-ch:
			cancel()
		}
	}
}
