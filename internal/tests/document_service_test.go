package tests

import (
	"os"
	"testing"

	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/services"
)

func TestDocumentService(t *testing.T) {
	testDB := "test_document_service.db"
	defer os.Remove(testDB)
	db.InitDB(testDB)

	docService := services.NewDocumentService()

	t.Run("Ensure Document - New", func(t *testing.T) {
		doc := &models.Document{
			Title:       "Test Doc",
			URL:         "https://test.com/doc1",
			Source:      "test",
			IsInLibrary: true,
		}
		err := docService.Save(doc)
		if err != nil {
			t.Fatalf("Failed to save document: %v", err)
		}

		found, err := docService.GetByID(doc.ID)
		if err != nil {
			t.Fatalf("Failed to find document: %v", err)
		}
		if found.Title != "Test Doc" {
			t.Errorf("Expected 'Test Doc', got '%s'", found.Title)
		}
	})

	t.Run("Get All Active", func(t *testing.T) {
		docs, err := docService.GetAllInLibrary(false)
		if err != nil {
			t.Fatalf("Failed to get documents: %v", err)
		}
		if len(docs) == 0 {
			t.Errorf("Expected at least 1 document")
		}
	})
}
