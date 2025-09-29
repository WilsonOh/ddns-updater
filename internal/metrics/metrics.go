package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "ddns_updater"
	jobName   = "ddns_updater"
)

var (
	runsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "runs_total",
			Help:      "Total number of DDNS updater runs",
		},
		[]string{"status", "error_message"},
	)

	runDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "run_duration_seconds",
			Help:      "Duration of DDNS updater runs",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		},
		[]string{},
	)

	ipChanges = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "ip_changes_total",
			Help:      "Total number of IP address changes detected",
		},
		[]string{"old_ip", "new_ip", "hostname"},
	)
)

type MetricsRecorder interface {
	RecordRun(status string, duration time.Duration, err error)
	RecordIPChange(oldIP, newIP string)
	Push() error
}

func Init() {
	prometheus.DefaultRegisterer.MustRegister(
		runsTotal, runDuration, ipChanges,
	)
}

func NewRecorder(config *MetricsConfig) MetricsRecorder {
	if !config.Enabled {
		return NewNoopMetricsRecorder()
	}
	return NewPushGatewatMetricsRecorder(config)
}
