package tagtrimmer

import (
	"strings"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
 # tag_name = "intuit_offeringid"
 # trim_values = ["mint", "qbo", "qbse", "qbf", "turbotax", "ACCOUNT_VIEW",
  "triage", "fdptools", "6", "tax.ctg", "sbe.salsa.platform",
	 "smallbusiness.mmo.ui", "sbg.payments.account"]
`

// TagTrimmer stores the tag name and the values to be trimmed to
type TagTrimmer struct {
	TagName string `toml:"tag_name"`
	TrimValues []string `toml:"trim_values"`
}

// SampleConfig ...
func (tt *TagTrimmer) SampleConfig() string {
	return sampleConfig
}

// Description ...
func (tt *TagTrimmer) Description() string {
	return "Trims tag values containing a given string down to that string."
}


// Apply ...
func (tt *TagTrimmer) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		if val, ok := metric.GetTag(tt.TagName); ok {
			for _, trimValue := range tt.TrimValues {
				if strings.Contains(val, trimValue) {
					metric.RemoveTag(tt.TagName)
					metric.AddTag(tt.TagName, trimValue)
					break
				}
			}
		}
	}
	return in
}

func init() {
	processors.Add("tagtrimmer", func() telegraf.Processor {
		return &TagTrimmer{}
	})
}
