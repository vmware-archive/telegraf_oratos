package cloudfoundry

import (
	"context"
	"sync"
	"testing"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func TestCloudFoundryOutput(t *testing.T) {
	t.Run("receives gauges", func(t *testing.T) {
		lc := newSpyLogCacheClient()
		cfOutput := &CloudFoundry{
			LogCacheClient: lc,
			SourceIDTag:    "source",
			InstanceIDTag:  "instance",
		}

		m, _ := metric.New(
			"test-metric",
			map[string]string{
				"source":   "test-source",
				"instance": "test-instance",
			},
			map[string]interface{}{
				"gauge": 42.0,
			},
			time.Now(),
		)

		err := cfOutput.Write([]telegraf.Metric{m})
		require.NoError(t, err)

		require.Contains(t, lc.envelopes(), &loggregator_v2.Envelope{
			Timestamp:  m.Time().UnixNano(),
			SourceId:   "test-source",
			InstanceId: "test-instance",
			Tags:       m.Tags(),
			Message: &loggregator_v2.Envelope_Gauge{
				Gauge: &loggregator_v2.Gauge{
					Metrics: map[string]*loggregator_v2.GaugeValue{
						"test-metric": {
							Unit:  "",
							Value: 42.0,
						},
					},
				},
			},
		})
	})

	t.Run("receives counters", func(t *testing.T) {
		lc := newSpyLogCacheClient()
		cfOutput := &CloudFoundry{
			LogCacheClient: lc,
			SourceIDTag:    "source",
			InstanceIDTag:  "instance",
		}

		m, _ := metric.New(
			"test-metric",
			map[string]string{
				"source":   "test-source",
				"instance": "test-instance",
			},
			map[string]interface{}{
				"counter": uint64(43),
			},
			time.Now(),
		)

		err := cfOutput.Write([]telegraf.Metric{m})
		require.NoError(t, err)

		require.Contains(t, lc.envelopes(), &loggregator_v2.Envelope{
			Timestamp:  m.Time().UnixNano(),
			SourceId:   "test-source",
			InstanceId: "test-instance",
			Tags:       m.Tags(),
			Message: &loggregator_v2.Envelope_Counter{
				Counter: &loggregator_v2.Counter{
					Name:  "test-metric",
					Total: uint64(43),
				},
			},
		})
	})

	t.Run("gets source_id from tag", func(t *testing.T) {
		lc := newSpyLogCacheClient()
		cfOutput := &CloudFoundry{
			LogCacheClient: lc,
			SourceIDTag:    "source_tag_test",
			InstanceIDTag:  "instance",
		}

		m, _ := metric.New(
			"test-metric",
			map[string]string{
				"source":          "test-source",
				"source_tag_test": "test-source-other-tag",
				"instance":        "test-instance",
			},
			map[string]interface{}{
				"counter": uint64(43),
			},
			time.Now(),
		)

		err := cfOutput.Write([]telegraf.Metric{m})
		require.NoError(t, err)

		require.Contains(t, lc.envelopes(), &loggregator_v2.Envelope{
			Timestamp:  m.Time().UnixNano(),
			SourceId:   "test-source-other-tag",
			InstanceId: "test-instance",
			Tags:       m.Tags(),
			Message: &loggregator_v2.Envelope_Counter{
				Counter: &loggregator_v2.Counter{
					Name:  "test-metric",
					Total: uint64(43),
				},
			},
		})
	})

	t.Run("gets instance_id from tag", func(t *testing.T) {
		lc := newSpyLogCacheClient()
		cfOutput := &CloudFoundry{
			LogCacheClient: lc,
			SourceIDTag:    "source",
			InstanceIDTag:  "instance_tag_test",
		}

		m, _ := metric.New(
			"test-metric",
			map[string]string{
				"source":            "test-source",
				"instance":          "test-instance",
				"instance_tag_test": "test-instance-tag",
			},
			map[string]interface{}{
				"counter": uint64(43),
			},
			time.Now(),
		)

		err := cfOutput.Write([]telegraf.Metric{m})
		require.NoError(t, err)

		require.Contains(t, lc.envelopes(), &loggregator_v2.Envelope{
			Timestamp:  m.Time().UnixNano(),
			SourceId:   "test-source",
			InstanceId: "test-instance-tag",
			Tags:       m.Tags(),
			Message: &loggregator_v2.Envelope_Counter{
				Counter: &loggregator_v2.Counter{
					Name:  "test-metric",
					Total: uint64(43),
				},
			},
		})
	})

	t.Run("defaults source_id if tag is not found", func(t *testing.T) {
		lc := newSpyLogCacheClient()
		cfOutput := &CloudFoundry{
			LogCacheClient: lc,
			SourceIDTag:    "source",
			SourceID:       "test-source-default",
			InstanceIDTag:  "instance",
		}

		m, _ := metric.New(
			"test-metric",
			map[string]string{
				"instance": "test-instance",
			},
			map[string]interface{}{
				"gauge": 42.0,
			},
			time.Now(),
		)

		err := cfOutput.Write([]telegraf.Metric{m})
		require.NoError(t, err)

		require.Contains(t, lc.envelopes(), &loggregator_v2.Envelope{
			Timestamp:  m.Time().UnixNano(),
			SourceId:   "test-source-default",
			InstanceId: "test-instance",
			Tags:       m.Tags(),
			Message: &loggregator_v2.Envelope_Gauge{
				Gauge: &loggregator_v2.Gauge{
					Metrics: map[string]*loggregator_v2.GaugeValue{
						"test-metric": {
							Unit:  "",
							Value: 42.0,
						},
					},
				},
			},
		})
	})

	t.Run("defaults instance_id if tag is not found", func(t *testing.T) {
		lc := newSpyLogCacheClient()
		cfOutput := &CloudFoundry{
			LogCacheClient: lc,
			SourceIDTag:    "source",
			InstanceIDTag:  "instance",
			InstanceID:     "test-instance-default",
		}

		m, _ := metric.New(
			"test-metric",
			map[string]string{
				"source": "test-source",
			},
			map[string]interface{}{
				"gauge": 42.0,
			},
			time.Now(),
		)

		err := cfOutput.Write([]telegraf.Metric{m})
		require.NoError(t, err)

		require.Contains(t, lc.envelopes(), &loggregator_v2.Envelope{
			Timestamp:  m.Time().UnixNano(),
			SourceId:   "test-source",
			InstanceId: "test-instance-default",
			Tags:       m.Tags(),
			Message: &loggregator_v2.Envelope_Gauge{
				Gauge: &loggregator_v2.Gauge{
					Metrics: map[string]*loggregator_v2.GaugeValue{
						"test-metric": {
							Unit:  "",
							Value: 42.0,
						},
					},
				},
			},
		})
	})
}

func newSpyLogCacheClient() *spyLogCacheClient {
	return &spyLogCacheClient{}
}

type spyLogCacheClient struct {
	sync.Mutex
	envs      []*loggregator_v2.Envelope
	sendError error
}

func (s *spyLogCacheClient) Send(ctx context.Context, in *logcache_v1.SendRequest, opts ...grpc.CallOption) (*logcache_v1.SendResponse, error) {
	s.Lock()
	defer s.Unlock()

	if s.sendError != nil {
		return nil, s.sendError
	}

	s.envs = append(s.envs, in.Envelopes.Batch...)
	return &logcache_v1.SendResponse{}, nil
}

func (s *spyLogCacheClient) envelopes() []*loggregator_v2.Envelope {
	s.Lock()
	defer s.Unlock()

	return s.envs
}
