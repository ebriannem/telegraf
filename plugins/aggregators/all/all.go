package all

import (
	_ "github.com/influxdata/telegraf/plugins/aggregators/basicstats"
	_ "github.com/influxdata/telegraf/plugins/aggregators/final"
	_ "github.com/influxdata/telegraf/plugins/aggregators/histogram"
	_ "github.com/influxdata/telegraf/plugins/aggregators/minmax"
	_ "github.com/influxdata/telegraf/plugins/aggregators/valuecounter"
	_ "github.com/influxdata/telegraf/plugins/aggregators/tagcounter"
)
