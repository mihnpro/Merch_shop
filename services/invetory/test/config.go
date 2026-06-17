package e2e

import (
	"fmt"
	"os"
	"time"
)

type pgConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func (c pgConfig) dsn() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

type Config struct {
	GatewayURL string
	APIPrefix  string

	UserDB      pgConfig
	InventoryDB pgConfig

	ReadyTimeout time.Duration
}

func loadConfig() Config {
	password := env("POSTGRES_PASSWORD", "changeme")
	user := env("POSTGRES_USER", "postgres")
	ssl := env("POSTGRES_SSLMODE", "disable")

	return Config{
		GatewayURL: env("GW", "http://localhost:8080"),
		APIPrefix:  "/api/v1",

		UserDB: pgConfig{
			Host:     env("USER_POSTGRES_HOST", "localhost"),
			Port:     env("USER_POSTGRES_PORT", "5433"),
			User:     user,
			Password: password,
			Database: env("USER_POSTGRES_DB", "merch_users"),
			SSLMode:  ssl,
		},
		InventoryDB: pgConfig{
			Host:     env("INVENTORY_POSTGRES_HOST", "localhost"),
			Port:     env("INVENTORY_POSTGRES_PORT", "5435"),
			User:     user,
			Password: password,
			Database: env("INVENTORY_POSTGRES_DB", "merch_service"),
			SSLMode:  ssl,
		},

		ReadyTimeout: 30 * time.Second,
	}
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
