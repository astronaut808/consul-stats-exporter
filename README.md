# Consul Stats Exporter

Export [Hashicorp Consul](https://github.com/hashicorp/consul) cluster leader(followers) metrics to [Prometheus](https://github.com/prometheus/prometheus)

## Exported Metrics

Two metrics are currently available:

* `consul_stats_leader`: 1 - leader, 0 - follower.
* `consul_stats_last_scrape_error`: 1 - failed to scrape consul_stats leader metric, 0 - no scraping errors.
[Consul Operator Raft list-peers](https://www.consul.io/docs/commands/operator/raft.html#list-peers)

* `consul_stats_info` - Example: `consul_stats_info{version="1.5.3"} 1`
[Consul Agent Self](https://www.consul.io/api/agent.html#read-configuration)

## Flags

```bash
$ ./consul-stats-exporter --help
usage: consul-stats-exporter [<flags>]

Flags:
  -h, --help          Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":8313"  
                      Address to listen on for web interface and telemetry.
      --consul-address="http://127.0.0.1:8500"  
                      Consul agent address.
      --token=""      Consul ACL token. ACL required: `operator:read`,`agent:read` [$CONSUL_HTTP_TOKEN]
      --web.telemetry-path="/metrics"  
                      Path under which to expose metrics.
      --insecure-ssl  Set SSL to ignore certificate validation.
      --version       Show application version.
```
