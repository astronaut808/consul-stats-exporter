# Consul Stats Exporter

Export [Hashicorp Consul](https://github.com/hashicorp/consul) cluster metrics to [Prometheus](https://github.com/prometheus/prometheus)

## Exported Metrics

Metrics are currently available:

* `consul_stats_leader`: 1 - leader, 0 - follower.
* `consul_stats_last_scrape_error`: 1 - failed to scrape consul_stats leader metric, 0 - no scraping errors.
* `consul_stats_lan_members_count`: number of lan members that a Consul agent knows about.
* `consul_stats_lan_failed_members_count`: number of lan members with status `failed`
* `consul_stats_lan_left_members_count`: number of lan members with status `left`
* `consul_stats_wan_members_count`: number of wan members that a Consul agent knows about.
* `consul_stats_wan_failed_members_count`: number of wan members with status `failed`
* `consul_stats_wan_left_members_count`: number of wan members with status `left`
* `consul_stats_services_count`: number of all known unique services.
* `consul_stats_bootstrap_expect`: number of expected servers in the datacenter.
* `consul_stats_info` - example: `consul_stats_info{datacenter="testdc",version="1.5.3"} 1`

## Docs

* [Operator Raft list-peers](https://www.consul.io/docs/commands/operator/raft.html#list-peers)
* [Agent Read Configuration](https://www.consul.io/api/agent.html#read-configuration)
* [Consul Catalog List Services](https://www.consul.io/docs/commands/catalog/services.html)
* [Consul Catalog List Nodes](https://www.consul.io/docs/commands/catalog/nodes.html)

## Flags

```bash
./consul-stats-exporter --help
usage: consul-stats-exporter [<flags>]

Flags:
  -h, --help          Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":8313"  
                      Address to listen on for web interface and telemetry.
      --consul-address="http://127.0.0.1:8500"  
                      Consul agent address.
      --token=""      Consul ACL token for read Consul stats. [$CONSUL_HTTP_TOKEN]
      --tokenfile=""  File with Consul ACL token for read Consul stats. [$CONSUL_HTTP_TOKENFILE]
      --web.telemetry-path="/metrics"  
                      Path under which to expose metrics.
      --insecure-ssl  Set SSL to ignore certificate validation.
      --version       Show application version.
```

## Information

Exporter tested on Consul 1.5.3. If using a different version and observing a problem, create an issue.
