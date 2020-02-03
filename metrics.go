package main

import (
	"fmt"
	"strconv"

	consulApi "github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const namespace = "consul"

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
		"Consul LAN Members",
		nil, nil,
	)

	consulLanFailedMembers = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "stats_lan_failed_members_count"),
		"Consul LAN Failed Members count",
		nil, nil,
	)

	consulLanLeftMembers = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "stats_lan_left_members_count"),
		"Consul LAN Left Members count",
		nil, nil,
	)

	consulWanMembers = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "stats_wan_members_count"),
		"Consul WAN Members",
		nil, nil,
	)
	consulWanFailedMembers = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "stats_wan_failed_members_count"),
		"Consul WAN Failed Members count",
		nil, nil,
	)

	consulWanLeftMembers = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "stats_wan_left_members_count"),
		"Consul WAN Left Members count",
		nil, nil,
	)
	consulBootstrapExpect = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "stats_bootstap_expect"),
		"Consul Bootstrap Expect",
		nil, nil,
	)

	consulServices = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "stats_services_count"),
		"Consul Services",
		nil, nil,
	)

	consulInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "stats_info"),
		"Consul Version",
		[]string{"version", "datacenter"}, nil,
	)
)

func bool2float(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// Collect fetches the stats from configured Consul and delivers them as Prom metrics
func (e *Exporter) collectLeaderMetric(ch chan<- prometheus.Metric) error {
	reply, err := e.client.Operator().RaftGetConfiguration(nil)
	if err != nil {
		return err
	}

	for _, server := range reply.Servers {
		if server.Node == e.hostname {
			ch <- prometheus.MustNewConstMetric(
				leader, prometheus.GaugeValue, bool2float(server.Leader),
			)
		}
	}

	return nil
}

func (e *Exporter) collectLanMembersMetric(ch chan<- prometheus.Metric) error {
	self, _, err := e.client.Catalog().Nodes(&consulApi.QueryOptions{})
	if err != nil {
		return err
	}

	f := float64(len(self))

	selfLan, err := e.client.Agent().Self()
	if err != nil {
		return err
	}

	serfLan, ok := selfLan["Stats"]["serf_lan"].(map[string]interface{})
	if !ok {
		return err
	}

	fl, err := strconv.ParseFloat(serfLan["failed"].(string), 64)
	if err != nil {
		return err
	}

	ll, err := strconv.ParseFloat(serfLan["left"].(string), 64)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		consulLanMembers, prometheus.GaugeValue, f,
	)

	ch <- prometheus.MustNewConstMetric(
		consulLanFailedMembers, prometheus.GaugeValue, fl,
	)

	ch <- prometheus.MustNewConstMetric(
		consulLanLeftMembers, prometheus.GaugeValue, ll,
	)

	return nil
}

func (e *Exporter) collectWanMembersMetric(ch chan<- prometheus.Metric) error {
	self, err := e.client.Agent().Self()
	if err != nil {
		return err
	}

	serfWan, ok := self["Stats"]["serf_wan"].(map[string]interface{})
	if !ok {
		return err
	}

	f, err := strconv.ParseFloat(serfWan["members"].(string), 64)
	if err != nil {
		return err
	}

	fw, err := strconv.ParseFloat(serfWan["failed"].(string), 64)
	if err != nil {
		return err
	}

	lw, err := strconv.ParseFloat(serfWan["left"].(string), 64)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		consulWanMembers, prometheus.GaugeValue, f,
	)

	ch <- prometheus.MustNewConstMetric(
		consulWanFailedMembers, prometheus.GaugeValue, fw,
	)

	ch <- prometheus.MustNewConstMetric(
		consulWanLeftMembers, prometheus.GaugeValue, lw,
	)
	return nil
}

func (e *Exporter) collectconsulServicesCountMetric(ch chan<- prometheus.Metric) error {
	self, _, err := e.client.Catalog().Services(nil)
	if err != nil {
		return err
	}

	f := float64(len(self))

	ch <- prometheus.MustNewConstMetric(
		consulServices, prometheus.GaugeValue, f,
	)

	return nil
}

func (e *Exporter) collectBootstrapExpectMetric(ch chan<- prometheus.Metric) error {
	self, err := e.client.Agent().Self()
	if err != nil {
		return nil
	}

	s := fmt.Sprintf("%v", self["DebugConfig"]["BootstrapExpect"])

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		consulBootstrapExpect, prometheus.GaugeValue, f,
	)

	return nil
}

// Collect version info from configured Consul and delivers them as Prom metrics
func (e *Exporter) collectConsulInfoMetric(ch chan<- prometheus.Metric) error {
	self, err := e.client.Agent().Self()
	if err != nil {
		return err
	}

	consulVersion := fmt.Sprintf("%v", self["Config"]["Version"])
	consulDataCenter := fmt.Sprintf("%v", self["Config"]["Datacenter"])

	ch <- prometheus.MustNewConstMetric(
		consulInfo, prometheus.GaugeValue, 1, consulVersion, consulDataCenter,
	)

	return nil
}

// Collect last scrape error
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	var scrapeError bool

	if err := e.collectLeaderMetric(ch); err != nil {
		scrapeError = true
		log.Error(err)
	}

	if err := e.collectConsulInfoMetric(ch); err != nil {
		scrapeError = true
		log.Error(err)
	}

	if err := e.collectLanMembersMetric(ch); err != nil {
		scrapeError = true
		log.Error(err)
	}

	if err := e.collectWanMembersMetric(ch); err != nil {
		scrapeError = true
		log.Error(err)
	}

	if err := e.collectBootstrapExpectMetric(ch); err != nil {
		scrapeError = true
		log.Error(err)
	}

	if err := e.collectconsulServicesCountMetric(ch); err != nil {
		scrapeError = true
		log.Error(err)
	}

	scrapeErrorFloat := 0.0
	if scrapeError {
		scrapeErrorFloat = 1.0
	}

	ch <- prometheus.MustNewConstMetric(
		lastScrapeError, prometheus.GaugeValue, scrapeErrorFloat,
	)
}

// Describe describes the metric ever exported by Consul Stats Exporter
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- leader
	ch <- consulInfo
	ch <- lastScrapeError
	ch <- consulLanMembers
	ch <- consulLanFailedMembers
	ch <- consulLanLeftMembers
	ch <- consulWanMembers
	ch <- consulWanFailedMembers
	ch <- consulWanLeftMembers
	ch <- consulServices
	ch <- consulBootstrapExpect
}
