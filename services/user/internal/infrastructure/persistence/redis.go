package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	URL      string
	Host     string
	Port     string
	Username string
	Password string
	DB       int
}

func NewRedis(ctx context.Context, cfg RedisConfig) (*redis.Client, error) {
	var client *redis.Client

	if cfg.URL != "" {
		opt, err := redis.ParseURL(cfg.URL)
		if err != nil {
			return nil, err
		}
		client = redis.NewClient(opt)
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
			Username:     cfg.Username,
			Password:     cfg.Password,
			DB:           cfg.DB,
			MaxRetries:   5,
			DialTimeout:  10 * time.Second,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		})
	}

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}
