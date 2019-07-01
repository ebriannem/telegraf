package tagtrimmer

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func newMetric(name string, tags map[string]string, fields map[string]interface{}) telegraf.Metric {
	if tags == nil {
		tags = map[string]string{}
	}
	if fields == nil {
		fields = map[string]interface{}{}
	}
	m, _ := metric.New(name, tags, fields, time.Now())
	return m
}

func TestBasic(t *testing.T) {
	r := TagTrimmer {
		TagName: "name1",
		TrimValues: []string{"a", "b"},
	}
	m1 := newMetric("foo1", map[string]string{"name1": "1a"}, nil)
	m2 := newMetric("foo2", map[string]string{"name1": "b2"}, nil)
	m3 := newMetric("foo3", map[string]string{"name1": "3a3"}, nil)

	results := r.Apply(m1, m2, m3)

	m12 := newMetric("foo1", map[string]string{"name1": "a"}, nil)
	m22 := newMetric("foo2", map[string]string{"name1": "b"}, nil)
	m32 := newMetric("foo3", map[string]string{"name1": "a"}, nil)

	assert.Equal(t, m12.Tags(), results[0].Tags())
	assert.Equal(t, m22.Tags(), results[1].Tags())
	assert.Equal(t, m32.Tags(), results[2].Tags())
}

func TestBasic2(t *testing.T) {
	r := TagTrimmer {
		TagName: "name1",
		TrimValues: []string{"abcbc"},
	}
	m1 := newMetric("foo1", map[string]string{"name1": "abcbc"}, nil)
	m2 := newMetric("foo2", map[string]string{"name1": "abcdc"}, nil)
	m3 := newMetric("foo3", map[string]string{"name1": "123abcbc456"}, nil)

	results := r.Apply(m1, m2, m3)

	m12 := newMetric("foo1", map[string]string{"name1": "abcbc"}, nil)
	m22 := newMetric("foo2", map[string]string{"name1": "abcdc"}, nil)
	m32 := newMetric("foo3", map[string]string{"name1": "abcbc"}, nil)

	assert.Equal(t, m12.Tags(), results[0].Tags())
	assert.Equal(t, m22.Tags(), results[1].Tags())
	assert.Equal(t, m32.Tags(), results[2].Tags())
}

func TestIgnoresOtherTags(t *testing.T) {
	r := TagTrimmer {
		TagName: "name1",
		TrimValues: []string{"a", "b"},
	}
	m1 := newMetric("foo1", map[string]string{"name3": "1a"}, nil)
	m2 := newMetric("foo2", map[string]string{"name2": "b2"}, nil)
	m3 := newMetric("foo3", map[string]string{"name1": "3c3"}, nil)

	results := r.Apply(m1, m2, m3)

	m12 := newMetric("foo1", map[string]string{"name3": "1a"}, nil)
	m22 := newMetric("foo2", map[string]string{"name2": "b2"}, nil)
	m32 := newMetric("foo3", map[string]string{"name1": "3c3"}, nil)

	assert.Equal(t, m12.Tags(), results[0].Tags())
	assert.Equal(t, m22.Tags(), results[1].Tags())
	assert.Equal(t, m32.Tags(), results[2].Tags())
}
