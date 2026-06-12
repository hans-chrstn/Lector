package main

import (
	"log"
	"os"

	"github.com/user/lector/internal/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	log.Println("Starting Lector Database Migration Tool (SQLite -> PostgreSQL)")

	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "lector.db"
	}

	pgURL := os.Getenv("DATABASE_URL")
	if pgURL == "" {
		log.Fatalf("DATABASE_URL environment variable is required (e.g. postgres://user:pass@localhost:5432/lector)")
	}

	log.Printf("Connecting to SQLite at %s...", sqlitePath)
	sqliteDB, err := gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to SQLite: %v", err)
	}

	log.Printf("Connecting to Postgres...")
	os.Setenv("DB_DRIVER", "postgres")
	db.InitDB("")
	pgDB := db.DB

	direction := os.Getenv("MIGRATE_DIRECTION")
	if direction == "" {
		direction = "sqlite2postgres"
	}

	var sourceDB, targetDB *gorm.DB
	if direction == "postgres2sqlite" {
		log.Println("Mode: PostgreSQL -> SQLite")
		sourceDB = pgDB
		targetDB = sqliteDB
	} else {
		log.Println("Mode: SQLite -> PostgreSQL")
		sourceDB = sqliteDB
		targetDB = pgDB
	}

	if err := db.RunDataMigration(sourceDB, targetDB); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration completed successfully!")
}
