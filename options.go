package graphite

import "time"

type Option func(GraphiteClient)

func WithFlushInterval(interval time.Duration) Option {
	return func(gc GraphiteClient) {
		gc.flushInterval = interval
	}
}
