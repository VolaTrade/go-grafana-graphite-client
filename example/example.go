package main

import (
	"time"

	stats "github.com/volatrade/go-grafana-graphite-client"
)

func main() {

	cfg := &stats.Config{
		ConveyorCount:  1,
		FlushTime:      200,
		Url:            "https://graphite-us-central1.grafana.net/metrics",
		ApiKey:         "22241:eyJrIjoiMWE2ODI4MTEyOGQ0YWE2NDNlNjIwOWFiMzk2YzYyMmUwMjU2ZWEwOSIsIm4iOiJWb2xhdHJhZGUiLCJpZCI6NDQ4NTc0fQ==",
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
		statz.Increment("fudge", 1.0)
	}
}
