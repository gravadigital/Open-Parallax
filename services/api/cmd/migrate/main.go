package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gravadigital/openparallax/api/internal/config"
	"github.com/gravadigital/openparallax/api/internal/logger"
	"github.com/gravadigital/openparallax/api/internal/storage"
	"github.com/gravadigital/openparallax/api/internal/storage/migrations"
	"github.com/gravadigital/openparallax/api/internal/storage/postgres"
)

func main() {
	cfg := config.Load()

	logger.Initialize("info")
	log := logger.Migration()

	rollback := flag.Bool("rollback", false, "Rollback the last migration")
	flag.Parse()

	log.Info("Starting migration process", "rollback", *rollback)

	db, err := postgres.Connect(cfg)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	if *rollback {
		log.Info("Rolling back migrations...")
		if err := migrations.RollbackMigration(db); err != nil {
			log.Error("Migration rollback failed", "error", err)
			os.Exit(1)
		}
		log.Info("Migration rollback completed successfully")
	} else {
		log.Info("Running migrations...")
		// The files of the records a migration removes are deleted through the file storage; if it
		// is not reachable they are logged and the migration goes on (DEC-10).
		var files migrations.FileRemover
		if fileStorage, err := storage.NewFileStorage(cfg); err != nil {
			log.Warn("File storage unavailable: files of removed records will only be logged", "error", err)
		} else {
			files = fileStorage
		}
		if err := migrations.RunMigrations(db, files); err != nil {
			log.Error("Migration failed", "error", err)
			os.Exit(1)
		}
		log.Info("Migrations completed successfully")
	}

	fmt.Println("Migration process completed!")
}
