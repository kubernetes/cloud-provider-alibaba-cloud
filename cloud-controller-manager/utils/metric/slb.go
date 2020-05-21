package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

var (
	// SLBLatency reconcile SLB latency
	SLBLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:   "CCMLoadBalancerLatency",
			Help:   "ccm reconcile slb latency in ms",
			MaxAge: time.Duration(30) * time.Minute,
		},
		[]string{"verb"},
	)
)
