package tests

import (
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"os"
	"testing"
)

func TestDatabaseModels(t *testing.T) {
	testDB := "test_models.db"
	defer os.Remove(testDB)
	db.InitDB(testDB)

	document := models.Document{
		Title:  "Test Document",
		URL:    "https://test.com",
		Source: "test",
		Chapters: []models.Chapter{
			{Title: "Chapter 1", URL: "https://test.com/1", Order: 1},
			{Title: "Chapter 2", URL: "https://test.com/2", Order: 2},
		},
	}

	if err := db.DB.Create(&document).Error; err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	progress := models.ReadingProgress{
		DocumentID: document.ID,
		ChapterID:  document.Chapters[0].ID,
		ScrollPos:  0.5,
	}
	if err := db.DB.Create(&progress).Error; err != nil {
		t.Fatalf("Failed to create progress: %v", err)
	}

	var saved models.Document
	db.DB.Preload("Chapters").First(&saved, document.ID)
	if saved.Title != "Test Document" {
		t.Errorf("Expected 'Test Document', got '%s'", saved.Title)
	}
	if len(saved.Chapters) != 2 {
		t.Errorf("Expected 2 chapters, got %d", len(saved.Chapters))
	}

	db.DB.Model(&models.Chapter{}).Where("id = ?", document.Chapters[0].ID).Update("is_read", true)

	var count int64
	db.DB.Model(&models.Chapter{}).Where("document_id = ? AND is_read = ?", document.ID, true).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 read chapter, got %d", count)
	}
}
