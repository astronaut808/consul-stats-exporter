package main

import (
	"os"
	"net/http"
	_ "net/http/pprof"

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
	ch <- lastScrapeError
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

// Collect last scrape error
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	err := e.collectLeaderMetric(ch)
	if err != nil {
		log.Error(err)
	}

	ch <- prometheus.MustNewConstMetric(
		lastScrapeError, prometheus.GaugeValue, bool2float(err != nil),
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
