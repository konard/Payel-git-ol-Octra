// Package storage wires the database and Redis connections for the monolith.
package storage

import (
	"context"

	"backend/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// OpenPostgres opens a GORM connection to PostgreSQL and runs AutoMigrate for
// every model.
func OpenPostgres(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := Migrate(db); err != nil {
		return nil, err
	}
	return db, nil
}

// Migrate runs AutoMigrate for every model. Exposed separately so tests can
// migrate an in-memory database.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(model.AllModels()...)
}

// OpenRedis parses a redis URL and returns a connected client.
func OpenRedis(ctx context.Context, url string) (*redis.Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
