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
	ctx            context.Context
	defaultRequest http.Request
	mux            sync.Mutex
	length         uint32
	client         http.Client
	increments     map[string]*commons.MetricData
}

//LaunchIncrementsProcessor constructs incrementProcessor struct to then run increment processing routines
func NewIncrementsProcessor(ctx context.Context, url string, apiKey string,
	requestTimeout time.Duration, incrementChannel chan *commons.MetricData) (*incrementProcessor, error) {

	req, err := commons.GetDefaultRequest(url, apiKey)
	if err != nil {
		return nil, err
	}
	ip := &incrementProcessor{
		ctx:            ctx,
		client:         http.Client{Timeout: requestTimeout},
		length:         0,
		defaultRequest: *req,
		increments:     make(map[string]*commons.MetricData),
		mux:            sync.Mutex{},
	}
	go ip.processIncrements(incrementChannel)
	return ip, nil
}

func (ip *incrementProcessor) insertIncrement(md *commons.MetricData) {

	if _, exists := ip.increments[md.Name]; !exists {
		ip.length += 1
		ip.increments[md.Name] = md
	} else {
		ip.increments[md.Name].Value += md.Value
	}
}

func (ip *incrementProcessor) processIncrements(incrementChannel chan *commons.MetricData) {
	for {
		select {

		case <-ip.ctx.Done():
			log.Println("[GRAPHITE-CLIENT] received kill signal from context in increment processing routine")
			return

		case md := <-incrementChannel:
			ip.mux.Lock()
			ip.insertIncrement(md)
			ip.mux.Unlock()
		}
	}
}

func (ip *incrementProcessor) GraphiteDispatchRoutine() {

	ip.mux.Lock()

	if ip.length == 0 {
		ip.mux.Unlock()
		return
	}
	incrementsSlice := make(commons.MetricDataSlice, len(ip.increments))
	i := 0
	for _, val := range ip.increments {
		incrementsSlice[i] = val
		i++
	}
	ip.increments = make(map[string]*commons.MetricData)
	ip.mux.Unlock()

	mdArrBytes, err := json.Marshal(incrementsSlice)
	if err != nil {
		return
	}

	request := ip.defaultRequest.Clone(context.Background())
	request.Body = ioutil.NopCloser(bytes.NewBuffer(mdArrBytes))

	_, err = ip.client.Do(request)
	if err != nil {
		println(err.Error())
		return
	}
	return
}
