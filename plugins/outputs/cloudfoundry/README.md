# Cloud Foundry Output Plugin

This plugin sends metrics to a CF Log Cache

### Configuration:

```toml
# A plugin that can transmit metrics over HTTP
[[outputs.cloudfoundry]]
  ## Log Cache address
  address = ":8080"

  ## TLS configuration for Log Cache
  ## CA path
  ca_path = "/ca.pem" # required
  ## Certificate path
  cert_path = "/cert.pem" # required
  ## Key path
  key_path = "/key.pem" # required

  ## Metric tag to map to CF source_id
  source_id_tag = "host" # required
  ## Metric tag to map to CF instance_id
  instance_id_tag = "host" # required

  ## Default source_id if tag isn't present
  source_id = "host" # required
  ## Default instance_id if tag isn't present
  instance_id = "host" # required
```
