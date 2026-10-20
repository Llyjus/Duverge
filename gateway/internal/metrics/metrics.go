package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var RequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "service_operation_duration_seconds",
		Help:    "Duration of service operations by service, operation, and stage.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"service", "operation", "stage"},
)

func init() {
	prometheus.MustRegister(RequestDuration)
}
