package services

import (
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/repository"
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

type documentService struct {
	docRepo     repository.Repository[models.Document]
	chapterRepo repository.Repository[models.Chapter]
}

func NewDocumentService(docRepo repository.Repository[models.Document], chapterRepo repository.Repository[models.Chapter]) DocumentService {
	return &documentService{
		docRepo:     docRepo,
		chapterRepo: chapterRepo,
	}
}

func (s *documentService) GetByID(id uint) (*models.Document, error) {
	return s.docRepo.FindByID(id)
}

func (s *documentService) GetByURL(url string) (*models.Document, error) {
	all, err := s.docRepo.FindAll()
	if err != nil {
		return nil, err
	}
	for _, doc := range all {
		if doc.URL == url {
			return &doc, nil
		}
	}
	return nil, nil
}

func (s *documentService) GetAllInLibrary(archived bool) ([]models.Document, error) {
	all, err := s.docRepo.FindAll()
	if err != nil {
		return nil, err
	}

	var filtered []models.Document
	for _, doc := range all {
		if doc.IsInLibrary && doc.IsArchived == archived {
			filtered = append(filtered, doc)
		}
	}
	return filtered, nil
}

func (s *documentService) Save(doc *models.Document) error {
	if doc.ID == 0 {
		return s.docRepo.Create(doc)
	}
	return s.docRepo.Update(doc)
}

func (s *documentService) DeleteBatch(ids []uint) error {
	for _, id := range ids {
		if err := s.docRepo.Delete(id); err != nil {
			return err
		}
	}
	return nil
}

func (s *documentService) UpdateMetadata(id uint, metadata map[string]interface{}) error {
	doc, err := s.docRepo.FindByID(id)
	if err != nil {
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

	return s.docRepo.Update(doc)
}

func (s *documentService) ToggleLibrary(id uint, inLibrary bool, groupID uint) error {
	doc, err := s.docRepo.FindByID(id)
	if err != nil {
		return err
	}
	doc.IsInLibrary = inLibrary
	doc.GroupID = groupID
	return s.docRepo.Update(doc)
}

func (s *documentService) MoveBatch(ids []uint, groupID uint) error {
	for _, id := range ids {
		doc, err := s.docRepo.FindByID(id)
		if err == nil {
			doc.GroupID = groupID
			s.docRepo.Update(doc)
		}
	}
	return nil
}

func (s *documentService) ArchiveBatch(ids []uint, archive bool) error {
	for _, id := range ids {
		doc, err := s.docRepo.FindByID(id)
		if err == nil {
			doc.IsArchived = archive
			s.docRepo.Update(doc)
		}
	}
	return nil
}

func (s *documentService) MarkReadBatch(ids []uint, isRead bool) error {
	for _, id := range ids {
		doc, err := s.docRepo.FindByID(id)
		if err == nil {
			for i := range doc.Chapters {
				doc.Chapters[i].IsRead = isRead
			}
			s.docRepo.Update(doc)
		}
	}
	return nil
}
