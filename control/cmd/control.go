package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/hatstand/shinywaffle/control"
	"github.com/hatstand/shinywaffle/weather"
	"google.golang.org/grpc"
)

var config = flag.String("config", "config.textproto", "Path to config proto")
var dryRun = flag.Bool("n", false, "Disables radiator commands")
var port = flag.Int("port", 8081, "Status port")
var grpcPort = flag.Int("grpc", 8082, "GRPC service port")

var (
	statusHtml = template.Must(template.New("status.html").Funcs(template.FuncMap{
		"convertColour":  convertColour,
		"renderSchedule": renderSchedule,
	}).ParseFiles("status.html", "weather.html"))
)

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

type stubRadiatorController struct {
}

func (*stubRadiatorController) TurnOn(addr []byte) {
	log.Printf("Turning on radiator: %v\n", addr)
}

func (*stubRadiatorController) TurnOff(addr []byte) {
	log.Printf("Turning off radiator: %v\n", addr)
}

func createRadiatorController() control.RadiatorController {
	if *dryRun {
		return &stubRadiatorController{}
	} else {
		return control.NewRadioController()
	}
}

type ServeMux struct {
	api http.Handler
	ui  http.Handler
}

func NewServeMux(api http.Handler, ui http.Handler) *ServeMux {
	return &ServeMux{
		api: api,
		ui:  ui,
	}
}

func (s *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/v1") {
		s.api.ServeHTTP(w, r)
	} else {
		s.ui.ServeHTTP(w, r)
	}
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	controller := control.NewController(*config, createRadiatorController())
	go controller.ControlRadiators(ctx)

	s := grpc.NewServer()
	control.RegisterHeatingControlServiceServer(s, controller)

	l, err := net.Listen("tcp", ":"+strconv.Itoa(*grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on GRPC port: %v", err)
	}
	go s.Serve(l)

	apiMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err = control.RegisterHeatingControlServiceHandlerFromEndpoint(ctx, apiMux, ":8081", opts)
	if err != nil {
		log.Fatalf("Error starting GRPC gateway: %v", err)
	}

	uiMux := http.NewServeMux()
	uiMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, world!")
	})
	uiMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})
	uiMux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
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
		weath, err := weather.FetchCurrentWeather("London")
		if err != nil {
			log.Printf("Failed to fetch current weather: %v", err)
			weath = nil
		}
		data := struct {
			Title   string
			Now     time.Time
			Zones   []*control.GetZoneStatusReply
			Error   error
			Weather *weather.Observation
		}{
			"foobar",
			time.Now(),
			ret,
			err,
			weath,
		}
		err = statusHtml.Execute(w, data)
		if err != nil {
			panic(err)
		}
	})

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(*port),
		Handler: NewServeMux(apiMux, uiMux),
	}
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
