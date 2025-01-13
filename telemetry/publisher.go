package telemetry

import (
	"context"
	"fmt"
	"strings"

	"github.com/hatstand/shinywaffle/wirelesstag"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

type Publisher struct {
	mp metric.MeterProvider

	logger *zap.SugaredLogger
}

func instruments(m map[string]metric.Float64ObservableGauge) []metric.Observable {
	var values []metric.Observable
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

	gs := make(map[string]metric.Float64ObservableGauge)
	for _, tag := range tags {
		g, err := m.Float64ObservableGauge(strings.ReplaceAll(tag.Name, " ", "_"), metric.WithUnit("C"))
		if err != nil {
			return fmt.Errorf("failed to create gauge: %w", err)
		}
		gs[tag.Name] = g
	}

	m.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		tags, err := wirelesstag.GetTags()
		if err != nil {
			p.logger.Errorf("failed to fetch tag data: %v", err)
			return fmt.Errorf("failed to fetch tag data: %w", err)
		}
		for _, tag := range tags {
			if g, ok := gs[tag.Name]; ok {
				o.ObserveFloat64(g, tag.Temperature)
			}
		}
		return nil
	}, instruments(gs)...)
	return nil
}

func NewPublisher(mp metric.MeterProvider, logger *zap.SugaredLogger) *Publisher {
	return &Publisher{
		mp:     mp,
		logger: logger,
	}
}
