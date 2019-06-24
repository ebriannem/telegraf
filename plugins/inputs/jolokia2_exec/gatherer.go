package jolokia2_exec

import (
	"fmt"
	"strings"
	"encoding/json"
	"github.com/influxdata/telegraf"
)

const defaultFieldName = "value"

type Gatherer struct {
	metrics  []Metric
	requests []ExecRequest
}

func NewGatherer(metrics []Metric) *Gatherer {
	return &Gatherer{
		metrics:  metrics,
		requests: makeExecRequests(metrics),
	}
}

// Gather adds points to an accumulator from responses returned
// by a Jolokia agent.
func (g *Gatherer) Gather(client *Client, acc telegraf.Accumulator) error {
	var tags map[string]string

	if client.config.ProxyConfig != nil {
		tags = map[string]string{"jolokia_proxy_url": client.URL}
	} else {
		tags = map[string]string{"jolokia_agent_url": client.URL}
	}

    requests := makeExecRequests(g.metrics)
    responses, err := client.exec(requests)
	if err != nil {
		return err
	}

    for _, response := range responses {

    	val := strings.Replace(response.Value, "\\", "", -1) 
        // DEBUG
        // fmt.Printf("%v", val)
    	executionDetails := make(map[string]interface{})

    	if err = json.Unmarshal([]byte(val), &executionDetails); err != nil {
		    fmt.Errorf("Error decoding JSON response: %s: %s", err, response.Value)
		    return nil
	    }
	    acc.AddFields("response", executionDetails, mergeTags(tags, executionDetails)) 
    }

	return nil
}

// mergeTags combines two tag sets into a single tag set.
func mergeTags(metricTags map[string]string, outerTags map[string]interface{}) map[string]string {
	tags := make(map[string]string)
	for k, v := range outerTags {
		if v == nil {
			tags[k] = "null"
		} else {
			tags[k] = v.(string)
		}
	}
	for k, v := range metricTags {
		tags[k] = v
	}

	return tags
}

// makeExecRequests creates ExecRequest objects from metrics definitions.
func makeExecRequests(metrics []Metric) []ExecRequest {
	var requests []ExecRequest
	for _, metric := range metrics {
		execRequest := ExecRequest{
			Mbean:      metric.Mbean,
			Operation:	metric.Operation,
			Arguments:  []string{},
		}

		args := metric.Arguments
		if args != nil {
			for _, arg := range args {
				execRequest.Arguments = append(execRequest.Arguments, arg)
			}
		}
		requests = append(requests, execRequest)
	}

	return requests
}