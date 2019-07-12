package tagreplacer

import (
	"strings"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
 # tag_name = "intuit_offeringid"
 # trim_values = [
  ["mint"], ["qbo", "sbe", "smallbusiness", "sbg"], ["qbse"],
	["qbf", "financ"], ["turbotax", "tax"], ["fdptools", "fdp", "agg"],
	["ACCOUNT_VIEW"]
	]
`

// TagReplacer stores the tag name and the values to be replaced to
type TagReplacer struct {
	TagName string `toml:"tag_name"`
	TrimValues [][]string `toml:"trim_values"`
}

// SampleConfig ...
func (tr *TagReplacer) SampleConfig() string {
	return sampleConfig
}

// Description ...
func (tr *TagReplacer) Description() string {
	return "Replaces tag values containing given strings down to a specified string."
}


// Apply ...
func (tr *TagReplacer) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		if val, ok := metric.GetTag(tr.TagName); ok {
			for _, trimValueSet := range tr.TrimValues {
				for _, trimValue := range trimValueSet {
					if strings.Contains(val, trimValue) {
						metric.RemoveTag(tr.TagName)
						metric.AddTag(tr.TagName, trimValueSet[0])
						break
					}
				}
			}
		}
	}
	return in
}

func init() {
	processors.Add("tagreplacer", func() telegraf.Processor {
		return &TagReplacer{}
	})
}
