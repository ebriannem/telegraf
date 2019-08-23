package maptagger

import (
	"strings"

	"github.com/go-redis/redis"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type MapTagger struct {
	client      *redis.Client
	Addr        string
	Password    string
	DB          int
	conversions []conversion
}

type conversion struct {
	OldKey   string
	NewKey   string
	Contains bool
	Default  string
	Mapping  map[string]string
}

const sampleConfig = `
# [[processors.maptagger.conversions]]
#	 ## The key of the value that will be searched for in the mapping
#	 old_key = "id_num"
#	 ## The key to be added
#	 new_key = "name"
#	 ## Whether the value must exactly match a key in mapping or just contain
#	 ## 	a key in the mapping
#	 # contains = false
#	 ## The default value when the key is missing
#	 # default = "Unknown"
#	 ## The mapping
#	 [processors.maptagger.conversions.mapping]
#	 "123" = "Name 1"
#  "345" = "Name 2"
# ## More than one conversion can be added.
# ## Order of conversions matters: a later conversion can access tags added by earlier conversions
`

func NewMapTagger() *MapTagger {
	return &MapTagger{}
}

func (mt *MapTagger) SampleConfig() string {
	return sampleConfig
}

func (mt *MapTagger) Description() string {
	return ""
}

func (mt *MapTagger) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		for _, tagConversion := range mt.conversions {
			//Set default value for if there is no match
			newValue := tagConversion.Default
			if len(newValue) == 0 {
				newValue = "Unknown"
			}
			//Look for a match only if there is an original value
			if oldValue, exists := metric.GetTag(tagConversion.OldKey); exists {
				if !tagConversion.Contains {
					//Get exact match
					if value, exists := tagConversion.Mapping[oldValue]; exists {
						newValue = value
					}
				} else {
					//Get inexact match
					for mapKey, mapValue := range tagConversion.Mapping {
						if strings.Contains(oldValue, mapKey) {
							newValue = mapValue
							break
						}
					}
				}
			}
			metric.AddTag(tagConversion.NewKey, newValue)
		}
	}
	return in
}

func init() {
	processors.Add("maptagger", func() telegraf.Processor {
		return NewMapTagger()
	})
}
