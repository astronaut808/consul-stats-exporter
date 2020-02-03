package main

import (
	"net/http"

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
		"Consul ACL token for read Consul stats. [$CONSUL_HTTP_TOKEN]").
		Default("").String()
	consulTokenFile = kingpin.Flag("tokenfile",
		"File with Consul ACL token for read Consul stats. [$CONSUL_HTTP_TOKEN_FILE]").
		Default("").String()
	metricsPath = kingpin.Flag("web.telemetry-path",
		"Path under which to expose metrics.").
		Default("/metrics").String()
	sslInsecure = kingpin.Flag("insecure-ssl",
		"Set SSL to ignore certificate validation.").
		Default("false").Bool()
)

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
