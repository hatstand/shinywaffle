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

	"cloud.google.com/go/logging"
	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/coreos/go-systemd/daemon"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/hatstand/shinywaffle/calendar"
	"github.com/hatstand/shinywaffle/control"
	"github.com/hatstand/shinywaffle/telemetry"
	"github.com/hatstand/shinywaffle/weather"
	"github.com/jonstaryuk/gcloudzap"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	octrace "go.opencensus.io/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bridge/opencensus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

var config = flag.String("config", "config.textproto", "Path to config proto")
var dryRun = flag.Bool("n", false, "Disables radiator commands")
var port = flag.Int("port", 8081, "Status port")
var grpcPort = flag.Int("grpc", 8082, "GRPC service port")

var (
	statusHtml = template.Must(template.New("status.html").Funcs(template.FuncMap{
		"convertColour": convertColour,
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

func createLogger(ctx context.Context) (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	cfg.EncoderConfig.LevelKey = "severity"
	cfg.EncoderConfig.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		switch l {
		case zapcore.DebugLevel:
			enc.AppendString("DEBUG")
		case zapcore.InfoLevel:
			enc.AppendString("INFO")
		case zapcore.WarnLevel:
			enc.AppendString("WARNING")
		case zapcore.ErrorLevel:
			enc.AppendString("ERROR")
		case zapcore.DPanicLevel:
			enc.AppendString("CRITICAL")
		case zapcore.PanicLevel:
			enc.AppendString("ALERT")
		case zapcore.FatalLevel:
			enc.AppendString("EMERGENCY")
		}
	}
	cfg.EncoderConfig.EncodeDuration = zapcore.MillisDurationEncoder
	cfg.EncoderConfig.NameKey = "logger"
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.MessageKey = "message"
	cfg.EncoderConfig.StacktraceKey = "stacktrace"

	c, err := logging.NewClient(ctx, "projects/shinywaffle-1540815179440")
	if err != nil {
		return nil, fmt.Errorf("failed to create logging client: %w", err)
	}

	logger, err := gcloudzap.New(cfg, c, "muddy-pond")
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	return logger.Sugar(), nil
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := createLogger(ctx)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	exporter, err := texporter.New(
		texporter.WithProjectID("shinywaffle-1540815179440"),
		// Disable telemetry on the exporter client otherwise it will trace itself!
		texporter.WithTraceClientOptions([]option.ClientOption{option.WithTelemetryDisabled()}),
	)
	if err != nil {
		logger.Fatalf("failed to create the Cloud Trace exporter: %v", err)
	}

	res, err := resource.New(ctx,
		// Keep the default detectors
		resource.WithTelemetrySDK(),
		// Add your own custom attributes to identify your application
		resource.WithAttributes(
			semconv.ServiceNameKey.String("shinywaffle"),
		),
	)
	if err != nil {
		logger.Fatalf("resource.New: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	defer tp.ForceFlush(ctx)
	otel.SetTracerProvider(tp)
	// Export legacy OpenCensus spans generated by Google API Client Libraries.
	octrace.ApplyConfig(octrace.Config{DefaultSampler: octrace.AlwaysSample()})
	octrace.DefaultTracer = opencensus.NewTracer(tp.Tracer("simple"))

	metricsExporter, err := mexporter.New(
		mexporter.WithProjectID("shinywaffle-1540815179440"),
	)
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricsExporter, sdkmetric.WithInterval(time.Minute))),
		sdkmetric.WithResource(res),
	)
	defer mp.ForceFlush(ctx)

	monitoringMux := http.NewServeMux()
	monitoringMux.Handle("/metrics", promhttp.Handler())
	monitoringSrv := &http.Server{
		Addr:    ":2112",
		Handler: monitoringMux,
	}
	go func() {
		if err := monitoringSrv.ListenAndServe(); err != nil {
			logger.Fatalf("failed to run monitoring service: %v", err)
		}
	}()

	telemetry := telemetry.NewPublisher(mp)

	calendarService, err := calendar.NewCalendarScheduleService()
	if err != nil {
		logger.Fatalf("Failed to start calendar service: %v", err)
	}

	controller := control.NewController(*config, createRadiatorController(), calendarService, telemetry)
	go controller.ControlRadiators(ctx)

	s := grpc.NewServer()
	control.RegisterHeatingControlServiceServer(s, controller)

	l, err := net.Listen("tcp", ":"+strconv.Itoa(*grpcPort))
	if err != nil {
		logger.Fatalf("Failed to listen on GRPC port: %v", err)
	}
	go s.Serve(l)

	apiMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err = control.RegisterHeatingControlServiceHandlerFromEndpoint(ctx, apiMux, ":8081", opts)
	if err != nil {
		logger.Fatalf("Error starting GRPC gateway: %v", err)
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
			logger.Warnf("Failed to fetch current weather: %v", err)
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
		logger.Info("Listening...")
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(*port))
		if err != nil {
			logger.Fatalf("Failed to listen on port: %v", *port)
		}
		go func() {
			// Tells systemd that requests can now be served.
			daemon.SdNotify(false, daemon.SdNotifyReady)
			for {
				// Watchdog check.
				resp, err := http.Get("http://127.0.0.1:" + strconv.Itoa(*port))
				if err == nil {
					daemon.SdNotify(false, daemon.SdNotifyWatchdog)
				}
				resp.Body.Close()
				time.Sleep(5 * time.Second)
			}
		}()
		if err := srv.Serve(ln); err != nil {
			logger.Error(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Shutting down...")
			timeout, httpCancel := context.WithTimeout(ctx, 5*time.Second)
			defer httpCancel()
			srv.Shutdown(timeout)
			return
		case <-ch:
			cancel()
		}
	}
}
