package e2e

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	GatewayURL string
	APIPrefix  string

	PGHost     string
	PGPort     string
	PGUser     string
	PGPassword string
	PGDatabase string
	PGSSLMode  string

	ReadyTimeout time.Duration
}

func loadConfig() Config {
	return Config{
		GatewayURL: env("GW", "http://localhost:8080"),
		APIPrefix:  "/api/v1",

		PGHost:     env("POSTGRES_HOST", "localhost"),
		PGPort:     env("POSTGRES_PORT", "5433"),
		PGUser:     env("POSTGRES_USER", "postgres"),
		PGPassword: env("POSTGRES_PASSWORD", "changeme"),
		PGDatabase: env("POSTGRES_DB", "merch_users"),
		PGSSLMode:  env("POSTGRES_SSLMODE", "disable"),

		ReadyTimeout: 30 * time.Second,
	}
}

func (c Config) dsn() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.PGHost, c.PGPort, c.PGUser, c.PGPassword, c.PGDatabase, c.PGSSLMode,
	)
}

func (c Config) apiURL(path string) string {
	return c.GatewayURL + c.APIPrefix + path
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
