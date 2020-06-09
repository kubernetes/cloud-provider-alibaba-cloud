package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// NodeLatency reconcile node latency
	NodeLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "ccm_node_latencies_duration_milliseconds",
			Help: "CCM node reconcile latency distribution in milliseconds for each verb.",
			Buckets: []float64{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000,
				1500, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000},
		},
		[]string{"verb"},
	)
)
