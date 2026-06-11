package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresConfig struct {
	URL      string
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (c PostgresConfig) dsn() string {
	if c.URL != "" {
		return c.URL
	}
	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		c.User, c.Password, c.DBName, c.Host, c.Port, c.SSLMode)
}

func NewPostgres(ctx context.Context, cfg PostgresConfig) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", cfg.dsn())
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(60 * time.Minute)
	db.SetConnMaxIdleTime(30 * time.Minute)
	return db, nil
}
