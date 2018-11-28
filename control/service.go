package control

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"time"

	"github.com/felixge/pidctrl"
	"github.com/golang/protobuf/proto"
	"github.com/hatstand/shinywaffle/calendar"
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
	LastTemp float64
}

type RadiatorController interface {
	TurnOn([]byte)
	TurnOff([]byte)
}

type Controller struct {
	Config          map[string]*Room
	controller      RadiatorController
	lastUpdated     time.Time
	calendarService *calendar.CalendarScheduleService
	statusPublisher StatusPublisher
}

type StatusPublisher interface {
	Publish(string, float64, bool) error
}

func NewController(
	path string,
	controller RadiatorController,
	calendarService *calendar.CalendarScheduleService,
	statusPublisher StatusPublisher) *Controller {
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
			Pid:    ctrl,
			config: room,
		}
	}
	return &Controller{
		Config:          m,
		controller:      controller,
		calendarService: calendarService,
		statusPublisher: statusPublisher,
	}
}

func (c *Controller) checkSchedule(room *Room) (int32, error) {
	on, err := c.calendarService.GetSchedule(room.config.CalendarId)
	if err != nil {
		return -1, fmt.Errorf("Failed to fetch schedule for room %s: %v", room.config.Name, err)
	}
	now := time.Now()
	for _, period := range on {
		start, err := time.Parse(time.RFC3339, period.Start)
		if err != nil {
			log.Printf("Failed to parse time %s: %v", period.Start, err)
			continue
		}
		end, err := time.Parse(time.RFC3339, period.End)
		if err != nil {
			log.Printf("Failed to parse time %s: %v", period.End, err)
			continue
		}

		if now.After(start) && now.Before(end) {
			return room.config.TargetTemperature, nil
		}
	}
	return -1, nil
}

func (c *Controller) GetNextState(room *Room) HeatingState {
	scheduledTemp, err := c.checkSchedule(room)
	if err != nil {
		log.Printf("Failed to get schedule for room %s: %v", room.config.Name, err)
		return HeatingState_OFF
	}
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
				go func() {
					c.statusPublisher.Publish(room.config.Name, room.LastTemp, false)
				}()
			case HeatingState_ON:
				for _, r := range room.config.Radiator {
					log.Printf("Turning off %s %v", room.config.Name, r.GetAddress())
					c.controller.TurnOn(r.GetAddress())
				}
				go func() {
					c.statusPublisher.Publish(room.config.Name, room.LastTemp, true)
				}()
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
			target, err := s.checkSchedule(r)
			if err != nil {
				return &GetZoneStatusReply{
					Name:               r.config.GetName(),
					CurrentTemperature: float32(r.LastTemp),
					State:              s.GetNextState(r),
				}, fmt.Errorf("Failed to get current target temp for request %+v: %v", req, err)
			}
			return &GetZoneStatusReply{
				Name:               r.config.GetName(),
				TargetTemperature:  float32(target),
				CurrentTemperature: float32(r.LastTemp),
				State:              s.GetNextState(r),
			}, nil
		}
	}
	return &GetZoneStatusReply{}, nil
}

func (s *Controller) SetZoneSchedule(ctx context.Context, req *SetZoneScheduleRequest) (*SetZoneScheduleReply, error) {
	return &SetZoneScheduleReply{}, nil
}
