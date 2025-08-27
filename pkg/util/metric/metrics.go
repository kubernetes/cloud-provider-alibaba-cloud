package metric

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

type SLBType string

var (
	CLBType = "CLBType"
	NLBType = "NLBType"
	ALBType = "ALBType"
)

type Verb string

var (
	VerbCreation = "Creation"
	VerbDeletion = "Deletion"
	VerbUpdate   = "Update"
)

type OperationResult string

var (
	ResultFail    = "Fail"
	ResultSuccess = "Success"
)

var (
	// NodeLatency reconcile node latency
	NodeLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "ccm_node_latencies_duration_milliseconds",
			Help: "CCM node reconcile latency distribution in milliseconds for each verb.",
			Buckets: []float64{500, 700, 1000, 1500, 2000, 3000, 4000, 5000, 7000, 8000, 9000, 10000,
				30000, 60000, 100000, 150000, 200000, 300000, 600000, 700000, 800000},
		},
		[]string{"verb"},
	)

	// RouteLatency reconcile route latency
	RouteLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "ccm_route_latencies_duration_milliseconds",
			Help: "CCM route reconcile latency distribution in milliseconds for each verb.",
			Buckets: []float64{500, 700, 1000, 1500, 2000, 3000, 4000, 5000, 7000, 8000, 9000, 10000,
				30000, 60000, 100000, 150000, 200000, 300000, 600000, 700000, 800000},
		},
		[]string{"verb"},
	)
	// SLBLatency reconcile SLB latency
	SLBLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "ccm_slb_latencies_duration_milliseconds",
			Help: "CCM load balancer reconcile latency distribution in milliseconds for each verb.",
			Buckets: []float64{500, 700, 1000, 1500, 2000, 3000, 4000, 5000, 7000, 8000, 9000, 10000,
				30000, 60000, 100000, 150000, 200000, 300000, 600000, 700000, 800000},
		},
		[]string{"type", "verb"},
	)

	// SLBOperationStatus counts verb status for SLB operation
	SLBOperationStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ccm_slb_operation_result",
			Help: "CCM load balancer operation result",
		},
		[]string{"type", "verb", "status"},
	)
)

// MsSince returns milliseconds since start.
func MsSince(start time.Time) float64 {
	return float64(time.Since(start) / time.Millisecond)
}

var serviceUIDMap = sync.Map{}

func UniqueServiceCnt(uid string) float64 {
	if _, ok := serviceUIDMap.Load(uid); !ok {
		return 0
	}
	serviceUIDMap.Store(uid, true)
	return 1
}

// RegisterPrometheus register metrics to prometheus server
func RegisterPrometheus() {
	metrics.Registry.MustRegister(RouteLatency)
	metrics.Registry.MustRegister(NodeLatency)
	metrics.Registry.MustRegister(SLBLatency)
	metrics.Registry.MustRegister(SLBOperationStatus)
}
