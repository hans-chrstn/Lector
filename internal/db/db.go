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

	if path != ":memory:" {
		os.Chmod(path, 0600)
	}

	DB.AutoMigrate(&models.Document{}, &models.Chapter{}, &models.ReadingProgress{}, &models.Group{}, &models.CacheItem{}, &models.Bookmark{}, &models.Note{}, &models.Plugin{})
	log.Println("Database initialized in Silent Mode")
}
