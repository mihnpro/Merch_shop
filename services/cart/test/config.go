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
	ProductsDB  pgConfig
	InventoryDB pgConfig
	CartDB      pgConfig

	ReadyTimeout time.Duration
}

func loadConfig() Config {
	password := env("POSTGRES_PASSWORD", "changeme")
	user := env("POSTGRES_USER", "postgres")
	ssl := env("POSTGRES_SSLMODE", "disable")

	pg := func(hostKey, portKey, port, dbKey, db string) pgConfig {
		return pgConfig{
			Host:     env(hostKey, "localhost"),
			Port:     env(portKey, port),
			User:     user,
			Password: password,
			Database: env(dbKey, db),
			SSLMode:  ssl,
		}
	}

	return Config{
		GatewayURL: env("GW", "http://localhost:8080"),
		APIPrefix:  "/api/v1",

		UserDB:      pg("USER_POSTGRES_HOST", "USER_POSTGRES_PORT", "5433", "USER_POSTGRES_DB", "merch_users"),
		ProductsDB:  pg("PRODUCTS_POSTGRES_HOST", "PRODUCTS_POSTGRES_PORT", "5434", "PRODUCTS_POSTGRES_DB", "merch_service"),
		InventoryDB: pg("INVENTORY_POSTGRES_HOST", "INVENTORY_POSTGRES_PORT", "5435", "INVENTORY_POSTGRES_DB", "merch_service"),
		CartDB:      pg("CART_POSTGRES_HOST", "CART_POSTGRES_PORT", "5436", "CART_POSTGRES_DB", "merch_cart"),

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
