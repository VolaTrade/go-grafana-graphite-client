package graphite

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/volatrade/go-grafana-graphite-client/internal/commons"
	"github.com/volatrade/go-grafana-graphite-client/internal/conveyor"
	"github.com/volatrade/go-grafana-graphite-client/internal/incrementer"
)

type (
	Stats interface {
		Increment(string, float64)
		Gauge(string, float64)
	}
	Config struct {
		Url     string
		ApiKey  string
		Env     string
		Service string
	}
	GraphiteClient struct {
		conveyorChannel   chan *commons.MetricData
		incrementsChannel chan *commons.MetricData
		config            *Config
		ctx               context.Context
		flushInterval     time.Duration
		prefix            string
		requestTimeout    time.Duration
	}
)

//NewClient returns a graphite client interface
func NewClient(cfg *Config, opts ...Option) (Stats, func(), error) {

	if cfg == nil {
		return nil, func() {}, errors.New("Config should not be nil")
	}
	if !strings.Contains(cfg.Url, "http") && !strings.Contains(cfg.Url, "https") {
		return nil, func() {}, errors.New("Unsupported protocol provided in URL")
	}

	ctx, cancel := context.WithCancel(context.Background())

	incChan, convChan := make(chan *commons.MetricData), make(chan *commons.MetricData)

	client := GraphiteClient{
		ctx:               ctx,
		config:            cfg,
		conveyorChannel:   convChan,
		incrementsChannel: incChan,
		flushInterval:     time.Duration(800) * time.Millisecond,
		prefix:            strings.ToLower(fmt.Sprintf("%s.%s", cfg.Env, cfg.Service)),
		requestTimeout:    time.Duration(3) * time.Second,
	}

	for _, opt := range opts {
		opt(client)
	}

	end := func() {
		cancel()
		close(convChan)
		close(incChan)
	}

	ip, err := incrementer.NewIncrementsProcessor(ctx, cfg.Url, cfg.ApiKey, client.requestTimeout, incChan)
	if err != nil {
		return nil, func() {}, err
	}
	c, err := conveyor.NewConveyor(ctx, cfg.Url, cfg.ApiKey, client.requestTimeout, convChan)
	if err != nil {
		return nil, func() {}, err
	}
	go commons.RunPoller(ctx, client.flushInterval, c.DispatchMetricsToGrafanaCloud, ip.GraphiteDispatchRoutine)
	if err != nil {
		return nil, end, err
	}
	return &client, end, nil
}

//Increment performs an increment operation
func (gc *GraphiteClient) Increment(location string, value float64) {

	metricPoint := commons.CreatePoint(
		fmt.Sprintf("%s.%s", gc.prefix, location),
		1, commons.Counter, value,
		time.Now().Unix(),
	)
	gc.incrementsChannel <- metricPoint
}

//Gauge performs a gauge operation
func (gc *GraphiteClient) Gauge(location string, value float64) {

	metricPoint := commons.CreatePoint(
		fmt.Sprintf("%s.%s", gc.prefix, location),
		10, commons.Gauge, value,
		time.Now().Unix(),
	)
	gc.conveyorChannel <- metricPoint
}
