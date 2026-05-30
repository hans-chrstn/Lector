package db

import (
	"log"
	"os"
	"time"

	"github.com/user/lector/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(path string) {
	var err error

	if envPath := os.Getenv("DATABASE_PATH"); envPath != "" {
		path = envPath
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	DB, err = gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	sqlDB, _ := DB.DB()
	sqlDB.Exec("PRAGMA journal_mode=WAL;")
	sqlDB.Exec("PRAGMA busy_timeout=5000;")
	sqlDB.Exec("PRAGMA synchronous=NORMAL;")

	if path != ":memory:" {
		os.Chmod(path, 0600)
	}

	DB.AutoMigrate(&models.Document{}, &models.Chapter{}, &models.ReadingProgress{}, &models.Group{}, &models.CacheItem{}, &models.Bookmark{}, &models.Note{}, &models.Plugin{}, &models.LibraryPath{})

	setupFTS(DB)

	log.Println("Database initialized in Silent Mode")
}

func setupFTS(db *gorm.DB) {
	var ftsCheck int
	err := db.Raw("SELECT count(*) FROM sqlite_master WHERE name='document_search'").Scan(&ftsCheck).Error
	if err == nil && ftsCheck > 0 {
		return
	}

	if err := db.Exec(`CREATE VIRTUAL TABLE document_search USING fts5(
		id UNINDEXED,
		title,
		author,
		genres,
		synopsis,
		content='documents',
		content_rowid='id'
	)`).Error; err != nil {
		log.Printf("[DB] FTS5 search module not available on this system. Falling back to standard queries.")
		return
	}

	triggers := []string{
		`DROP TRIGGER IF EXISTS documents_ai`,
		`CREATE TRIGGER documents_ai AFTER INSERT ON documents BEGIN
			INSERT INTO document_search(rowid, title, author, genres, synopsis) 
			VALUES (new.id, new.title, new.author, new.genres, new.synopsis);
		END`,
		`DROP TRIGGER IF EXISTS documents_ad`,
		`CREATE TRIGGER documents_ad AFTER DELETE ON documents BEGIN
			INSERT INTO document_search(document_search, rowid, title, author, genres, synopsis) 
			VALUES('delete', old.id, old.title, old.author, old.genres, old.synopsis);
		END`,
		`DROP TRIGGER IF EXISTS documents_au`,
		`CREATE TRIGGER documents_au AFTER UPDATE ON documents BEGIN
			INSERT INTO document_search(document_search, rowid, title, author, genres, synopsis) 
			VALUES('delete', old.id, old.title, old.author, old.genres, old.synopsis);
			INSERT INTO document_search(rowid, title, author, genres, synopsis) 
			VALUES (new.id, new.title, new.author, new.genres, new.synopsis);
		END`,
	}

	for _, q := range triggers {
		db.Exec(q)
	}
}
