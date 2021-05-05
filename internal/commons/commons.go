package commons

import (
	"net/http"
	"net/url"
)

func GetDefaultRequest(grafanaUrl string, apiKey string) (*http.Request, error) {
	url, err := url.Parse(grafanaUrl)

	if err != nil {
		return nil, err
	}

	req := http.Request{
		Method: http.MethodPost,
		URL:    url,
		Header: http.Header{},
	}

	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Add("Content-Type", "application/json")
	return &req, nil
}

// createPoint creates a datapoint, i.e. a metricData structure, and makes sure the id is set.
func CreatePoint(name string, interval int, mtype string,
	val float64, time int64) *MetricData {
	md := MetricData{
		Name:     name,       // in graphite style format. should be same as Metric field below (used for partitioning, schema matching, indexing)
		Metric:   name,       // in graphite style format. should be same as Name field above (used to generate Id)
		Interval: interval,   // the resolution of the metric
		Value:    val,        // float64 value
		Unit:     "",         // not needed or used yet
		Time:     time,       // unix timestamp in seconds
		Mtype:    mtype,      // not used yet. but should be one of gauge rate count counter timestamp
		Tags:     []string{}, // not needed or used yet. can be single words or key=value pairs
	}
	md.SetId()
	return &md
}
