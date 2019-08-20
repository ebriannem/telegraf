package maptagger

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

func TestExact(t *testing.T) {
	mt := NewMapTagger()
	mt.Tags = []conversion{
		conversion{
			OldKey:  "name1",
			NewKey:  "name2",
			Mapping: map[string]string{"a": "b", "c": "d"},
			Exact:   true,
		},
	}
	m1 := newMetric("foo1", map[string]string{"name1": "a"}, nil)
	m2 := newMetric("foo2", map[string]string{"name1": "abc"}, nil)
	m3 := newMetric("foo3", map[string]string{"name1": "c"}, nil)

	results := mt.Apply(m1, m2, m3)

	m12 := newMetric("foo1", map[string]string{"name1": "a", "name2": "b"}, nil)
	m22 := newMetric("foo2", map[string]string{"name1": "abc", "name2": "Unknown"}, nil)
	m32 := newMetric("foo3", map[string]string{"name1": "c", "name2": "d"}, nil)

	assert.Equal(t, m12.Tags(), results[0].Tags())
	assert.Equal(t, m22.Tags(), results[1].Tags())
	assert.Equal(t, m32.Tags(), results[2].Tags())
}

func TestMultiple(t *testing.T) {
	mt := NewMapTagger()
	mt.Tags = []conversion{
		conversion{
			OldKey:  "name1a",
			NewKey:  "name1b",
			Mapping: map[string]string{"a": "b", "c": "d"},
			Exact:   true,
		},
		conversion{
			OldKey:  "name2a",
			NewKey:  "name2b",
			Mapping: map[string]string{"a": "b", "e": "f"},
			Exact:   true,
		},
		conversion{
			OldKey:  "name3a",
			NewKey:  "name3b",
			Mapping: map[string]string{"1": "2", "3": "4"},
			Exact:   true,
		},
	}
	m1 := newMetric("foo1", map[string]string{"name1a": "a", "name2a": "a"}, nil)
	m2 := newMetric("foo2", map[string]string{"name2a": "b"}, nil)
	m3 := newMetric("foo3", map[string]string{"name1a": "c", "name2a": "e", "name3a": "3"}, nil)

	results := mt.Apply(m1, m2, m3)

	m1b := newMetric("foo1", map[string]string{"name1a": "a", "name2a": "a",
		"name1b": "b", "name2b": "b", "name3b": "Unknown"}, nil)
	m2b := newMetric("foo2", map[string]string{"name2a": "b", "name1b": "Unknown",
		"name2b": "Unknown", "name3b": "Unknown"}, nil)
	m3b := newMetric("foo3", map[string]string{"name1a": "c", "name2a": "e",
		"name3a": "3", "name1b": "d", "name2b": "f", "name3b": "4"}, nil)

	assert.Equal(t, m1b.Tags(), results[0].Tags())
	assert.Equal(t, m2b.Tags(), results[1].Tags())
	assert.Equal(t, m3b.Tags(), results[2].Tags())
}

func TestInexact(t *testing.T) {
	mt := NewMapTagger()
	mt.Tags = []conversion{
		conversion{
			OldKey:  "name1",
			NewKey:  "name2",
			Mapping: map[string]string{"a": "b"},
			Exact:   false,
		},
	}
	m1 := newMetric("foo1", map[string]string{"name1": "abc"}, nil)
	m2 := newMetric("foo2", map[string]string{"name1": "a"}, nil)
	m3 := newMetric("foo3", map[string]string{"name1": "bc"}, nil)

	results := mt.Apply(m1, m2, m3)

	m12 := newMetric("foo1", map[string]string{"name1": "abc", "name2": "b"}, nil)
	m22 := newMetric("foo2", map[string]string{"name1": "a", "name2": "b"}, nil)
	m32 := newMetric("foo3", map[string]string{"name1": "bc", "name2": "Unknown"}, nil)

	assert.Equal(t, m12.Tags(), results[0].Tags())
	assert.Equal(t, m22.Tags(), results[1].Tags())
	assert.Equal(t, m32.Tags(), results[2].Tags())
}
