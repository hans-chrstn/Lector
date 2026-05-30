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
	} else {
		return nil, fmt.Errorf("unsupported format: %s", ext)
	}

	if err != nil {
		return nil, err
	}

	applySidecarMetadata(doc, path)

	return doc, nil
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

func stringBetween(str, start, end string) string {
	s := strings.Index(str, start)
	if s == -1 {
		return ""
	}
	s += len(start)
	e := strings.Index(str[s:], end)
	if e == -1 {
		return ""
	}
	return str[s : s+e]
}

func EnsureUploadsDir() {
	os.MkdirAll("uploads", 0755)
}
