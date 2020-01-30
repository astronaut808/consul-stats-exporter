package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"

	consulApi "github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress = kingpin.Flag("web.listen-address",
		"Address to listen on for web interface and telemetry.").
		Default(":8313").String()
	consulAddress = kingpin.Flag("consul-address",
		"Consul agent address.").
		Default("http://127.0.0.1:8500").String()
	consulToken = kingpin.Flag("token",
		"Consul ACL token for read raft peers. [$CONSUL_HTTP_TOKEN]").
		Default("").String()
	metricsPath = kingpin.Flag("web.telemetry-path",
		"Path under which to expose metrics.").
		Default("/metrics").String()
	sslInsecure = kingpin.Flag("insecure-ssl",
		"Set SSL to ignore certificate validation.").
		Default("false").Bool()
)

const (
	namespace      = "consul"
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

	consulMembers = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "stats_members_count"),
		"Consul Members",
		nil, nil,
	)

	consulBootstrapExpect = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "stats_bootstap_expect"),
		"Consul Bootstrap Expect",
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

// Describe describes the metric ever exported by Consul Stats Exporter
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- leader
	ch <- consulInfo
	ch <- lastScrapeError
	ch <- consulMembers
	ch <- consulBootstrapExpect
}

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

func (e *Exporter) collectMembersMetric(ch chan<- prometheus.Metric) error {
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
		consulMembers, prometheus.GaugeValue, f,
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

	if err := e.collectMembersMetric(ch); err != nil {
		scrapeError = true
		log.Error(err)
	}

	if err := e.collectBootstrapExpectMetric(ch); err != nil {
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

func init() {
	prometheus.MustRegister()
}

func main() {
	kingpin.Version(Version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting Consul Stats Exporter", Version)
	log.Infoln("Consul Address:", *consulAddress)

	exporter, err := NewExporter()
	if err != nil {
		log.Fatalln(err)
	}

	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
             <head><title>Consul Stats Exporter</title></head>
             <body>
             <h1>Consul Stats Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
			 </html>`))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Infoln("Listening on", *listenAddress+*metricsPath)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
