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

type TagTest struct {
	name 	string
	metricTags map[string]string
	outerTags  map[string]interface{}
	expected  map[string]string
}

func TestJolokia2_mergeTags(t *testing.T) {

	cases := [] TagTest {
	{
		name: "case1",
		metricTags: map[string]string{
			"metricTag1": "val1",
			"metricTag2": "val2",
		}, 
		outerTags: map[string]interface{}{
	    	"outerTag1":"val1",
			"outerTag2":"val2",
			"outerTag3":"val3",
	    },
	    expected: map[string]string{
	    	"metricTag1": "val1",
	    	"metricTag2": "val2",
	    	"outerTag1":"val1",
			"outerTag2":"val2",
			"outerTag3":"val3",
	    },
	},
	{
		name: "case2_empty_metrictags",
		metricTags: map[string]string{}, 
		outerTags: map[string]interface{}{
	    	"outerTag1":"val1",
			"outerTag2":"val2",
			"outerTag3":"val3",
	    },
	    expected: map[string]string{
	    	"outerTag1":"val1",
			"outerTag2":"val2",
			"outerTag3":"val3",
	    },
	},
	{
		name: "case3_empty_outertags",
		metricTags: map[string]string{
			"metricTag1": "val1",
			"metricTag2": "val2",
			"metricTag3": "val3",
		}, 
		outerTags: map[string]interface{}{},
	    expected: map[string]string{
	    	"metricTag1": "val1",
			"metricTag2": "val2",
			"metricTag3": "val3",
	    },
	},
	{
		name: "case4_empty",
		metricTags: map[string]string{}, 
		outerTags: map[string]interface{}{},
	    expected: map[string]string{},
	},
}

	for _, c := range cases {
		result := mergeTags(c.metricTags, c.outerTags)
		assert.Equal(t, result, c.expected, "Failing case: " + c.name)
	}
}