package redismapper

import (
	"testing"
	"time"

	"github.com/go-redis/redis"
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

func TestHandleQuotes(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	//Everything outside surrounding quotes will be ignored
	client.HMSet("TestHandleQuotes", map[string]interface{}{"\"a\"": "\x00\x00\x06\"b\""})
	//Keys without quotes around them will be ignored (a value of "c" maps to a key of "\"c\"", not "c")
	//Values without quotes will be unchanged
	client.HMSet("TestHandleQuotes2", map[string]interface{}{"\"a\"": "b", "c": "d"})
	//Interior quotes should not be removed ("\x00\x00\x06\"\"b\"c\"d\"" maps to "b\"c\"d")
	client.HMSet("TestHandleQuotes3", map[string]interface{}{"\"a\"": "\x00\x00\x06\"\"b\"c\"d\""})

	rm := NewRedisMapper()
	rm.client = client
	rm.Conversions = []conversion{
		conversion{
			HandleQuotes: true,
			OldKey:       "name1",
			NewKey:       "name2",
			CacheName:    "TestHandleQuotes",
		},
		conversion{
			HandleQuotes: true,
			OldKey:       "name3",
			NewKey:       "name4",
			CacheName:    "TestHandleQuotes2",
		},
		conversion{
			HandleQuotes: true,
			OldKey:       "name1",
			NewKey:       "name5",
			CacheName:    "TestHandleQuotes3",
		},
	}
	m1 := newMetric("foo1", map[string]string{"name1": "a"}, nil)
	m2 := newMetric("foo2", map[string]string{"name1": "a", "name3": "c"}, nil)
	results := rm.Apply(m1, m2)
	m12 := newMetric("foo1", map[string]string{"name1": "a", "name2": "b",
		"name4": "Unknown", "name5": "\"b\"c\"d"}, nil)
	m22 := newMetric("foo2", map[string]string{"name1": "a", "name3": "c",
		"name2": "b", "name4": "Unknown", "name5": "\"b\"c\"d"}, nil)
	assert.Equal(t, m12.Tags(), results[0].Tags())
	assert.Equal(t, m22.Tags(), results[1].Tags())
}

func TestMultiple(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	client.HMSet("TestMap1", map[string]interface{}{"a": "b", "c": "d"})
	client.HMSet("TestMap2", map[string]interface{}{"a": "b", "e": "f"})
	client.HMSet("TestMap3", map[string]interface{}{"1": "2", "3": "4"})

	rm := NewRedisMapper()
	rm.client = client
	rm.Conversions = []conversion{
		conversion{
			OldKey:    "name1a",
			NewKey:    "name1b",
			CacheName: "TestMap1",
		},
		conversion{
			OldKey:    "name2a",
			NewKey:    "name2b",
			CacheName: "TestMap2",
		},
		conversion{
			OldKey:    "name3a",
			NewKey:    "name3b",
			CacheName: "TestMap3",
		},
	}
	m1 := newMetric("foo1", map[string]string{"name1a": "a", "name2a": "a"}, nil)
	m2 := newMetric("foo2", map[string]string{"name2a": "b"}, nil)
	m3 := newMetric("foo3", map[string]string{"name1a": "c", "name2a": "e", "name3a": "3"}, nil)

	results := rm.Apply(m1, m2, m3)

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
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	client.HMSet("TestMap4", map[string]interface{}{"a": "b"})

	rm := NewRedisMapper()
	rm.client = client
	rm.Conversions = []conversion{
		conversion{
			OldKey:    "name1",
			NewKey:    "name2",
			CacheName: "TestMap4",
			Contains:  true,
		},
	}
	m1 := newMetric("foo1", map[string]string{"name1": "abc"}, nil)
	m2 := newMetric("foo2", map[string]string{"name1": "a"}, nil)
	m3 := newMetric("foo3", map[string]string{"name1": "bc"}, nil)

	results := rm.Apply(m1, m2, m3)

	m12 := newMetric("foo1", map[string]string{"name1": "abc", "name2": "b"}, nil)
	m22 := newMetric("foo2", map[string]string{"name1": "a", "name2": "b"}, nil)
	m32 := newMetric("foo3", map[string]string{"name1": "bc", "name2": "Unknown"}, nil)

	assert.Equal(t, m12.Tags(), results[0].Tags())
	assert.Equal(t, m22.Tags(), results[1].Tags())
	assert.Equal(t, m32.Tags(), results[2].Tags())
}

func TestOrder(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	client.HMSet("TestOrderMap1", map[string]interface{}{"1": "a"})
	client.HMSet("TestOrderMap2", map[string]interface{}{"a": "b"})
	client.HMSet("TestOrderMap3", map[string]interface{}{"b": "c"})

	rm := NewRedisMapper()
	rm.client = client
	rm.Conversions = []conversion{
		conversion{
			OldKey:    "1",
			NewKey:    "2",
			CacheName: "TestOrderMap1",
			Contains:  true,
		},
		//This conversion will always give unknown for m1 because the key "3" is
		// added *after* this conversion happens
		conversion{
			OldKey:    "3",
			NewKey:    "4",
			CacheName: "TestOrderMap3",
			Contains:  true,
		},
		conversion{
			OldKey:    "2",
			NewKey:    "3",
			CacheName: "TestOrderMap2",
			Contains:  true,
		},
	}
	m1 := newMetric("foo1", map[string]string{"1": "1"}, nil)

	results := rm.Apply(m1)
	m12 := newMetric("foo1", map[string]string{"1": "1", "2": "a", "3": "b", "4": "Unknown"}, nil)

	assert.Equal(t, m12.Tags(), results[0].Tags())
}

func TestDefault(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	client.HMSet("TestMapDefault", map[string]interface{}{})

	rm := NewRedisMapper()
	rm.client = client
	rm.Conversions = []conversion{
		conversion{
			OldKey:    "nameA1",
			NewKey:    "nameA2",
			CacheName: "TestMapDefault",
		},
		conversion{
			OldKey:    "nameB1",
			NewKey:    "nameB2",
			CacheName: "TestMapDefault",
			Default:   "customDefault",
		},
	}
	m1 := newMetric("foo1", map[string]string{"nameA1": "a", "nameB1": "c"}, nil)

	results := rm.Apply(m1)

	m1b := newMetric("foo1", map[string]string{"nameA1": "a", "nameB1": "c",
		"nameA2": "Unknown", "nameB2": "customDefault"}, nil)

	assert.Equal(t, m1b.Tags(), results[0].Tags())
}
