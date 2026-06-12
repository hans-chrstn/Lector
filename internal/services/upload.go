package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
)

func ProcessLocalFile(path string) (*models.Document, error) {
	ext := strings.ToLower(filepath.Ext(path))
	var doc *models.Document
	var err error

	if ext == ".epub" {
		doc, err = processEPUB(path)
	} else if ext == ".cbz" || ext == ".cbr" {
		doc, err = processComic(path)
	} else if ext == ".pdf" {
		doc, err = processPDF(path)
	} else {
		return nil, fmt.Errorf("unsupported format: %s", ext)
	}

	if err != nil {
		return nil, err
	}

	applySidecarMetadata(doc, path)

	return doc, nil
}

func processPDF(path string) (*models.Document, error) {
	docURL := "local://" + filepath.Base(path)
	var document models.Document
	db.DB.Where("url = ?", docURL).First(&document)
	if document.ID == 0 {
		document = models.Document{URL: docURL}
	}

	document.Title = strings.TrimSuffix(filepath.Base(path), ".pdf")
	document.Source = "local"
	document.IsLocal = true
	document.LocalPath = path
	document.IsInLibrary = true
	document.Type = "text"

	if err := db.DB.Save(&document).Error; err != nil {
		return nil, err
	}

	proxyURL := fmt.Sprintf("/api/proxy-image?url=/%s", filepath.ToSlash(path))
	content := fmt.Sprintf(`<iframe src="%s" style="width: 100%%; height: 90vh; border: none;"></iframe>`, proxyURL)

	chapter := models.Chapter{
		DocumentID: document.ID,
		Title:      "Read PDF",
		URL:        document.URL + "/read",
		Content:    content,
		Order:      1,
		Status:     "done",
	}

	db.DB.Unscoped().Where("document_id = ?", document.ID).Delete(&models.Chapter{})
	db.DB.Create(&chapter)

	document.Chapters = []models.Chapter{chapter}
	return &document, nil
}

func applySidecarMetadata(doc *models.Document, bookPath string) {
	opfPath := strings.TrimSuffix(bookPath, filepath.Ext(bookPath)) + ".opf"
	if _, err := os.Stat(opfPath); err == nil {
		if meta, err := ParseSidecarOPF(opfPath); err == nil {
			if meta.Title != "" {
				doc.Title = meta.Title
			}
			if meta.Author != "" {
				doc.Author = meta.Author
			}
			if meta.Synopsis != "" {
				doc.Synopsis = meta.Synopsis
			}
			if meta.Genres != "" {
				doc.Genres = meta.Genres
			}
			if meta.Status != "" {
				doc.Status = meta.Status
			}
			db.DB.Save(doc)
		}
	}
}

func EnsureUploadsDir() {
	os.MkdirAll("uploads", 0755)
}
