package config

import (
	"errors"
	"os"
)

type Config struct {
	YandexIAMToken string
	YandexOrgID    string
	TrackerHost    string
	ServerAddr     string
}

func Load() (*Config, error) {
	config := &Config{
		YandexIAMToken: os.Getenv("YANDEX_IAM_TOKEN"),
		YandexOrgID:    os.Getenv("YANDEX_ORG_ID"),
		TrackerHost:    os.Getenv("TRACKER_HOST"),
		ServerAddr:     getEnvOrDefault("SERVER_ADDR", ":8080"),
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) validate() error {
	if c.YandexIAMToken == "" {
		return errors.New("YANDEX_IAM_TOKEN is required")
	}
	if c.YandexOrgID == "" {
		return errors.New("YANDEX_ORG_ID is required")
	}
	if c.TrackerHost == "" {
		return errors.New("TRACKER_HOST is required")
	}
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
