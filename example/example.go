package main

import (
	"os"
	"time"

	stats "github.com/volatrade/go-grafana-graphite-client"
)

func main() {

	cfg := &stats.Config{
		Url:     "https://graphite-us-central1.grafana.net/metrics",
		ApiKey:  os.Getenv("GRAPHITE_API_KEY"),
		Env:     "test_me",
		Service: "fake_test_service",
	}
	statz, close, err := stats.NewClient(cfg)

	defer close()
	if err != nil {
		panic(err)
	}

	for {
		println("Starting increment")
		for i := 0; i < 10; i++ {
			statz.Increment("bug", 1.0)
		}
		time.Sleep(time.Duration(100) * time.Millisecond)

	}
}
