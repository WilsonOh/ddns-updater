package metrics

import (
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	// Application run metrics
	runsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ddns_runs_total",
			Help: "Total number of DDNS updater runs",
		},
		[]string{"status", "hostname"},
	)

	runDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ddns_run_duration_seconds",
			Help:    "Duration of DDNS updater runs",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		},
		[]string{"status", "hostname"},
	)

	// IP address metrics
	ipChanges = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ddns_ip_changes_total",
			Help: "Total number of IP address changes detected",
		},
		[]string{"old_ip", "new_ip", "hostname"},
	)

	ipFetchDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ddns_ip_fetch_duration_seconds",
			Help:    "Duration of IP address fetch operations",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2, 5},
		},
		[]string{"result", "hostname"},
	)

	// DNS record metrics
	dnsRecordsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ddns_dns_records_processed_total",
			Help: "Total number of DNS records processed",
		},
		[]string{"action", "domain", "hostname"},
	)

	dnsUpdateDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ddns_dns_update_duration_seconds",
			Help:    "Duration of DNS batch update operations",
			Buckets: []float64{0.1, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"result", "hostname"},
	)

	// Cloudflare API metrics
	cloudflareAPIDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ddns_cloudflare_api_duration_seconds",
			Help:    "Duration of Cloudflare API calls",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"operation", "result", "hostname"},
	)

	cloudflareAPIErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ddns_cloudflare_api_errors_total",
			Help: "Total number of Cloudflare API errors",
		},
		[]string{"operation", "error_type", "hostname"},
	)

	// Application info
	buildInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ddns_build_info",
			Help: "Build information",
		},
		[]string{"version", "go_version", "hostname"},
	)

	// Timing metrics
	lastRun = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ddns_last_run_timestamp_seconds",
			Help: "Timestamp of last run",
		},
		[]string{"status", "hostname"},
	)
)

type RunRecorder struct {
	startTime time.Time
	hostname  string
}

func NewRunRecorder() *RunRecorder {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	return &RunRecorder{
		startTime: time.Now(),
		hostname:  hostname,
	}
}

func (r *RunRecorder) RecordRun(status string) {
	duration := time.Since(r.startTime).Seconds()

	runsTotal.WithLabelValues(status, r.hostname).Inc()
	runDuration.WithLabelValues(status, r.hostname).Observe(duration)
	lastRun.WithLabelValues(status, r.hostname).Set(float64(time.Now().Unix()))
}

func (r *RunRecorder) RecordIPChange(oldIP, newIP string) {
	ipChanges.WithLabelValues(oldIP, newIP, r.hostname).Inc()
}

func (r *RunRecorder) RecordIPFetch(duration time.Duration, success bool) {
	result := "success"
	if !success {
		result = "failure"
	}
	ipFetchDuration.WithLabelValues(result, r.hostname).Observe(duration.Seconds())
}

func (r *RunRecorder) RecordDNSRecord(action, domain string) {
	dnsRecordsProcessed.WithLabelValues(action, domain, r.hostname).Inc()
}

func (r *RunRecorder) RecordDNSUpdate(duration time.Duration, success bool) {
	result := "success"
	if !success {
		result = "failure"
	}
	dnsUpdateDuration.WithLabelValues(result, r.hostname).Observe(duration.Seconds())
}

func (r *RunRecorder) RecordCloudflareAPI(operation string, duration time.Duration, err error) {
	result := "success"
	if err != nil {
		result = "failure"
		errorType := parseErrorType(err)
		cloudflareAPIErrors.WithLabelValues(operation, errorType, r.hostname).Inc()
	}
	cloudflareAPIDuration.WithLabelValues(operation, result, r.hostname).Observe(duration.Seconds())
}

func parseErrorType(err error) string {
	errStr := err.Error()
	switch {
	case contains(errStr, "timeout"):
		return "timeout"
	case contains(errStr, "401"), contains(errStr, "403"):
		return "auth"
	case contains(errStr, "429"):
		return "rate_limit"
	case contains(errStr, "context deadline exceeded"):
		return "timeout"
	default:
		return "other"
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// PushMetrics pushes all metrics to the Push Gateway
func (r *RunRecorder) PushMetrics(pushGatewayURL, jobName string) error {
	if pushGatewayURL == "" {
		// Skip pushing if no URL provided
		return nil
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		runsTotal,
		runDuration,
		ipChanges,
		ipFetchDuration,
		dnsRecordsProcessed,
		dnsUpdateDuration,
		cloudflareAPIDuration,
		cloudflareAPIErrors,
		buildInfo,
		lastRun,
	)

	pusher := push.New(pushGatewayURL, jobName).
		Gatherer(registry).
		Grouping("instance", r.hostname)

	return pusher.Push()
}

// SetBuildInfo sets build information
func SetBuildInfo(version, goVersion string) {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	buildInfo.WithLabelValues(version, goVersion, hostname).Set(1)
}
