package parser

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//compares metrics without comparing time
func compareMetrics(t *testing.T, expected, actual []telegraf.Metric) {
	assert.Equal(t, len(expected), len(actual))
	for i, metric := range actual {
		require.Equal(t, expected[i].Name(), metric.Name())
		require.Equal(t, expected[i].Fields(), metric.Fields())
		require.Equal(t, expected[i].Tags(), metric.Tags())
	}
}

func Metric(v telegraf.Metric, err error) telegraf.Metric {
	if err != nil {
		panic(err)
	}
	return v
}

func TestApply(t *testing.T) {
	tests := []struct {
		name          string
		analyzeFields []string
		input         telegraf.Metric
		expected      []telegraf.Metric
	}{
		{
			name:          "test sentiment of one sentence",
			analyzeFields: []string{"header", "body"},
			input: Metric(
				metric.New(
					"MyMetric",
					map[string]string{},
					map[string]interface{}{
						"header": "This thing sucks",
						"body":   "Wow thats great",
						"ignore": "This is the best",
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"MyMetric",
					map[string]string{},
					map[string]interface{}{
						"ignore":           "This is the best",
						"header":           "This thing sucks",
						"body":             "Wow thats great",
						"sentiment_header": 0,
						"sentiment_body":   1,
					},
					time.Unix(0, 0))),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sentiment := Sentiment{
				AnalyzeFields: tt.analyzeFields,
			}

			output := sentiment.Apply(tt.input)
			t.Logf("Testing: %s", tt.name)
			compareMetrics(t, tt.expected, output)
		})
	}
}

// func TestBadApply(t *testing.T) {
// 	tests := []struct {
// 		name        string
// 		parseFields []string
// 		config      parsers.Config
// 		input       telegraf.Metric
// 		expected    []telegraf.Metric
// 	}{
// 		{
// 			name:        "field not found",
// 			parseFields: []string{"bad_field"},
// 			config: parsers.Config{
// 				DataFormat: "json",
// 			},
// 			input: Metric(
// 				metric.New(
// 					"bad",
// 					map[string]string{},
// 					map[string]interface{}{
// 						"some_field": 5,
// 					},
// 					time.Unix(0, 0))),
// 			expected: []telegraf.Metric{
// 				Metric(metric.New(
// 					"bad",
// 					map[string]string{},
// 					map[string]interface{}{
// 						"some_field": 5,
// 					},
// 					time.Unix(0, 0))),
// 			},
// 		},
// 		{
// 			name:        "non string field",
// 			parseFields: []string{"some_field"},
// 			config: parsers.Config{
// 				DataFormat: "json",
// 			},
// 			input: Metric(
// 				metric.New(
// 					"bad",
// 					map[string]string{},
// 					map[string]interface{}{
// 						"some_field": 5,
// 					},
// 					time.Unix(0, 0))),
// 			expected: []telegraf.Metric{
// 				Metric(metric.New(
// 					"bad",
// 					map[string]string{},
// 					map[string]interface{}{
// 						"some_field": 5,
// 					},
// 					time.Unix(0, 0))),
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			parser := Parser{
// 				Config:      tt.config,
// 				ParseFields: tt.parseFields,
// 			}

// 			output := parser.Apply(tt.input)

// 			compareMetrics(t, output, tt.expected)
// 		})
// 	}
// }

// // Benchmarks

// func getMetricFields(metric telegraf.Metric) interface{} {
// 	key := "field3"
// 	if value, ok := metric.Fields()[key]; ok {
// 		return value
// 	}
// 	return nil
// }

// func getMetricFieldList(metric telegraf.Metric) interface{} {
// 	key := "field3"
// 	fields := metric.FieldList()
// 	for _, field := range fields {
// 		if field.Key == key {
// 			return field.Value
// 		}
// 	}
// 	return nil
// }

// func BenchmarkFieldListing(b *testing.B) {
// 	metric := Metric(metric.New(
// 		"test",
// 		map[string]string{
// 			"some": "tag",
// 		},
// 		map[string]interface{}{
// 			"field0": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
// 			"field1": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
// 			"field2": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
// 			"field3": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
// 			"field4": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
// 			"field5": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
// 			"field6": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
// 		},
// 		time.Unix(0, 0)))

// 	for n := 0; n < b.N; n++ {
// 		getMetricFieldList(metric)
// 	}
// }

// func BenchmarkFields(b *testing.B) {
// 	metric := Metric(metric.New(
// 		"test",
// 		map[string]string{
// 			"some": "tag",
// 		},
// 		map[string]interface{}{
// 			"field0": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
// 			"field1": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
// 			"field2": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
// 			"field3": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
// 			"field4": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
// 			"field5": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
// 			"field6": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
// 		},
// 		time.Unix(0, 0)))

// 	for n := 0; n < b.N; n++ {
// 		getMetricFields(metric)
// 	}
// }
