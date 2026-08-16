package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
)

// Config holds all BFF service settings.
type Config struct {
	AppPort string `env:"APP_PORT" envDefault:"8080"`

	AuthURL      string `env:"AUTH_URL" envDefault:"http://auth-service:8000"`
	ProfileURL   string `env:"PROFILE_URL" envDefault:"http://profile-service:8080"`
	AdvertCmdURL string `env:"ADVERT_CMD_URL" envDefault:"http://advert-cmd-svc:8080"`
	AdvertQueryURL string `env:"ADVERT_QUERY_URL" envDefault:"http://advert-query:8080"`
	OrderURL     string `env:"ORDER_URL" envDefault:"http://order-svc:8080"`
	DeliveryURL  string `env:"DELIVERY_URL" envDefault:"http://delivery-service:8080"`
	BillingURL   string `env:"BILLING_URL" envDefault:"http://billing-svc:8080"`
	DialogURL    string `env:"DIALOG_URL" envDefault:"http://dialog-svc:8080"`

	InternalToken string `env:"INTERNAL_TOKEN" envDefault:""`
}

// Load parses environment variables into Config.
func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.InternalToken == "" {
		slog.Warn("INTERNAL_TOKEN is empty; internal endpoints may reject requests")
	}

	return &cfg, nil
}

// MustLoad is like Load but exits on error.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	return cfg
}
