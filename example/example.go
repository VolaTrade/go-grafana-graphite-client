package main

import (
	"time"

	stats "github.com/volatrade/go-grafana-graphite-client"
)

func main() {

	cfg := &stats.Config{
		ConveyorCount:  1,
		FlushTime:      200,
		Url:            "SHEESH",
		ApiKey:         "KEY",
		Env:            "dev_env",
		Service:        "fake_test_service",
		RequestTimeout: 3,
	}
	statz, close, err := stats.NewClient(cfg)

	defer close()
	if err != nil {
		panic(err)
	}

	for {
		statz.Increment("bleh", 1.0)
		time.Sleep(time.Duration(100) * time.Millisecond)
		statz.Increment("bleh", 1.0)
	}
}
