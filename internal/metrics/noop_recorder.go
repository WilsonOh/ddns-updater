package metrics

import "time"

type noOpRecorder struct{}

func (n *noOpRecorder) RecordRun(status string, duration time.Duration, err error) {}
func (n *noOpRecorder) RecordIPChange(oldIP, newIP string)                         {}
func (n *noOpRecorder) Push() error                                                { return nil }

func NewNoopMetricsRecorder() MetricsRecorder {
	return &noOpRecorder{}
}
