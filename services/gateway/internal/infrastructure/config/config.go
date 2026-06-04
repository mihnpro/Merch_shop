package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Route struct {
	Prefix string
	Target string
}

type Config struct {
	HTTPPort        string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration

	JWTAccessSecret string

	Routes         []Route
	PublicPrefixes []string

	CORSAllowedOrigins []string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		HTTPPort:        getEnv("HTTP_PORT", "8080"),
		ReadTimeout:     getEnvDuration("HTTP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getEnvDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
		ShutdownTimeout: getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),

		JWTAccessSecret: os.Getenv("JWT_ACCESS_SECRET"),

		Routes:         parseRoutes(os.Getenv("GATEWAY_ROUTES")),
		PublicPrefixes: getEnvList("GATEWAY_PUBLIC_PREFIXES", nil),

		CORSAllowedOrigins: getEnvList("CORS_ALLOWED_ORIGINS", []string{"*"}),
	}

	if cfg.JWTAccessSecret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET is required")
	}
	if len(cfg.Routes) == 0 {
		return nil, fmt.Errorf("GATEWAY_ROUTES is required (format: /prefix=http://host:port,...)")
	}
	return cfg, nil
}

func parseRoutes(v string) []Route {
	var routes []Route
	for _, pair := range strings.Split(v, ",") {
		prefix, target, ok := strings.Cut(strings.TrimSpace(pair), "=")
		prefix, target = strings.TrimSpace(prefix), strings.TrimSpace(target)
		if !ok || prefix == "" || target == "" {
			continue
		}
		routes = append(routes, Route{Prefix: prefix, Target: target})
	}
	return routes
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvList(key string, def []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	var out []string
	for _, p := range strings.Split(v, ",") {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return def
	}
	return out
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
