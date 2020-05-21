package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

var (
	// RouteLatency reconcile route latency
	RouteLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:   "CCMRouteLatency",
			Help:   "ccm reconcile route latency in ms",
			MaxAge: time.Duration(30) * time.Minute,
		},
		[]string{"verb"},
	)
)
