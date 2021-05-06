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
	ctx           context.Context
	client        http.Client
	defRequest    http.Request
	mdArray       commons.MetricDataSlice
	metricChannel chan *commons.MetricData
	arrayMux      sync.Mutex
}

func NewConveyor(ctx context.Context, grafanaUrl string,
	apiKey string, requestTimeout time.Duration, mc chan *commons.MetricData) (*conveyor, error) {

	req, err := commons.GetDefaultRequest(grafanaUrl, apiKey)
	if err != nil {
		return nil, err
	}
	c := &conveyor{
		ctx:           ctx,
		client:        http.Client{Timeout: requestTimeout},
		defRequest:    *req,
		mdArray:       commons.MetricDataSlice{},
		metricChannel: mc,
		arrayMux:      sync.Mutex{},
	}

	go c.runTransportRoutine()
	return c, nil
}

func (c *conveyor) runTransportRoutine() {

	for {
		select {

		case <-c.ctx.Done():
			log.Println("[GO-STATSD] received kill signal from context in transport routine")

		case md := <-c.metricChannel:
			c.arrayMux.Lock()
			c.mdArray = append(c.mdArray, md)
			c.arrayMux.Unlock()
		}
	}
}

func (c *conveyor) DispatchMetricsToGrafanaCloud() {

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
	}
}
