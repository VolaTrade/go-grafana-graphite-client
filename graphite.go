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
		ConveyorCount  int
		FlushTime      int
		Url            string
		ApiKey         string
		Env            string
		Service        string
		RequestTimeout int
	}
	GraphiteClient struct {
		ctx             context.Context
		config          *Config
		prefix          string
		conveyorChannel chan *commons.MetricData
	}
)

func NewClient(cfg *Config) (Stats, func(), error) {

	if cfg.FlushTime < 0 {
		return nil, func() {}, errors.New("flush time should be >= 0")
	}
	if cfg.ConveyorCount <= 0 {
		return nil, func() {}, errors.New("conveyor count should be >= 0")
	}
	if !strings.Contains(cfg.Url, "http") && !strings.Contains(cfg.Url, "https") {
		return nil, func() {}, errors.New("Unsupported protocol provided in URL")
	}

	ctx, cancel := context.WithCancel(context.Background())
	incChan := make(chan *commons.MetricData)

	client := GraphiteClient{
		ctx:             ctx,
		config:          cfg,
		conveyorChannel: make(chan *commons.MetricData),
		prefix:          strings.ToLower(fmt.Sprintf("%s.%s", cfg.Env, cfg.Service)),
	}

	end := func() {
		cancel()
		close(client.conveyorChannel)
		close(incChan)
	}

	incProcessor, err := incrementer.NewIncrementsProcessor(ctx, cfg.Url,
		cfg.ApiKey, time.Duration(cfg.RequestTimeout)*time.Second)

	if err != nil {
		return nil, end, err
	}

	for i := 0; i < cfg.ConveyorCount; i++ {
		launchConveyor(ctx, cfg.Url, cfg.ApiKey,
			cfg.RequestTimeout, cfg.FlushTime, client.conveyorChannel, incChan)
	}

	go incProcessor.ProcessIncrements(incChan)
	go incProcessor.RunIncrementTransportRoutine(time.Duration(cfg.FlushTime) * time.Millisecond)

	return &client, end, nil
}

func launchConveyor(ctx context.Context, url string,
	apiKey string, timeout int, flushTime int, ch chan *commons.MetricData, incChan chan *commons.MetricData) error {
	conveyor, err := conveyor.NewConveyor(
		ctx,
		url,
		apiKey,
		time.Duration(timeout)*time.Second,
	)

	if err != nil {
		return err
	}

	go conveyor.RunTransportRoutine(ch, incChan)
	go conveyor.RunMetricsRoutine(time.Duration(flushTime) * time.Millisecond)
	return nil
}

func (gc *GraphiteClient) Increment(location string, value float64) {

	metricPoint := commons.CreatePoint(
		fmt.Sprintf("%s.%s", gc.prefix, location),
		1, commons.Counter, value,
		time.Now().Unix(),
	)
	gc.conveyorChannel <- metricPoint
}

func (gc *GraphiteClient) Gauge(location string, value float64) {

	metricPoint := commons.CreatePoint(
		fmt.Sprintf("%s.%s", gc.prefix, location),
		10, commons.Gauge, value,
		time.Now().Unix(),
	)
	gc.conveyorChannel <- metricPoint
}
