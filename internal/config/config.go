package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"time"
)

const DEFAULT_TIMEOUT = time.Second * 5

type Duration time.Duration

type DomainConfig struct {
	SubDomains []string `json:"subdomains"`
}

type DomainMap map[string]DomainConfig

type Config struct {
	Domains  DomainMap `json:"domains"`
	Timeout  Duration  `json:"timeout"`
	Email    string    `json:"email"`
	APIToken string    `json:"api_token"`
	ZoneID   string    `json:"zone_id"`
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	parsedDuration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(parsedDuration)
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func getConfigFilePath() (string, error) {
	homePath := os.Getenv("HOME")
	configPath := ".config/ddns-updater/config.json"
	fullPath := path.Join(homePath, configPath)
	_, err := os.Stat(fullPath)
	return fullPath, err
}

func Get() (*Config, error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}
	defer f.Close()
	var config *Config
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to read and decode config file: %w", err)
	}
	if err := validateConfigAndSetDefaults(config); err != nil {
		return nil, fmt.Errorf("config file validation error: %w", err)
	}
	return config, nil
}

func validateConfigAndSetDefaults(cfg *Config) error {
	if cfg.Email == "" {
		return errors.New("required field: email is missing")
	}
	if cfg.APIToken == "" {
		return errors.New("required field: api_token is missing")
	}
	if len(cfg.Domains) == 0 {
		return errors.New("list of domains is empty")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = Duration(DEFAULT_TIMEOUT)
	}
	return nil
}
