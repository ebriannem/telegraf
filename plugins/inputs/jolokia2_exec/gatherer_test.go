package jolokia2_exec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJolokia2_makeExecRequests(t *testing.T) {
	cases := []struct {
		metric   Metric
		expected []ExecRequest
	}{
		{
			metric: Metric{
				Name:  "object",
				Mbean: "test:foo=bar",
				Operation:	"poll",
			},
			expected: []ExecRequest{
				{
					Mbean:      "test:foo=bar",
					Operation:	"poll",
					Arguments: []string{},
				},
			},
		}, {
			metric: Metric{
				Name:  "object_with_an_argument",
				Mbean: "test:foo=bar",
				Operation:	"poll",
				Arguments: []string{"biz"},
			},
			expected: []ExecRequest{
				{
					Mbean:      "test:foo=bar",
					Operation:	"poll",
				    Arguments: []string{"biz"},
				},
			},
		}, {
			metric: Metric{
				Name:  "object_with_arguments",
				Mbean: "test:foo=bar",
				Operation:	"poll",
				Arguments: []string{"baz", "biz"},
			},
			expected: []ExecRequest{
				{
					Mbean:      "test:foo=bar",
					Operation:	"poll",
				    Arguments: []string{"baz", "biz"},
				},
			},
		}, 
	}

	for _, c := range cases {
		payload := makeExecRequests([]Metric{c.metric})

		assert.Equal(t, len(c.expected), len(payload), "Failing case: "+c.metric.Name)
		for _, actual := range payload {
			assert.Contains(t, c.expected, actual, "Failing case: "+c.metric.Name)
		}
	}
}