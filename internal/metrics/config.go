package metrics

import (
	"errors"
	"os"
)

type MetricsConfig struct {
	Enabled        bool
	PushGatewayURL string
}

func LoadConfigFromEnv() (*MetricsConfig, error) {
	cfg := &MetricsConfig{
		Enabled:        os.Getenv("METRICS_ENABLED") == "true",
		PushGatewayURL: os.Getenv("PUSHGATEWAY_URL"),
	}
	if cfg.Enabled && cfg.PushGatewayURL == "" {
		return nil, errors.New("metrics is enabled but pushgateway url is empty")
	}
	return cfg, nil
}
