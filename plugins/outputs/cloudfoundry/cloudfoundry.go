package cloudfoundry

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/outputs"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
)

const (
	defaultURL = ":8080"
)

var sampleConfig = `
  ## Log Cache address
  address = ":8080"

  ## Log Cache requires TLS options to function
  # tls_ca = "/etc/log-cache/ca.pem"
  # tls_cert = "/etc/log-cache/cert.pem"
  # tls_key = "/etc/log-cache/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Metric tag to map to CF source_id
  source_id_tag = "host" # required
  ## Metric tag to map to CF instance_id
  instance_id_tag = "host" # required

  ## Default source_id if tag isn't present
  source_id = "host" # required
  ## Default instance_id if tag isn't present
  instance_id = "host" # required
`

type logCacheClient interface {
	Send(ctx context.Context, in *logcache_v1.SendRequest, opts ...grpc.CallOption) (*logcache_v1.SendResponse, error)
}

type CloudFoundry struct {
	LogCacheClient logCacheClient
	Address        string `toml:"address"`

	tls.ClientConfig
	SourceIDTag   string `toml:"source_id_tag"`
	InstanceIDTag string `toml:"instance_id_tag"`
	SourceID      string `toml:"source_id"`
	InstanceID    string `toml:"instance_id"`
	connection    *grpc.ClientConn
}

func (c *CloudFoundry) Connect() error {
	cfg, err := c.TLSConfig()
	if err != nil {
		return fmt.Errorf("Unable to create TLS configuration: %v\n", err)
	}
	tc := credentials.NewTLS(cfg)

	c.connection, err = grpc.Dial(
		c.Address,
		grpc.WithTransportCredentials(tc),
	)
	if err != nil {
		return fmt.Errorf("Unable to establish gRPC connection to log-cache: %v\n", err)
	}

	c.LogCacheClient = logcache_v1.NewIngressClient(c.connection)
	return nil
}

func (c *CloudFoundry) Close() error {
	return c.connection.Close()
}

func (c *CloudFoundry) Description() string {
	return "A plugin that can transmit metrics to CloudFoundry"
}

func (c *CloudFoundry) SampleConfig() string {
	return sampleConfig
}

func (c *CloudFoundry) Write(metrics []telegraf.Metric) error {
	b := make([]*loggregator_v2.Envelope, 0)
	for _, m := range metrics {
		switch {
		case m.HasField("gauge"):
			b = append(b, c.getGaugeEnvelope(m))
		case m.HasField("counter"):
			b = append(b, c.getCounterEnvelope(m))
		}
	}

	_, err := c.LogCacheClient.Send(context.TODO(), &logcache_v1.SendRequest{
		Envelopes: &loggregator_v2.EnvelopeBatch{Batch: b},
	})

	return err
}

func init() {
	outputs.Add("cloudfoundry", func() telegraf.Output {
		return &CloudFoundry{}
	})
}

func (c *CloudFoundry) getGaugeEnvelope(m telegraf.Metric) *loggregator_v2.Envelope {
	g, _ := m.GetField("gauge")
	value, _ := g.(float64)
	return &loggregator_v2.Envelope{
		Timestamp:  m.Time().UnixNano(),
		SourceId:   c.getSourceID(m),
		InstanceId: c.getInstanceID(m),
		Tags:       m.Tags(),
		Message: &loggregator_v2.Envelope_Gauge{
			Gauge: &loggregator_v2.Gauge{
				Metrics: map[string]*loggregator_v2.GaugeValue{
					m.Name(): {
						Unit:  "",
						Value: value,
					},
				},
			},
		},
	}
}

func (c *CloudFoundry) getCounterEnvelope(m telegraf.Metric) *loggregator_v2.Envelope {
	g, _ := m.GetField("counter")
	value, _ := g.(uint64)
	return &loggregator_v2.Envelope{
		Timestamp:  m.Time().UnixNano(),
		SourceId:   c.getSourceID(m),
		InstanceId: c.getInstanceID(m),
		Tags:       m.Tags(),
		Message: &loggregator_v2.Envelope_Counter{

			Counter: &loggregator_v2.Counter{
				Name:  m.Name(),
				Total: value,
			},
		},
	}
}

func (c *CloudFoundry) getSourceID(m telegraf.Metric) string {
	if s, ok := m.GetTag(c.SourceIDTag); ok {
		return s
	}

	return c.SourceID
}

func (c *CloudFoundry) getInstanceID(m telegraf.Metric) string {
	if i, ok := m.GetTag(c.InstanceIDTag); ok {
		return i
	}

	return c.InstanceID
}
