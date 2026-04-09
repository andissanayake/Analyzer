package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	analyzeRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analyze_requests_total",
			Help: "Total number of analyze requests by status code.",
		},
		[]string{"status"},
	)
	analyzeDurationSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "analyze_duration_seconds",
			Help:    "Duration of analyze requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
	)
	linksCheckedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "links_checked_total",
			Help: "Total number of links checked during analysis.",
		},
	)
	linksInaccessibleTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "links_inaccessible_total",
			Help: "Total number of inaccessible links found during analysis.",
		},
	)
)

func init() {
	prometheus.MustRegister(
		analyzeRequestsTotal,
		analyzeDurationSeconds,
		linksCheckedTotal,
		linksInaccessibleTotal,
	)
}

func ObserveAnalyzeRequest(statusCode int, durationSeconds float64) {
	analyzeRequestsTotal.WithLabelValues(strconv.Itoa(statusCode)).Inc()
	analyzeDurationSeconds.Observe(durationSeconds)
}

func AddLinksChecked(count int) {
	if count <= 0 {
		return
	}
	linksCheckedTotal.Add(float64(count))
}

func AddLinksInaccessible(count int) {
	if count <= 0 {
		return
	}
	linksInaccessibleTotal.Add(float64(count))
}
