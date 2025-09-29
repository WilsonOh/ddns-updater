package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

type pushGatewayMetricsRecorder struct {
	pusher *push.Pusher
}

func NewPushGatewatMetricsRecorder(config *MetricsConfig) MetricsRecorder {
	pusher := push.New(config.PushGatewayURL, jobName).Gatherer(prometheus.DefaultGatherer)
	return &pushGatewayMetricsRecorder{pusher: pusher}
}

func (p *pushGatewayMetricsRecorder) RecordRun(status string, duration time.Duration, err error) {
	runsTotal.WithLabelValues(status, err.Error()).Inc()
	runDuration.WithLabelValues(status).Observe(duration.Seconds())
}

func (p *pushGatewayMetricsRecorder) RecordIPChange(oldIP, newIP string) {
	ipChanges.WithLabelValues(oldIP, newIP).Inc()
}

func (p *pushGatewayMetricsRecorder) Push() error {
	return p.pusher.Push()
}
