package services

import (
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
)

type DocumentService interface {
	GetByID(id uint) (*models.Document, error)
	GetByURL(url string) (*models.Document, error)
	GetAllInLibrary(archived bool) ([]models.Document, error)
	Save(doc *models.Document) error
	DeleteBatch(ids []uint) error
	UpdateMetadata(id uint, metadata map[string]interface{}) error
	ToggleLibrary(id uint, inLibrary bool, groupID uint) error
	MoveBatch(ids []uint, groupID uint) error
	ArchiveBatch(ids []uint, archive bool) error
	MarkReadBatch(ids []uint, isRead bool) error
}

type documentService struct{}

func NewDocumentService() DocumentService {
	return &documentService{}
}

func (s *documentService) GetByID(id uint) (*models.Document, error) {
	var doc models.Document
	err := db.DB.First(&doc, id).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (s *documentService) GetByURL(url string) (*models.Document, error) {
	var doc models.Document
	err := db.DB.Where("url = ?", url).First(&doc).Error
	if err != nil {
		return nil, nil
	}
	return &doc, nil
}

func (s *documentService) GetAllInLibrary(archived bool) ([]models.Document, error) {
	var filtered []models.Document
	err := db.DB.Where("is_in_library = ? AND is_archived = ?", true, archived).Find(&filtered).Error
	return filtered, err
}

func (s *documentService) Save(doc *models.Document) error {
	if doc.ID == 0 {
		return db.DB.Create(doc).Error
	}
	return db.DB.Save(doc).Error
}

func (s *documentService) DeleteBatch(ids []uint) error {
	return db.DB.Where("id IN ?", ids).Delete(&models.Document{}).Error
}

func (s *documentService) UpdateMetadata(id uint, metadata map[string]interface{}) error {
	var doc models.Document
	if err := db.DB.First(&doc, id).Error; err != nil {
		return err
	}
	if title, ok := metadata["title"].(string); ok {
		doc.Title = title
	}
	if author, ok := metadata["author"].(string); ok {
		doc.Author = author
	}
	if synopsis, ok := metadata["synopsis"].(string); ok {
		doc.Synopsis = synopsis
	}
	if genres, ok := metadata["genres"].(string); ok {
		doc.Genres = genres
	}
	if status, ok := metadata["status"].(string); ok {
		doc.Status = status
	}
	return db.DB.Save(&doc).Error
}

func (s *documentService) ToggleLibrary(id uint, inLibrary bool, groupID uint) error {
	return db.DB.Model(&models.Document{}).Where("id = ?", id).Updates(map[string]interface{}{"is_in_library": inLibrary, "group_id": groupID}).Error
}

func (s *documentService) MoveBatch(ids []uint, groupID uint) error {
	return db.DB.Model(&models.Document{}).Where("id IN ?", ids).Update("group_id", groupID).Error
}

func (s *documentService) ArchiveBatch(ids []uint, archive bool) error {
	return db.DB.Model(&models.Document{}).Where("id IN ?", ids).Update("is_archived", archive).Error
}

func (s *documentService) MarkReadBatch(ids []uint, isRead bool) error {
	return db.DB.Model(&models.Chapter{}).Where("document_id IN ?", ids).Update("is_read", isRead).Error
}
