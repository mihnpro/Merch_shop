package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Route struct {
	Prefix string
	Target string
	Public bool
}

type Config struct {
	HTTPPort        string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration

	JWTAccessSecret string

	Routes []Route

	CORSAllowedOrigins []string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	routes, err := loadRoutes(getEnv("GATEWAY_CONFIG", "config.yaml"))
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		HTTPPort:        getEnv("HTTP_PORT", "8080"),
		ReadTimeout:     getEnvDuration("HTTP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getEnvDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
		ShutdownTimeout: getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),

		JWTAccessSecret: os.Getenv("JWT_ACCESS_SECRET"),

		Routes: routes,

		CORSAllowedOrigins: getEnvList("CORS_ALLOWED_ORIGINS", []string{"*"}),
	}

	if cfg.JWTAccessSecret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET is required")
	}
	return cfg, nil
}

type routesFile struct {
	Services map[string]string `yaml:"services"`
	Routes   []struct {
		Prefix  string `yaml:"prefix"`
		Service string `yaml:"service"`
		Public  bool   `yaml:"public"`
	} `yaml:"routes"`
}

func loadRoutes(path string) ([]Route, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read gateway config %q: %w", path, err)
	}

	var rf routesFile
	if err := yaml.Unmarshal(data, &rf); err != nil {
		return nil, fmt.Errorf("parse gateway config %q: %w", path, err)
	}
	if len(rf.Routes) == 0 {
		return nil, fmt.Errorf("gateway config %q: no routes defined", path)
	}

	seen := make(map[string]bool, len(rf.Routes))
	routes := make([]Route, 0, len(rf.Routes))
	for _, r := range rf.Routes {
		if r.Prefix == "" {
			return nil, fmt.Errorf("gateway config: route with empty prefix")
		}
		if seen[r.Prefix] {
			return nil, fmt.Errorf("gateway config: duplicate route prefix %q", r.Prefix)
		}
		seen[r.Prefix] = true

		target, ok := rf.Services[r.Service]
		if !ok {
			return nil, fmt.Errorf("gateway config: route %q references unknown service %q", r.Prefix, r.Service)
		}

		routes = append(routes, Route{Prefix: r.Prefix, Target: target, Public: r.Public})
	}
	return routes, nil
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
