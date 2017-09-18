package main

import (
	"flag"
	"io/ioutil"
	"log"
	"time"

	"github.com/felixge/pidctrl"
	"github.com/golang/protobuf/proto"
	"github.com/hatstand/shinywaffle/control"
	"github.com/hatstand/shinywaffle/wirelesstag"
)

var client = flag.String("client", "", "OAuth client id")
var secret = flag.String("secret", "", "OAuth client secret")
var config = flag.String("config", "config.textproto", "Path to config proto")

const (
	kP = 1
	kI = .5
	kD = .0
)

func main() {
	flag.Parse()

	configText, err := ioutil.ReadFile(*config)
	if err != nil {
		log.Fatalf("Failed to read config file: %s %v", *config, err)
	}
	var config control.Config
	err = proto.UnmarshalText(string(configText), &config)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	m := make(map[string]*pidctrl.PIDController)
	for _, room := range config.Room {
		log.Printf("Configuring controller for: %s", *room.Name)
		ctrl := pidctrl.NewPIDController(kP, kI, kD)
		ctrl.SetOutputLimits(0, 100)
		ctrl.Set(float64(*room.TargetTemperature))
		m[*room.Name] = ctrl
	}

	ch := time.Tick(15 * time.Second)
	lastUpdated := time.Now()
	for _ = range ch {
		tags, err := wirelesstag.GetTags(*client, *secret)
		if err != nil {
			log.Printf("Failed to fetch tag data: %v", err)
		}
		for _, t := range tags {
			ctrl := m[t.Name]
			if ctrl != nil {
				ctrl.UpdateDuration(t.Temperature, time.Since(lastUpdated))
			} else {
				log.Printf("No config for room: %s", t.Name)
			}
		}

		lastUpdated = time.Now()
	}
}
