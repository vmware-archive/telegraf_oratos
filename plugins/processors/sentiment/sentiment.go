package parser

import (
	"fmt"

	"github.com/cdipaolo/sentiment"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

var SampleConfig = `
  ## The name of the fields whose value will be analyzed.
  analyze_fields = []
`

type Sentiment struct {
	AnalyzeFields []string `toml:"analyze_fields"`
	Model         sentiment.Models
}

func (s *Sentiment) SampleConfig() string {
	return SampleConfig
}

func (s *Sentiment) Description() string {
	return "Run a sentiment analysis algorithm on string metrics and return the results"
}

func (s *Sentiment) Apply(metrics ...telegraf.Metric) []telegraf.Metric {

	for _, metric := range metrics {
		if len(metric.Fields()) == 0 {
			continue
		}

		for _, field := range metric.FieldList() {
			if contains(s.AnalyzeFields, field.Key) {
				switch value := field.Value.(type) {
				case string:
					analysis := s.Model.SentimentAnalysis(value, sentiment.English)
					metric.AddField("sentiment_"+field.Key, int(analysis.Score))
				}

			}
		}
	}

	return metrics
}

func init() {
	model, err := sentiment.Restore()
	if err != nil {
		panic(fmt.Sprintf("Could not restore model!\n\t%v\n", err))
	}

	processors.Add("sentiment", func() telegraf.Processor {
		return &Sentiment{AnalyzeFields: []string{}, Model: model}
	})
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func average(nums []int) float32 {
	sum := 0
	for _, n := range nums {
		sum = sum + n
	}

	return float32(sum) / float32(len(nums))
}
