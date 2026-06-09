package interfaces

import (
	"github.com/user/lector/internal/models"
)

type PluginDataStore interface {
	UpdateDocumentMetadata(docID int, source string, global bool, updates map[string]interface{}) bool
	UpdateChapterMetadata(chapterID int, source string, global bool, metadata string) bool
	UpdateChapterContent(chapterID int, source string, global bool, content string) bool
	GetChapters(docID int, source string, global bool) ([]models.Chapter, bool)
	ListDocuments(source string, global bool) []models.Document
	GetReadingProgress(docID int) (*models.ReadingProgress, bool)
	SetReadingProgress(docID int, chapterID *uint, scrollPos *float64, clientUpdatedAt *int64) bool
	GetDocumentForExport(docID int) (*models.Document, bool)

	SetCacheItem(key string, value []byte) bool
	GetCacheItem(key string) ([]byte, bool)
}
