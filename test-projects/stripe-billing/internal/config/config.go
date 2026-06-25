package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	StripeSecretKey     string
	StripeWebhookSecret string
	DatabaseURL         string
	RedisURL            string
	ProPriceID          string
	APIQuotaPro         int    // default 10000
	Port                string // default "8080"
	JWTSecret           string
	DevMode             bool
}

// Load reads configuration from environment variables. It fails fast if the
// STRIPE_SECRET_KEY is a live key (sk_live_) or not a test key (sk_test_).
func Load() (*Config, error) {
	secretKey := os.Getenv("STRIPE_SECRET_KEY")
	if strings.HasPrefix(secretKey, "sk_live_") {
		return nil, fmt.Errorf("config: refusing to start with a live Stripe key (sk_live_); use a test key (sk_test_)")
	}
	if secretKey != "" && !strings.HasPrefix(secretKey, "sk_test_") {
		return nil, fmt.Errorf("config: STRIPE_SECRET_KEY must start with sk_test_; got unsupported key format")
	}

	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	databaseURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")
	proPriceID := os.Getenv("PRO_PRICE_ID")
	jwtSecret := os.Getenv("JWT_SECRET")

	apiQuotaPro := 10000
	if v := os.Getenv("API_QUOTA_PRO"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("config: API_QUOTA_PRO must be an integer: %w", err)
		}
		apiQuotaPro = n
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	devMode := os.Getenv("DEV_MODE") == "true" || os.Getenv("DEV_MODE") == "1"

	cfg := &Config{
		StripeSecretKey:     secretKey,
		StripeWebhookSecret: webhookSecret,
		DatabaseURL:         databaseURL,
		RedisURL:            redisURL,
		ProPriceID:          proPriceID,
		APIQuotaPro:         apiQuotaPro,
		Port:                port,
		JWTSecret:           jwtSecret,
		DevMode:             devMode,
	}

	return cfg, nil
}
