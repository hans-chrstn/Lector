package db

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/lector/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type SchemaVersion struct {
	Version   int       `gorm:"primaryKey" json:"version"`
	AppliedAt time.Time `json:"applied_at"`
}

func runMigrations(db *gorm.DB) {
	db.AutoMigrate(&SchemaVersion{})

	var currentVersion int
	db.Model(&SchemaVersion{}).Select("COALESCE(MAX(version), 0)").Scan(&currentVersion)

	if currentVersion < 1 {
		db.AutoMigrate(
			&models.Document{},
			&models.Chapter{},
			&models.ReadingProgress{},
			&models.Group{},
			&models.CacheItem{},
			&models.Bookmark{},
			&models.Note{},
			&models.Plugin{},
			&models.LibraryPath{},
		)
		setupFTS(db)
		db.Create(&SchemaVersion{Version: 1, AppliedAt: time.Now()})
	}

	if currentVersion < 2 {
		db.AutoMigrate(&models.ReadingStat{})
		db.Create(&SchemaVersion{Version: 2, AppliedAt: time.Now()})
	}
}

func InitDB(path string) {
	var err error

	if envPath := os.Getenv("DATABASE_PATH"); envPath != "" {
		path = envPath
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		os.MkdirAll(dir, 0777)
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

	dsn := path
	if !strings.Contains(path, ":memory:") {
		if !strings.Contains(path, "?") {
			dsn = path + "?_journal_mode=WAL&_busy_timeout=5000"
		} else {
			dsn = path + "&_journal_mode=WAL&_busy_timeout=5000"
		}
	}

	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	sqlDB, _ := DB.DB()
	sqlDB.Exec("PRAGMA journal_mode=WAL;")
	sqlDB.Exec("PRAGMA busy_timeout=5000;")
	sqlDB.Exec("PRAGMA synchronous=NORMAL;")

	if !strings.Contains(path, ":memory:") {
		os.Chmod(path, 0600)
	}

	runMigrations(DB)

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
