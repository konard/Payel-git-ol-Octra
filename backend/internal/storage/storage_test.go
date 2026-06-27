package storage

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMigrate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate() returned error: %v", err)
	}

	hasTable := func(name string) bool {
		return db.Migrator().HasTable(name)
	}

	tables := []string{"users", "agents", "skills", "user_skills", "transactions", "user_api_keys", "dashboard_environments", "usage_metrics"}
	for _, table := range tables {
		if !hasTable(table) {
			t.Fatalf("table %s was not created by Migrate", table)
		}
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("second Migrate should be idempotent: %v", err)
	}
}
