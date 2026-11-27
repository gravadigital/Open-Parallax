//go:build integration
// +build integration

package main

import (
	"os"
	"testing"

	"github.com/gravadigital/openparallax/api/internal/config"
	"github.com/gravadigital/openparallax/api/internal/storage/postgres"
	"github.com/stretchr/testify/assert"
)

// Integration tests that require a real PostgreSQL database
// Run with: go test -tags=integration

func TestDatabaseConnection(t *testing.T) {
	cfg := config.Load()

	if testDB := os.Getenv("TEST_DB_NAME"); testDB != "" {
		cfg.DB.Name = testDB
	}

	db, err := postgres.Connect(cfg)
	assert.NoError(t, err, "Should be able to connect to test database")

	if err == nil {
		sqlDB, err := db.DB()
		assert.NoError(t, err)

		err = sqlDB.Ping()
		assert.NoError(t, err, "Should be able to ping the database")

		sqlDB.Close()
	}
}

func TestDatabaseMigration(t *testing.T) {
	cfg := config.Load()

	if testDB := os.Getenv("TEST_DB_NAME"); testDB != "" {
		cfg.DB.Name = testDB
	}

	db, err := postgres.Connect(cfg)
	assert.NoError(t, err, "Should be able to connect to test database")

	if err == nil {
		err = postgres.AutoMigrate(db, nil)
		assert.NoError(t, err, "Should be able to run migrations")

		sqlDB, _ := db.DB()
		sqlDB.Close()
	}
}
