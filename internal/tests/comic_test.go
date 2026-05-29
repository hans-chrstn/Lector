package tests

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/services"
)

func createMockCBZ(path string) error {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	f1, _ := w.Create("page1.jpg")
	f1.Write([]byte("fake image data 1"))

	f2, _ := w.Create("page10.jpg")
	f2.Write([]byte("fake image data 10"))

	f3, _ := w.Create("page2.jpg")
	f3.Write([]byte("fake image data 2"))

	w.Close()
	return os.WriteFile(path, buf.Bytes(), 0644)
}

func TestComicProcessing(t *testing.T) {
	db.InitDB(":memory:")
	services.EnsureUploadsDir()
	defer os.RemoveAll("uploads")

	t.Run("CBZ Ingestion and Natural Sorting", func(t *testing.T) {
		path := filepath.Join("uploads", "test.cbz")
		if err := createMockCBZ(path); err != nil {
			t.Fatalf("Failed to create mock CBZ: %v", err)
		}

		doc, err := services.ProcessLocalFile(path)
		if err != nil {
			t.Fatalf("ProcessLocalFile failed: %v", err)
		}

		if doc.Title != "test" {
			t.Errorf("Expected title 'test', got '%s'", doc.Title)
		}

		if len(doc.Chapters) != 1 {
			t.Fatalf("Expected 1 chapter, got %d", len(doc.Chapters))
		}

		content := doc.Chapters[0].Content
		idx1 := strings.Index(content, "page1.jpg")
		idx2 := strings.Index(content, "page2.jpg")
		idx10 := strings.Index(content, "page10.jpg")

		if idx1 == -1 || idx2 == -1 || idx10 == -1 {
			t.Errorf("Missing images in content")
		}

		if !(idx1 < idx2 && idx2 < idx10) {
			t.Errorf("Natural sorting failed")
		}
	})

	t.Run("Dynamic Image Retrieval", func(t *testing.T) {
		path := filepath.Join("uploads", "test.cbz")
		doc, _ := services.ProcessLocalFile(path)

		data, contentType, err := services.GetImageFromArchive(doc.ID, "page2.jpg")
		if err != nil {
			t.Fatalf("GetImageFromArchive failed: %v", err)
		}

		if string(data) != "fake image data 2" {
			t.Errorf("Incorrect image data retrieved")
		}

		if contentType != "image/jpeg" {
			t.Errorf("Expected image/jpeg, got %s", contentType)
		}
	})
}
