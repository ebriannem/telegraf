package maptagger

import (
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type MapTagger struct {
	Tags []conversion
}

type conversion struct {
	OldKey  string
	NewKey  string
	Exact   bool
	Mapping map[string]string
}

const sampleConfig = `
	# [[processors.maptagger]]
	# [[processors.maptagger.tags]]
	#		## The key of the value that determines the new key's value
	#		old_key = "id_num"
	#		## The key to be added
	#		new_key = "name"
	#		## Whether the value must exactly match a key in mapping or just contain
	#		## 	a key in the mapping to be assigned the corresponding value
	#		exact = true
	#		## The mapping from the old key's values to the new key's values
	#		[processors.maptagger.tags.mapping]
	#		"0123" = "name1"
	#		"1234" = "name2"
	#
	## More than one conversion can be added.
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
		for _, tagConversion := range mt.Tags {
			newValue := "Unknown"
			if oldValue, ok := metric.GetTag(tagConversion.OldKey); ok {
				if tagConversion.Exact {
					if mapValue, ok := tagConversion.Mapping[oldValue]; ok {
						newValue = mapValue
					}
				} else {
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
