package repository

import (
	"github.com/user/lector/internal/core/interfaces"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"gorm.io/gorm"
)

type PluginRepository struct {
	DB *gorm.DB
}

func NewPluginRepository() interfaces.PluginDataStore {
	return &PluginRepository{DB: db.DB}
}

func (r *PluginRepository) UpdateDocumentMetadata(docID int, source string, global bool, updates map[string]interface{}) bool {
	var result *gorm.DB
	if global {
		result = r.DB.Model(&models.Document{}).Where("id = ?", docID).Updates(updates)
	} else {
		result = r.DB.Model(&models.Document{}).Where("id = ? AND source = ?", docID, source).Updates(updates)
	}
	return result.Error == nil
}

func (r *PluginRepository) UpdateChapterMetadata(chapterID int, source string, global bool, metadata string) bool {
	var chapter models.Chapter
	if err := r.DB.First(&chapter, chapterID).Error; err != nil {
		return false
	}
	var doc models.Document
	if err := r.DB.First(&doc, chapter.DocumentID).Error; err != nil {
		return false
	}
	if doc.Source != source && !global {
		return false
	}
	result := r.DB.Model(&models.Chapter{}).Where("id = ?", chapterID).Updates(map[string]interface{}{
		"metadata": metadata,
		"status":   "done",
	})
	return result.Error == nil
}

func (r *PluginRepository) UpdateChapterContent(chapterID int, source string, global bool, content string) bool {
	var chapter models.Chapter
	if err := r.DB.First(&chapter, chapterID).Error; err != nil {
		return false
	}
	var doc models.Document
	if err := r.DB.First(&doc, chapter.DocumentID).Error; err != nil {
		return false
	}
	if doc.Source != source && !global {
		return false
	}
	result := r.DB.Model(&models.Chapter{}).Where("id = ?", chapterID).Updates(map[string]interface{}{
		"content": content,
		"status":  "done",
	})
	return result.Error == nil
}

func (r *PluginRepository) GetChapters(docID int, source string, global bool) ([]models.Chapter, bool) {
	var doc models.Document
	if err := r.DB.First(&doc, docID).Error; err != nil || (doc.Source != source && !global) {
		return nil, false
	}
	var chapters []models.Chapter
	r.DB.Where("document_id = ?", docID).Order("`order` ASC").Find(&chapters)
	return chapters, true
}

func (r *PluginRepository) ListDocuments(source string, global bool) []models.Document {
	var documents []models.Document
	if global {
		r.DB.Find(&documents)
	} else {
		r.DB.Where("source = ?", source).Find(&documents)
	}
	return documents
}

func (r *PluginRepository) GetReadingProgress(docID int) (*models.ReadingProgress, bool) {
	var prog models.ReadingProgress
	if err := r.DB.Where("document_id = ?", docID).First(&prog).Error; err != nil {
		return nil, false
	}
	return &prog, true
}

func (r *PluginRepository) SetReadingProgress(docID int, chapterID *uint, scrollPos *float64, clientUpdatedAt *int64) bool {
	var existing models.ReadingProgress
	r.DB.Where("document_id = ?", docID).First(&existing)
	existing.DocumentID = uint(docID)

	if chapterID != nil {
		existing.ChapterID = *chapterID
	}
	if scrollPos != nil {
		existing.ScrollPos = *scrollPos
	}
	if clientUpdatedAt != nil {
		existing.ClientUpdatedAt = *clientUpdatedAt
	}

	if existing.ID == 0 {
		r.DB.Create(&existing)
	} else {
		r.DB.Save(&existing)
	}
	if existing.ChapterID != 0 {
		r.DB.Model(&models.Chapter{}).Where("id = ?", existing.ChapterID).Update("is_read", true)
	}
	return true
}

func (r *PluginRepository) GetDocumentForExport(docID int) (*models.Document, bool) {
	var doc models.Document
	if err := r.DB.Preload("Chapters", func(db *gorm.DB) *gorm.DB {
		return db.Order("CAST(\"order\" AS INTEGER) ASC")
	}).First(&doc, uint(docID)).Error; err != nil {
		return nil, false
	}
	return &doc, true
}

func (r *PluginRepository) SetCacheItem(key string, value []byte) bool {
	r.DB.Save(&models.CacheItem{
		Key:   key,
		Value: value,
	})
	return true
}

func (r *PluginRepository) GetCacheItem(key string) ([]byte, bool) {
	var item models.CacheItem
	if err := r.DB.Where("key = ?", key).First(&item).Error; err == nil {
		return item.Value, true
	}
	return nil, false
}
