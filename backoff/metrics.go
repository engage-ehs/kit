package backoff

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	boMaxRetries = prometheus.NewDesc("kit_backoff_retry_max", "Maximum number of retries for backoff", nil, nil)
	boNumRetries = prometheus.NewDesc("kit_backoff_num_retries", "Number of retries in a backoff", nil, nil)
)

// use a type to wrap prometheus metrics, so that they don’t show in the API
type exporter Backoff

func (b exporter) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(b, ch)
}

func (b exporter) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(boMaxRetries, prometheus.CounterValue, float64(b.MaxRetries))
	ch <- prometheus.MustNewConstMetric(boNumRetries, prometheus.CounterValue, float64(b.numRetries))
}

// Register exports a backoff so it will be scraped by Prometheus
func Register(b Backoff, name string) {
	prometheus.WrapRegistererWith(prometheus.Labels{"name": name}, prometheus.DefaultRegisterer).Register(exporter(b))
}
