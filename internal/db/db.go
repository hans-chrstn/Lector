package db

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(path string) {
	var err error

	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "sqlite"
	}

	if envPath := os.Getenv("DATABASE_PATH"); envPath != "" {
		path = envPath
	}

	isMemory := strings.Contains(path, ":memory:")

	if isMemory && driver == "sqlite" {
		path = filepath.Join(os.TempDir(), "lector_test_"+time.Now().Format("20060102150405999999999")+".db")
		isMemory = false
	}

	if driver == "sqlite" {
		dir := filepath.Dir(path)
		if dir != "." && dir != "" && !isMemory {
			os.MkdirAll(dir, 0777)
		}
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

	var dialector gorm.Dialector

	if driver == "postgres" {
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			log.Fatalf("DATABASE_URL environment variable is required for postgres driver")
		}
		dialector = postgres.Open(dsn)
	} else {
		dsn := path
		if !isMemory {
			if !strings.Contains(path, "?") {
				dsn = path + "?_journal_mode=WAL&_busy_timeout=5000"
			} else {
				dsn = path + "&_journal_mode=WAL&_busy_timeout=5000"
			}
		}
		dialector = sqlite.Open(dsn)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	sqlDB, _ := DB.DB()

	if driver == "postgres" {
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetMaxIdleConns(20)
		sqlDB.SetConnMaxLifetime(time.Hour)
	} else if isMemory && driver == "sqlite" {
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
		sqlDB.SetConnMaxLifetime(0)
	} else {
		sqlDB.SetMaxOpenConns(1)
	}

	RunMigrations(sqlDB, driver)

	if driver == "sqlite" && !isMemory {
		sqlDB.Exec("PRAGMA journal_mode=WAL;")
		sqlDB.Exec("PRAGMA busy_timeout=5000;")
		sqlDB.Exec("PRAGMA synchronous=NORMAL;")
		os.Chmod(path, 0600)
	}

	setupFTS(DB, driver)

	log.Println("Database initialized in Silent Mode")
}

func setupFTS(db *gorm.DB, driver string) {
	if driver == "postgres" {
		return
	}

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
