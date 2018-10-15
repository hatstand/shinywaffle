package control

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"sort"
	"time"

	"github.com/felixge/pidctrl"
	"github.com/golang/protobuf/proto"
	"github.com/hatstand/shinywaffle/wirelesstag"
)

var client = flag.String("client", "", "OAuth client id")
var secret = flag.String("secret", "", "OAuth client secret")

const (
	kP = 1
	kI = .5
	kD = .0
)

type Room struct {
	Pid      *pidctrl.PIDController
	config   *Zone
	schedule *IntervalTree
	LastTemp float64
}

type RadiatorController interface {
	TurnOn([]byte)
	TurnOff([]byte)
}

type Controller struct {
	Config      map[string]*Room
	controller  RadiatorController
	lastUpdated time.Time
}

func NewController(path string, controller RadiatorController) *Controller {
	configText, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read config file: %s %v", path, err)
	}
	var config Config
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
			schedule: NewSchedule(room.GetSchedule()),
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

func (c *Controller) GetNextState(room *Room) HeatingState {
	scheduledTemp := c.checkSchedule(room)
	room.Pid.Set(float64(scheduledTemp))
	value := room.Pid.UpdateDuration(room.LastTemp, time.Since(c.lastUpdated))
	log.Printf("Room: %s Temperature: %.1f Target: %d PID: %f\n", room.config.GetName(), room.LastTemp, scheduledTemp, value)
	if value > 0.0 {
		return HeatingState_ON
	} else {
		return HeatingState_OFF
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
			case HeatingState_OFF, HeatingState_UNKNOWN:
				for _, r := range room.config.Radiator {
					log.Printf("Turning off %s %v", room.config.Name, r.GetAddress())
					c.controller.TurnOff(r.GetAddress())
				}
			case HeatingState_ON:
				for _, r := range room.config.Radiator {
					log.Printf("Turning off %s %v", room.config.Name, r.GetAddress())
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

type ByName []*Zone

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

func (s *Controller) GetZones(ctx context.Context, req *GetZonesRequest) (*GetZonesReply, error) {
	var ret GetZonesReply
	for _, value := range s.Config {
		ret.Zone = append(ret.Zone, value.config)
	}
	sort.Sort(ByName(ret.Zone))
	return &ret, nil
}

func (s *Controller) GetZoneStatus(ctx context.Context, req *GetZoneStatusRequest) (*GetZoneStatusReply, error) {
	for _, r := range s.Config {
		if r.config.GetName() == req.GetName() {
			return &GetZoneStatusReply{
				Name:               r.config.GetName(),
				TargetTemperature:  float32(s.checkSchedule(r)),
				CurrentTemperature: float32(r.LastTemp),
				State:              s.GetNextState(r),
				Schedule:           r.config.GetSchedule(),
			}, nil
		}
	}
	return &GetZoneStatusReply{}, nil
}

func (s *Controller) SetZoneSchedule(ctx context.Context, req *SetZoneScheduleRequest) (*SetZoneScheduleReply, error) {
	return &SetZoneScheduleReply{}, nil
}
