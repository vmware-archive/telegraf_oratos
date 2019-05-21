# Parser Processor Plugin

This plugin takes a defined list of fields to analyze and performs a sentiment analysis on the values of those fields.

## Configuration
```toml
[[processors.sentiment]]
  ## The name of the fields whose value will be analyzed.
  analyze_fields = []
```

### Example:

```toml
[[processors.sentiment]]
  analyze_fields = ["title", "body"]
```

**Input**:
```
network_interface_throughput,hostname=backend.example.com lower=10i,upper=1000i,mean=500i,title=this is awesome,body=this is not great, 1502489900000000000
```

**Output**:
```
network_interface_throughput,hostname=backend.example.com lower=10i,upper=1000i,mean=500i,title=this is awesome,body=this is not great, sentiment_header=1,sentiment_body=0, 1502489900000000000
```


