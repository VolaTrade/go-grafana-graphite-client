package conveyor

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"log"

	"github.com/volatrade/go-grafana-graphite-client/internal/commons"
)

type conveyor struct {
	ctx        context.Context
	client     http.Client
	defRequest http.Request
	mdArray    commons.MetricDataSlice
	arrayMux   sync.Mutex
}

func NewConveyor(ctx context.Context, grafanaUrl string, apiKey string, requestTimeout time.Duration) (*conveyor, error) {

	req, err := commons.GetDefaultRequest(grafanaUrl, apiKey)
	if err != nil {
		return nil, err
	}

	return &conveyor{
		ctx:        ctx,
		client:     http.Client{Timeout: requestTimeout},
		defRequest: *req,
		mdArray:    commons.MetricDataSlice{},
		arrayMux:   sync.Mutex{},
	}, nil
}

func (c *conveyor) RunTransportRoutine(conveyorChannel chan *commons.MetricData, incrementChannel chan *commons.MetricData) {

	for {
		select {

		case <-c.ctx.Done():
			log.Println("[GO-STATSD] received kill signal from context in transport routine")

		case md := <-conveyorChannel:
			switch md.Mtype {

			case commons.Counter:
				incrementChannel <- md
			default:
				c.arrayMux.Lock()
				c.mdArray = append(c.mdArray, md)
				c.arrayMux.Unlock()
			}
		}
	}
}

func (c *conveyor) RunMetricsRoutine(dispatchInterval time.Duration) {
	ticker := time.NewTicker(dispatchInterval)
	defer ticker.Stop()

	for {
		select {

		case <-c.ctx.Done():
			log.Println("[GO-STATSD] received kill signal from context in dispatch routine")

		case <-ticker.C:
			c.dispatchMetricsToGrafanaCloud()
		}
	}
}

func (c *conveyor) dispatchMetricsToGrafanaCloud() {

	if len(c.mdArray) == 0 {
		return
	}
	c.arrayMux.Lock()
	mdArrBytes, err := json.Marshal(c.mdArray)
	if err != nil {
		c.arrayMux.Unlock()
		return
	}
	c.mdArray = commons.MetricDataSlice{}
	c.arrayMux.Unlock()

	request := c.defRequest.Clone(context.Background())
	request.Body = ioutil.NopCloser(bytes.NewBuffer(mdArrBytes))

	_, err = c.client.Do(request)
	if err != nil {
		println(err.Error())
		return
	}
	return
}
