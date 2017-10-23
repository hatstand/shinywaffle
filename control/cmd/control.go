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
	"sort"
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
		"convertColour":  convertColour,
		"renderSchedule": renderSchedule,
	}).ParseFiles("status.html"))
)

type radiatorController interface {
	TurnOn([]byte)
	TurnOff([]byte)
}

type Room struct {
	Pid      *pidctrl.PIDController
	config   *control.Zone
	schedule *control.IntervalTree
	LastTemp float64
}

type Controller struct {
	Config      map[string]*Room
	controller  radiatorController
	lastUpdated time.Time
}

// convertColour converts a temperature in degrees Celsius into a hue value in the HSV space.
func convertColour(temp float64) int {
	clamped := math.Min(30, math.Max(0, temp)) * 4
	return int(240 + clamped)
}

type Interval struct {
	Width  int // Percentage from 0-100 of 24 hours
	Offset int // Percentage from 0-100 of 24 hours
}

func renderSchedule(s *control.Schedule) []Interval {
	var ret []Interval
	for _, i := range s.Interval {
		begin := i.Begin.GetHour()*60 + i.Begin.GetMinute()
		end := i.End.GetHour()*60 + i.End.GetMinute()
		ret = append(ret, Interval{
			Width:  int(float32(end-begin) / 24 / 60 * 100),
			Offset: int(float32(begin) / 24 / 60 * 100),
		})
	}
	return ret
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
	for _, room := range config.Zone {
		log.Printf("Configuring controller for: %s", room.GetName())
		ctrl := pidctrl.NewPIDController(kP, kI, kD)
		ctrl.SetOutputLimits(0, 100)
		m[room.GetName()] = &Room{
			Pid:      ctrl,
			config:   room,
			schedule: control.NewSchedule(room.GetSchedule()),
		}
	}
	return &Controller{
		Config:     m,
		controller: controller,
	}
}

func (c *Controller) checkSchedule(room *Room) int32 {
	return room.schedule.QueryTime(time.Now())
}

func (c *Controller) GetNextState(room *Room) control.HeatingState {
	scheduledTemp := c.checkSchedule(room)
	room.Pid.Set(float64(scheduledTemp))
	value := room.Pid.UpdateDuration(room.LastTemp, time.Since(c.lastUpdated))
	log.Printf("Room: %s Temperature: %.1f Target: %d PID: %f\n", room.config.GetName(), room.LastTemp, scheduledTemp, value)
	if value > 0.0 {
		return control.HeatingState_ON
	} else {
		return control.HeatingState_OFF
	}
}

func (c *Controller) tick() {
	tags, err := wirelesstag.GetTags(*client, *secret)
	if err != nil {
		log.Printf("Failed to fetch tag data: %v", err)
		return
	}
	c.lastUpdated = time.Now()
	for _, t := range tags {
		room := c.Config[t.Name]
		room.LastTemp = t.Temperature
		if room != nil {
			nextState := c.GetNextState(room)
			switch nextState {
			case control.HeatingState_OFF, control.HeatingState_UNKNOWN:
				for _, r := range room.config.Radiator {
					c.controller.TurnOff(r.GetAddress())
				}
			case control.HeatingState_ON:
				for _, r := range room.config.Radiator {
					c.controller.TurnOn(r.GetAddress())
				}
			}
		} else {
			log.Printf("No config for room: %s", t.Name)
		}
	}
}

func (c *Controller) ControlRadiators(ctx context.Context) {
	ch := time.Tick(1 * time.Minute)
	c.tick()
	for {
		select {
		case <-ch:
			c.tick()
		case <-ctx.Done():
			return
		}
	}
}

type ByName []*control.Zone

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

func (s *Controller) GetZones(ctx context.Context, req *control.GetZonesRequest) (*control.GetZonesReply, error) {
	var ret control.GetZonesReply
	for _, value := range s.Config {
		ret.Zone = append(ret.Zone, value.config)
	}
	sort.Sort(ByName(ret.Zone))
	return &ret, nil
}

func (s *Controller) GetZoneStatus(ctx context.Context, req *control.GetZoneStatusRequest) (*control.GetZoneStatusReply, error) {
	for _, r := range s.Config {
		if r.config.GetName() == req.GetName() {
			return &control.GetZoneStatusReply{
				Name:               r.config.GetName(),
				TargetTemperature:  float32(s.checkSchedule(r)),
				CurrentTemperature: float32(r.LastTemp),
				State:              s.GetNextState(r),
				Schedule:           r.config.GetSchedule(),
			}, nil
		}
	}
	return &control.GetZoneStatusReply{}, nil
}

func (s *Controller) SetZoneSchedule(ctx context.Context, req *control.SetZoneScheduleRequest) (*control.SetZoneScheduleReply, error) {
	return &control.SetZoneScheduleReply{}, nil
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
		var ret []*control.GetZoneStatusReply
		zones, err := controller.GetZones(ctx, &control.GetZonesRequest{})
		if err == nil {
			for _, z := range zones.Zone {
				status, err := controller.GetZoneStatus(ctx, &control.GetZoneStatusRequest{
					Name: z.GetName(),
				})
				if err == nil {
					ret = append(ret, status)
				}
			}
		}
		data := struct {
			Title string
			Now   time.Time
			Zones []*control.GetZoneStatusReply
			Error error
		}{
			"foobar",
			time.Now(),
			ret,
			err,
		}
		err = statusHtml.Execute(w, data)
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
