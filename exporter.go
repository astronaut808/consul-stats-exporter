package main

import (
	"os"

	"github.com/prometheus/client_golang/prometheus"
	consulApi "github.com/hashicorp/consul/api"
)

const (
	namespace = "consul"
	consulWaitTime = 3
)

var (
	leader = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "stats_leader"),
		"Consul cluster leader",
		nil, nil,
	)

	lastScrapeError = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "stats_last_scrape_error"),
		"Failed to scrape metrics",
		nil, nil,
	)

	consulLanMembers = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "stats_lan_members_count"),
		"Consul Members",
		nil, nil,
	)

	consulWanMembers = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "stats_wan_members_count"),
		"Consul Members",
		nil, nil,
	)

	consulBootstrapExpect = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "stats_bootstap_expect"),
		"Consul Bootstrap Expect",
		nil, nil,
	)

	consulNodeStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "stats_node_status"),
		"Consul Node Status: alive, left or failed",
		nil, nil,
	)

	consulInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "stats_info"),
		"Consul Version",
		[]string{"version", "datacenter"}, nil,
	)
)

// Exporter collects Consul leader
type Exporter struct {
	hostname string
	client   *consulApi.Client
}

// NewExporter returns an initialized Exporter
func NewExporter() (*Exporter, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	consulConfig := consulApi.DefaultConfig()
	consulConfig.Address = *consulAddress
	consulConfig.Token = *consulToken
	consulConfig.WaitTime = consulWaitTime

	if *sslInsecure {
		consulConfig.TLSConfig.InsecureSkipVerify = true
	}

	client, err := consulApi.NewClient(consulConfig)
	if err != nil {
		return nil, err
	}

	return &Exporter{
		client:   client,
		hostname: hostname,
	}, nil
}

func init() {
	prometheus.MustRegister()
}

// Describe describes the metric ever exported by Consul Stats Exporter
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- leader
	ch <- consulInfo
	ch <- lastScrapeError
	ch <- consulLanMembers
	ch <- consulWanMembers
	ch <- consulNodeStatus
	ch <- consulBootstrapExpect
}