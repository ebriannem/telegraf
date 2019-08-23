package redismapper

import (
	"strings"

	"github.com/go-redis/redis"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type RedisMapper struct {
	client      *redis.Client
	Addr        string
	Password    string
	DB          int
	Conversions []conversion
}

type conversion struct {
	OldKey       string
	NewKey       string
	Contains     bool
	HandleQuotes bool
	CacheName    string
	Default      string
}

const sampleConfig = `
	#		addr = "localhost:6379"
	#		password = ""
	#		DB = 0
	# 	[[processors.redismapper.conversions]]
	#			## The key of the value that determines the new key's value
	#			old_key = "id_num"
	#			## The key to be added
	#			new_key = "name"
	#			## Whether the value must exactly match a key in mapping or just contain
	#			## 	a key in the mapping to be assigned the corresponding value
	#			# contains = false
	#			## If true, trimms values of escape characters and quotes and adds quotes
	#			## to keys
	#			# handle_quotes = false
	#			## The redis hash key with the mapping from the old key's values to the new key's values
	#			cache_name = "NameMap"
	# 	## More than one conversion can be added.
	# 	## Order of conversions matters: a later conversion can access tags added by earlier conversions
`

func NewRedisMapper() *RedisMapper {
	rm := &RedisMapper{}
	rm.client = redis.NewClient(&redis.Options{
		Addr:     rm.Addr,
		Password: rm.Password,
		DB:       rm.DB,
	})
	return rm
}

func (rm *RedisMapper) SampleConfig() string {
	return sampleConfig
}

func (rm *RedisMapper) Description() string {
	return ""
}

func (rm *RedisMapper) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		for _, tagConversion := range rm.Conversions {
			newValue := tagConversion.Default
			if len(newValue) == 0 {
				newValue = "Unknown"
			}
			if oldValue, ok := metric.GetTag(tagConversion.OldKey); ok {
				if tagConversion.HandleQuotes {
					oldValue = addQuotes(oldValue)
				}
				if !tagConversion.Contains {
					if value, error := rm.client.HGet(tagConversion.CacheName, oldValue).Result(); error == nil {
						newValue = value
					}
				} else {
					if mappings, error := rm.client.HGetAll(tagConversion.CacheName).Result(); error == nil {
						for mapKey, mapValue := range mappings {
							if strings.Contains(oldValue, mapKey) {
								newValue = mapValue
								break
							}
						}
					}
				}
			}
			if tagConversion.HandleQuotes {
				newValue = removeQuotes(newValue)
			}
			metric.AddTag(tagConversion.NewKey, newValue)
		}
	}
	return in
}
func removeQuotes(s string) string {
	start := strings.Index(s, "\"") + 1
	end := strings.LastIndex(s, "\"")
	if start < end {
		return s[start:end]
	}
	return s
}
func addQuotes(s string) string {
	return "\"" + s + "\""
}

func init() {
	processors.Add("redismapper", func() telegraf.Processor {
		return NewRedisMapper()
	})
}
