package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "httping_requests_total",
		Help: "Total number of HTTP requests sent",
	})

	requestsSuccessTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "httping_requests_success_total",
		Help: "Number of successful HTTP responses",
	})

	requestsFailureTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "httping_requests_failure_total",
		Help: "Number of failed HTTP requests",
	})

	requestDurationHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "httping_request_duration_seconds",
		Help:    "Histogram of HTTP request latencies",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms up to ~16s
	})

	targetInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "httping_target_info",
			Help: "Static info about the target being pinged",
		},
		[]string{"target"},
	)
)

func RegisterPrometheusMetrics(target string) {
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestsSuccessTotal)
	prometheus.MustRegister(requestsFailureTotal)
	prometheus.MustRegister(requestDurationHistogram)
	prometheus.MustRegister(targetInfo)

	targetInfo.With(prometheus.Labels{"target": target}).Set(1)
}

func StartPrometheusExporter(port int) {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
		if err != nil {
			panic(fmt.Sprintf("Failed to start Prometheus exporter: %v", err))
		}
	}()
}

func RecordSuccess(latencySeconds float64) {
	requestsTotal.Inc()
	requestsSuccessTotal.Inc()
	requestDurationHistogram.Observe(latencySeconds)
}

func RecordFailure() {
	requestsTotal.Inc()
	requestsFailureTotal.Inc()
}
