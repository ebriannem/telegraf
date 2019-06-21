package tagcounter

import (
	"testing"
	"time"
  "github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

var m1, _ = metric.New("m1",
	map[string]string{
		 "a": "a_val1",
		 "b": "b_val1",
		 "c": "c_val1",
		 "d": "d_val1" },
	map[string]interface{}{
		"count": 1,
	},
	time.Now(),
)

var m2, _ = metric.New("m1",
	map[string]string{
		 "a": "a_val1",
		 "b": "b_val1",
		 "c": "c_val3",
		 "d": "d_val1" },
	map[string]interface{}{
		"count": 1,
	},
	time.Now(),
)

var m3, _ = metric.New("m1",
	map[string]string{
		 "a": "a_val2",
		 "b": "b_val1",
		 "c": "c_val1",
		 "d": "d_val3" },
	map[string]interface{}{
		"count": 1,
	},
	time.Now(),
)

var m4, _ = metric.New("m4",
	map[string]string{
		 "a": "a_val1",
		 "b": "b_val1",
		 "c": "c_val4",
		 "d": "d_val1" },
	map[string]interface{}{
		"count": 1,
	},
	time.Now(),
)

var m5, _ = metric.New("m1",
	map[string]string{
		 "a": "a_val1",
		 "b": "b_val1",
		 "c": "c_val4",
		 "d": "d_val1" },
	map[string]interface{}{
		"count": 1,
	},
	time.Now(),
)

func NewTestTagCounter(tagNames []string) telegraf.Aggregator {
	vc := &TagCounter{
		TagNames: tagNames,
	}
	vc.Reset()
	return vc
}

func BenchmarkApply(b *testing.B) {
	tc := NewTestTagCounter([]string{"a", "c"})

	for n := 0; n < b.N; n++ {
		tc.Add(m1)
		tc.Add(m3)
	}
}

// Test basic functionality
func TestBasic(t *testing.T) {
	tc := NewTestTagCounter([]string{"a", "b"})
	acc := testutil.Accumulator{}

	tc.Add(m1)
	tc.Add(m2)
	tc.Add(m1)
	tc.Push(&acc)

	expectedFields := map[string]interface{}{
		"count": 3,}

	expectedTags := map[string]string{
	 		 "a": "a_val1",
	 		 "b": "b_val1",}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)
}

// Test basic functionality
func TestBasic2(t *testing.T) {
	tc := NewTestTagCounter([]string{"a", "b", "d"})
	acc := testutil.Accumulator{}

	for i := 0; i < 100; i++ {
		tc.Add(m1)
		tc.Add(m2)
		tc.Add(m3)
		tc.Add(m5)
	}
	tc.Push(&acc)

	expectedFields1 := map[string]interface{}{
		"count": 300,}

	expectedTags1 := map[string]string{
	 		 "a": "a_val1",
	 		 "b": "b_val1",
		 	 "d": "d_val1"}

	expectedFields2 := map[string]interface{}{
 		"count": 100,}

 	expectedTags2 := map[string]string{
 	 		 "a": "a_val2",
 	 		 "b": "b_val1",
		 	 "d": "d_val3"}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields1, expectedTags1)
	acc.AssertContainsTaggedFields(t, "m1", expectedFields2, expectedTags2)
}

// Test with missing tags
func TestMissing(t *testing.T) {
	tc := NewTestTagCounter([]string{"a", "e"})
	acc := testutil.Accumulator{}

	tc.Add(m1)
	tc.Add(m2)
	tc.Add(m1)
	tc.Push(&acc)

	expectedFields := map[string]interface{}{
		"count": 3,}

	expectedTags := map[string]string{
	 		 "a": "a_val1",
	 		 "e": "",}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)
}

// Test with multiple tag combinations to count
func TestMultipleTagGroups(t *testing.T) {
	tc := NewTestTagCounter([]string{"a", "b"})
	acc := testutil.Accumulator{}

	tc.Add(m1)
	tc.Add(m2)
	tc.Add(m3)
	tc.Add(m3)
	tc.Add(m3)
	tc.Push(&acc)

	expectedFields1 := map[string]interface{}{
		"count":      2,
	}
	expectedTags1 := map[string]string{
		 "a": "a_val1",
		 "b": "b_val1"}

	 expectedFields2 := map[string]interface{}{
 		"count":      3,
 	}
 	expectedTags2 := map[string]string{
 		 "a": "a_val2",
 		 "b": "b_val1"}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields1, expectedTags1)
	acc.AssertContainsTaggedFields(t, "m1", expectedFields2, expectedTags2)
}

// Test with a reset between two runs
func TestWithReset(t *testing.T) {
	tc := NewTestTagCounter([]string{"a", "b"})
	acc := testutil.Accumulator{}

	tc.Add(m1)
	tc.Add(m2)
	tc.Push(&acc)

	expectedFields1 := map[string]interface{}{
		"count":      2,
	}
	expectedTags1 := map[string]string{
		 "a": "a_val1",
		 "b": "b_val1"}

	acc.AssertContainsTaggedFields(t, "m1", expectedFields1, expectedTags1)

	acc.ClearMetrics()
	tc.Reset()

	tc.Add(m3)
	tc.Add(m3)
	tc.Add(m3)
	tc.Add(m1)
	tc.Push(&acc)

	expectedFields2 := map[string]interface{}{
 		"count":      3,
 	}
 	expectedTags2 := map[string]string{
 		 "a": "a_val2",
 		 "b": "b_val1"}

	expectedFields3 := map[string]interface{}{
 		"count":      1,
 	}
 	expectedTags3 := map[string]string{
 		 "a": "a_val1",
 		 "b": "b_val1"}

	acc.AssertContainsTaggedFields(t, "m1", expectedFields2, expectedTags2)
	acc.AssertContainsTaggedFields(t, "m1", expectedFields3, expectedTags3)
}

// Test with multiple metrics
func TestMultipleMetrics(t *testing.T) {
	tc := NewTestTagCounter([]string{"a", "b"})
	acc := testutil.Accumulator{}

	tc.Add(m1)
	tc.Add(m2)
	tc.Add(m4)
	tc.Push(&acc)

	expectedFields1 := map[string]interface{}{
		"count":      2,
	}
	expectedTags1 := map[string]string{
		 "a": "a_val1",
		 "b": "b_val1"}

	 expectedFields2 := map[string]interface{}{
 		"count":      1,
 	}
 	expectedTags2 := map[string]string{
 		 "a": "a_val1",
 		 "b": "b_val1"}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields1, expectedTags1)
	acc.AssertContainsTaggedFields(t, "m4", expectedFields2, expectedTags2)
}
