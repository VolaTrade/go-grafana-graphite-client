# Go Grafana Graphite Client

A Go client for relaying stats to Graphite using simply a url and API key.

## The Why
Sending metrics to Graphite should be extremely simple. Ideally, it should take a developer 5 minutes to setup metric aggregation for any service where they follow the following steps:
1. Import a stats library
2. Configure the library using an api key and a metrics endpoint and some other variables
    -  Url should look something like: `https://<graphite-host>/metrics`
3. Now you can call methods like `stats.Increment()` and metrics will be published to your Graphite server