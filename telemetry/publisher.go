package telemetry

import (
	"context"
	"fmt"
	"strings"

	"github.com/hatstand/shinywaffle/wirelesstag"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
	"go.uber.org/zap"
)

type Publisher struct {
	mp metric.MeterProvider

	logger *zap.SugaredLogger
}

func instruments(m map[string]asyncfloat64.Gauge) []instrument.Asynchronous {
	var values []instrument.Asynchronous
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func (p *Publisher) Publish() error {
	// Create gauges for each.
	m := p.mp.Meter("github.com/hatstand/shinywaffle", metric.WithSchemaURL("custom.googleapis.com/shinywaffle"))

	tags, err := wirelesstag.GetTags()
	if err != nil {
		return fmt.Errorf("failed to fetch tag data: %w", err)
	}

	gs := make(map[string]asyncfloat64.Gauge)
	for _, tag := range tags {
		g, err := m.AsyncFloat64().Gauge(strings.ReplaceAll(tag.Name, " ", "_"), instrument.WithUnit("C"))
		if err != nil {
			return fmt.Errorf("failed to create gauge: %w", err)
		}
		gs[tag.Name] = g
	}

	m.RegisterCallback(instruments(gs), func(ctx context.Context) {
		tags, err := wirelesstag.GetTags()
		if err != nil {
			p.logger.Errorf("failed to fetch tag data: %w", err)
			return
		}
		for _, tag := range tags {
			if g, ok := gs[tag.Name]; ok {
				g.Observe(ctx, tag.Temperature)
			}
		}
	})
	return nil
}

func NewPublisher(mp metric.MeterProvider, logger *zap.SugaredLogger) *Publisher {
	return &Publisher{
		mp:     mp,
		logger: logger,
	}
}
