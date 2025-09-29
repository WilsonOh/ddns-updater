package metrics

import (
	"os"
	"runtime"
	"time"
)

// MetricsRecorder interface for recording metrics
type MetricsRecorder interface {
	RecordRun(status string)
	RecordIPChange(oldIP, newIP string)
	RecordIPFetch(duration time.Duration, success bool)
	RecordDNSRecord(action, domain string)
	RecordDNSUpdate(duration time.Duration, success bool)
	RecordCloudflareAPI(operation string, duration time.Duration, err error)
	SetBuildInfo(version, goVersion string)
	Push() error // Push metrics to Push Gateway
}

// MetricsConfig holds configuration for Push Gateway metrics
type MetricsConfig struct {
	Enabled        bool
	PushGatewayURL string
	PushGatewayJob string
	Version        string
}

// NewRecorder creates a Push Gateway metrics recorder
func NewRecorder(config MetricsConfig) MetricsRecorder {
	if !config.Enabled || config.PushGatewayURL == "" {
		return &NoOpRecorder{}
	}

	return &PushGatewayRecorder{
		recorder: NewRunRecorder(),
		config:   config,
	}
}

// LoadConfigFromEnv loads metrics configuration from environment variables
func LoadConfigFromEnv() MetricsConfig {
	return MetricsConfig{
		Enabled:        os.Getenv("METRICS_ENABLED") == "true",
		PushGatewayURL: os.Getenv("PUSHGATEWAY_URL"),
		PushGatewayJob: getEnvWithDefault("PUSHGATEWAY_JOB", "ddns-updater"),
		Version:        getEnvWithDefault("VERSION", "unknown"),
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// PushGatewayRecorder wraps RunRecorder for Push Gateway
type PushGatewayRecorder struct {
	recorder *RunRecorder
	config   MetricsConfig
}

func (p *PushGatewayRecorder) RecordRun(status string) {
	p.recorder.RecordRun(status)
}

func (p *PushGatewayRecorder) RecordIPChange(oldIP, newIP string) {
	p.recorder.RecordIPChange(oldIP, newIP)
}

func (p *PushGatewayRecorder) RecordIPFetch(duration time.Duration, success bool) {
	p.recorder.RecordIPFetch(duration, success)
}

func (p *PushGatewayRecorder) RecordDNSRecord(action, domain string) {
	p.recorder.RecordDNSRecord(action, domain)
}

func (p *PushGatewayRecorder) RecordDNSUpdate(duration time.Duration, success bool) {
	p.recorder.RecordDNSUpdate(duration, success)
}

func (p *PushGatewayRecorder) RecordCloudflareAPI(operation string, duration time.Duration, err error) {
	p.recorder.RecordCloudflareAPI(operation, duration, err)
}

func (p *PushGatewayRecorder) SetBuildInfo(version, goVersion string) {
	SetBuildInfo(version, goVersion)
}

func (p *PushGatewayRecorder) Push() error {
	return p.recorder.PushMetrics(p.config.PushGatewayURL, p.config.PushGatewayJob)
}

// NoOpRecorder does nothing when metrics are disabled
type NoOpRecorder struct{}

func (n *NoOpRecorder) RecordRun(status string)                                                 {}
func (n *NoOpRecorder) RecordIPChange(oldIP, newIP string)                                      {}
func (n *NoOpRecorder) RecordIPFetch(duration time.Duration, success bool)                      {}
func (n *NoOpRecorder) RecordDNSRecord(action, domain string)                                   {}
func (n *NoOpRecorder) RecordDNSUpdate(duration time.Duration, success bool)                    {}
func (n *NoOpRecorder) RecordCloudflareAPI(operation string, duration time.Duration, err error) {}
func (n *NoOpRecorder) SetBuildInfo(version, goVersion string)                                  {}
func (n *NoOpRecorder) Push() error                                                             { return nil }

// Helper to initialize build info
func InitializeBuildInfo(recorder MetricsRecorder, version string) {
	recorder.SetBuildInfo(version, runtime.Version())
}
