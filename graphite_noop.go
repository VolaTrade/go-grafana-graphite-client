package graphite

type GraphiteNoop struct{}

func NewNoop() Stats {
	return &GraphiteNoop{}
}

func (gc *GraphiteNoop) Increment(location string, value float64) {
	return
}

func (gc *GraphiteNoop) Gauge(location string, value float64) {
	return
}
