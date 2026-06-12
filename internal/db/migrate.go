package db

import (
	"database/sql"
	"embed"
	"log"
	"strings"
)

//go:embed migrations/*/*.sql
var migrationFiles embed.FS

func RunMigrations(sqlDB *sql.DB, dialect string) {
	_, err := sqlDB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY);`)
	if err != nil {
		log.Fatalf("Could not create schema_migrations table: %v", err)
	}

	dirPath := "migrations/" + dialect
	entries, err := migrationFiles.ReadDir(dirPath)
	if err != nil {
		log.Fatalf("Could not read migrations directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		var applied string
		err = sqlDB.QueryRow(`SELECT version FROM schema_migrations WHERE version = '` + entry.Name() + `'`).Scan(&applied)
		if err == nil {
			continue
		}

		log.Printf("Applying migration: %s", entry.Name())
		content, err := migrationFiles.ReadFile(dirPath + "/" + entry.Name())
		if err != nil {
			log.Fatalf("Could not read migration file %s: %v", entry.Name(), err)
		}

		tx, err := sqlDB.Begin()
		if err != nil {
			log.Fatalf("Could not start transaction for %s: %v", entry.Name(), err)
		}

		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err = tx.Exec(stmt); err != nil {
				tx.Rollback()
				log.Fatalf("Migration %s failed on statement %q: %v", entry.Name(), stmt, err)
			}
		}

		if _, err = tx.Exec(`INSERT INTO schema_migrations (version) VALUES ('` + entry.Name() + `')`); err != nil {
			tx.Rollback()
			log.Fatalf("Failed to record migration %s: %v", entry.Name(), err)
		}

		if err = tx.Commit(); err != nil {
			log.Fatalf("Failed to commit migration %s: %v", entry.Name(), err)
		}
	}
	log.Println("Database up to date.")
}
