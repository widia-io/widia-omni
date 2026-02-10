package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	DatabaseURL       string   `env:"DATABASE_URL,required"`
	SupabaseURL       string   `env:"SUPABASE_URL,required"`
	SupabaseServiceKey string  `env:"SUPABASE_SERVICE_KEY,required"`
	SupabaseJWTSecret string   `env:"SUPABASE_JWT_SECRET,required"`
	RedisURL          string   `env:"REDIS_URL" envDefault:"redis://localhost:6379"`
	StripeSecretKey   string   `env:"STRIPE_SECRET_KEY"`
	StripeWebhookSecret string `env:"STRIPE_WEBHOOK_SECRET"`
	ResendAPIKey      string   `env:"RESEND_API_KEY"`
	OpenRouterAPIKey  string   `env:"OPENROUTER_API_KEY"`
	OpenRouterModel   string   `env:"OPENROUTER_MODEL" envDefault:"anthropic/claude-sonnet-4"`
	Port              int      `env:"PORT" envDefault:"8080"`
	Env               string   `env:"ENV" envDefault:"development"`
	AllowedOrigins    []string `env:"ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000,http://localhost:5173"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) IsProd() bool {
	return c.Env == "production"
}
