package incrementer

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/volatrade/go-grafana-graphite-client/internal/commons"
)

type incrementProcessor struct {
	defaultRequest http.Request
	incrementsMux  sync.Mutex
	ctx            context.Context
	client         http.Client
	increments     map[string]*commons.MetricData
}

func NewIncrementsProcessor(ctx context.Context, url string, apiKey string, requestTimeout time.Duration) (*incrementProcessor, error) {
	req, err := commons.GetDefaultRequest(url, apiKey)
	if err != nil {
		return nil, err
	}
	return &incrementProcessor{
		ctx:            ctx,
		client:         http.Client{Timeout: requestTimeout},
		defaultRequest: *req,
		increments:     make(map[string]*commons.MetricData),
		incrementsMux:  sync.Mutex{},
	}, nil
}

func (ip *incrementProcessor) insertIncrement(md *commons.MetricData) {
	ip.incrementsMux.Lock()
	if _, exists := ip.increments[md.Name]; !exists {
		ip.increments[md.Name] = md
	} else {
		ip.increments[md.Name].Value += md.Value
	}
	ip.incrementsMux.Unlock()
}

func (ip *incrementProcessor) ProcessIncrements(incrementChannel chan *commons.MetricData) {
	for {
		select {
		case <-ip.ctx.Done():
			return

		case md := <-incrementChannel:
			ip.insertIncrement(md)
		}
	}
}

func (ic *incrementProcessor) RunIncrementTransportRoutine(dispatchInterval time.Duration) {
	ticker := time.NewTicker(dispatchInterval)
	defer ticker.Stop()

	for {
		select {

		case <-ic.ctx.Done():
			log.Println("[GO-STATSD] received kill signal from context in dispatch routine")

		case <-ticker.C:
			go ic.dispatchIncrementsToGrafana()
		}
	}
}

func (ic *incrementProcessor) dispatchIncrementsToGrafana() {
	ic.incrementsMux.Lock()
	incCount := len(ic.increments)
	incrementsSlice := make(commons.MetricDataSlice, incCount)

	if incCount == 0 {
		ic.incrementsMux.Unlock()
		return
	}
	i := 0
	for _, val := range ic.increments {
		incrementsSlice[i] = val
		i++
	}
	ic.increments = make(map[string]*commons.MetricData)
	ic.incrementsMux.Unlock()

	mdArrBytes, err := json.Marshal(incrementsSlice)
	if err != nil {
		return
	}

	request := ic.defaultRequest.Clone(context.Background())
	request.Body = ioutil.NopCloser(bytes.NewBuffer(mdArrBytes))

	_, err = ic.client.Do(request)
	if err != nil {
		println(err.Error())
		return
	}
	return
}
