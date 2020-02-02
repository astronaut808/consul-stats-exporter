package main

import (
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
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
	self, err := e.client.Agent().Self()
	if err != nil {
		return err
	}

	serfLan, ok := self["Stats"]["serf_lan"].(map[string]interface{})
	if !ok {
		return err
	}

	f, err := strconv.ParseFloat(serfLan["members"].(string), 64)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		consulLanMembers, prometheus.GaugeValue, f,
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

	ch <- prometheus.MustNewConstMetric(
		consulWanMembers, prometheus.GaugeValue, f,
	)

	return nil
}

func (e *Exporter) collectNodeStatusMetric(ch chan<- prometheus.Metric) error {
	self, err := e.client.Agent().Self()
	if err != nil {
		return err
	}

	nodeStatus := fmt.Sprintf("%v", self["Member"]["Status"])

	f, err := strconv.ParseFloat(nodeStatus, 64)

	ch <- prometheus.MustNewConstMetric(
		consulNodeStatus, prometheus.GaugeValue, f,
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

	if err := e.collectNodeStatusMetric(ch); err != nil {
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
