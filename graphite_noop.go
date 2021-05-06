package graphite

type NoopClient struct {
}

//NewNoop returns noop client that can be used for testing
func NewNoop() Stats {
	return &NoopClient{}
}

//Increment ...
func (nc *NoopClient) Increment(location string, value float64) {
	return
}

//Gauge ...
func (nc *NoopClient) Gauge(location string, value float64) {
	return
}
