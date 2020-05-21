package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

var (
	// NodeLatency reconcile node latency
	NodeLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:   "CCMNodeLatency",
			Help:   "ccm reconcile node latency in ms",
			MaxAge: time.Duration(30) * time.Minute,
		},
		[]string{"verb"},
	)
)
