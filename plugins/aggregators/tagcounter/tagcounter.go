package tagcounter

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

// NumTags is the maximum number of tags that can be aggregated on.
const NumTags int = 4

//TagCounter stores the aggregated metrics and the relevant tags
type TagCounter struct {
	cache  map[string]map[tagCombo]int
	TagNames []string `toml:"tag_names"`
}

// tagCombo stores the data for an aggregated metric
type tagCombo struct {
	metricName string
	tagValues [NumTags]string
}

var sampleConfig = `
  ## The frequency at which the plugin sends aggregated metrics.
  # period = "30s"
	## Whether the old metric should still be stored.
  # drop_original = true
	## The tag keys to aggregate on. Note: NumTags must be set to be >= the length of this array.
	# tag_names = ["fdpErrorCode", "tier", "flow_category", "intuit_fdp_flowname"]
`

//SampleConfig ...
func (tc *TagCounter) SampleConfig() string {
	return sampleConfig
}

//Description ...
func (tc *TagCounter) Description() string {
	return "Aggregates by the count of specified tag combinations."
}

func init() {
	aggregators.Add("tagcounter", func() telegraf.Aggregator {
		return NewTagCounter()
	})
}

// NewTagCounter creates a new TagCounter
func NewTagCounter() telegraf.Aggregator {
	tc := &TagCounter{}
	if NumTags < len(tc.TagNames) {
		//error
	}
	tc.Reset()
	return tc
}

// Add takes in a new metric and aggegates it into the current cache
func (tc *TagCounter) Add(in telegraf.Metric) {
	hid := in.Name()
	if _, ok := tc.cache[hid]; !ok {
		tc.cache[hid] = make(map[tagCombo]int)
	}

	newtagCombo := tagCombo {
		metricName: in.Name()}
	tc.assignTags(&newtagCombo, in.Tags())

	if _, ok := tc.cache[hid][newtagCombo]; ok {
		tc.cache[hid][newtagCombo]++
	} else {
		tc.cache[hid][newtagCombo] = 1
	}
}

func (tc *TagCounter) assignTags(tagCombo *tagCombo, tags map[string]string) {
	for i, tagName := range tc.TagNames {
		if tagVal, ok := tags[tagName]; ok {
			tagCombo.tagValues[i] = tagVal
		}
	}
}

// Push sends aggregated metrics
func (tc *TagCounter) Push(acc telegraf.Accumulator) {
	for _, tagMap := range tc.cache {
		for agg, count := range tagMap {

			fields := map[string]interface{} {
				"count" : count,
			}

			newTags := make(map[string]string)
			for i, tagName := range tc.TagNames {
				newTags[tagName] = agg.tagValues[i]
			}

			acc.AddFields(agg.metricName, fields, newTags)
		}
	}
}

// Reset resets the cache of the TagCounter
func (tc *TagCounter) Reset() {
	tc.cache = make(map[string]map[tagCombo]int)
}
