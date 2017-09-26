package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
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

type RadiatorController interface {
	TurnOn([]byte)
	TurnOff([]byte)
}

type Room struct {
	pid    *pidctrl.PIDController
	config *control.Room
}

func ControlRadiators(ctx context.Context, controller RadiatorController) {
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
			pid:    ctrl,
			config: room,
		}
	}

	ch := time.Tick(15 * time.Second)
	lastUpdated := time.Now()
	for {
		select {
		case <-ch:
			tags, err := wirelesstag.GetTags(*client, *secret)
			if err != nil {
				log.Printf("Failed to fetch tag data: %v", err)
			}
			for _, t := range tags {
				room := m[t.Name]
				if room != nil {
					value := room.pid.UpdateDuration(t.Temperature, time.Since(lastUpdated))
					log.Printf("Room: %s Temperature: %.1f Target: %d PID: %.1f\n", t.Name, t.Temperature, *room.config.TargetTemperature, value)
					if value < 50.0 {
						controller.TurnOff(room.config.Radiator.Address)
					} else {
						controller.TurnOn(room.config.Radiator.Address)
					}
				} else {
					log.Printf("No config for room: %s", t.Name)
				}
			}
		case <-ctx.Done():
			return
		}
		lastUpdated = time.Now()
	}
}

type StubController struct {
}

func (*StubController) TurnOn(addr []byte) {
	log.Printf("Turning on radiator: %v\n", addr)
}

func (*StubController) TurnOff(addr []byte) {
	log.Printf("Turning off radiator: %v\n", addr)
}

func createController() RadiatorController {
	if *dryRun {
		return &StubController{}
	} else {
		return control.NewController()
	}
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, world!")
	})

	go ControlRadiators(ctx, createController())

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
