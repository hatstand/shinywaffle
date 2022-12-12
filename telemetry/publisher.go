package telemetry

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
)

type Publisher struct {
	mp metric.MeterProvider

	tempMetrics map[string]asyncfloat64.Gauge
}

func (p *Publisher) Publish(ctx context.Context, name string, temp float64, on bool) error {
	name = strings.ReplaceAll(name, " ", "_")
	g := p.tempMetrics[name]
	if g == nil {
		m, err := p.mp.Meter("shinywaffle").AsyncFloat64().Gauge(name)
		if err != nil {
			return fmt.Errorf("failed to create gauge: %w", err)
		}
		p.tempMetrics[name] = m
		g = m
	}
	g.Observe(ctx, temp, attribute.Bool("on", on))
	return nil
}

func NewPublisher(mp metric.MeterProvider) *Publisher {
	return &Publisher{
		mp:          mp,
		tempMetrics: make(map[string]asyncfloat64.Gauge),
	}
}
