package tests

import (
	"os"
	"testing"
	"time"

	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
)

func TestBidirectionalMigration(t *testing.T) {
	pgURL := os.Getenv("DATABASE_URL")
	if pgURL == "" {
		t.Skip("Skipping Postgres integration tests. Set DATABASE_URL to enable.")
	}

	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DATABASE_PATH", ":memory:")
	db.InitDB(":memory:")
	sqliteDB := db.DB

	os.Setenv("DB_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", pgURL)
	db.InitDB("")
	pgDB := db.DB

	seedDoc := models.Document{Title: "Migration Test Document", URL: "https://example.com/migtest"}
	sqliteDB.Create(&seedDoc)

	seedChap := models.Chapter{DocumentID: seedDoc.ID, Title: "Chapter 1", URL: "/chap1"}
	sqliteDB.Create(&seedChap)

	seedProg := models.ReadingProgress{DocumentID: seedDoc.ID, ChapterID: seedChap.ID, ScrollPos: 55.5, ClientUpdatedAt: time.Now().UnixMilli()}
	sqliteDB.Create(&seedProg)

	seedGroup := models.Group{Name: "Migration Testers"}
	sqliteDB.Create(&seedGroup)

	err := db.RunDataMigration(sqliteDB, pgDB)
	if err != nil {
		t.Fatalf("Migration SQLite->Postgres failed: %v", err)
	}

	var pgDoc models.Document
	if err := pgDB.First(&pgDoc).Error; err != nil {
		t.Fatalf("Failed to find document in Postgres: %v", err)
	}
	if pgDoc.Title != "Migration Test Document" {
		t.Errorf("Expected 'Migration Test Document', got '%s'", pgDoc.Title)
	}

	sqliteDB.Exec("DELETE FROM documents")
	sqliteDB.Exec("DELETE FROM chapters")
	sqliteDB.Exec("DELETE FROM reading_progresses")
	sqliteDB.Exec("DELETE FROM groups")

	err = db.RunDataMigration(pgDB, sqliteDB)
	if err != nil {
		t.Fatalf("Migration Postgres->SQLite failed: %v", err)
	}

	var backDoc models.Document
	if err := sqliteDB.First(&backDoc).Error; err != nil {
		t.Fatalf("Failed to find document in SQLite after reverse migration: %v", err)
	}
	if backDoc.Title != "Migration Test Document" {
		t.Errorf("Expected 'Migration Test Document' after reverse migration, got '%s'", backDoc.Title)
	}

	pgDB.Exec("DELETE FROM reading_progresses")
	pgDB.Exec("DELETE FROM chapters")
	pgDB.Exec("DELETE FROM documents")
	pgDB.Exec("DELETE FROM groups")
}
