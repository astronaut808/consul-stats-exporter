package main

import (
	"os"

	consulApi "github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus"
)

const consulWaitTime = 3


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
