package api

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/services"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (h *API) GetActivePluginNames() []string {
	var n []string
	for name, p := range h.Plugins {
		if p.HasCapability("catalog") {
			n = append(n, name)
		}
	}
	return n
}

func (h *API) GetDocuments(c *fiber.Ctx) error {
	showArchived := c.Query("archived") == "true"
	var docs []models.Document
	err := db.DB.Where("is_in_library = ? AND is_archived = ?", true, showArchived).Find(&docs).Error
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	for i := range docs {
		var readCount int64
		db.DB.Model(&models.Chapter{}).Where("document_id = ? AND is_read = ?", docs[i].ID, true).Count(&readCount)
		docs[i].ReadChapters = int(readCount)

		var totalCount int64
		db.DB.Model(&models.Chapter{}).Where("document_id = ?", docs[i].ID).Count(&totalCount)
		docs[i].TotalChapters = int(totalCount)
	}
	return c.JSON(docs)
}

func (h *API) EnsureDocument(c *fiber.Ctx) error {
	var req struct {
		URL    string `json:"url"`
		Source string `json:"source"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid request body")
	}

	url, plugin := req.URL, req.Source
	doc, _ := h.DocumentService.GetByURL(url)

	if doc == nil {
		if plugin == "local" {
			return c.Status(404).SendString("Local document not found")
		}

		s, ok := h.Plugins[plugin]
		if !ok {
			return c.Status(400).SendString("Invalid plugin")
		}

		fetched, err := s.GetDocument(url)
		if err != nil || fetched.Title == "" {
			return c.Status(500).SendString("Fetch failed")
		}
		fetched.Source = plugin
		chapters := fetched.Chapters
		fetched.Chapters = nil
		if err := db.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(&fetched).Error; err != nil {
			return c.Status(500).SendString("Failed to save document")
		}
		doc = &fetched
		for i := range chapters {
			chapters[i].DocumentID = doc.ID
			chapters[i].ID = 0
			chapters[i].Order = i + 1
		}
		if err := db.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "document_id"}, {Name: "url"}},
			DoUpdates: clause.AssignmentColumns([]string{"title", "order"}),
		}).CreateInBatches(chapters, 100).Error; err != nil {
			fmt.Printf("[API] Error creating chapters for %s: %v\n", doc.Title, err)
		}
	} else {
		var count int64
		db.DB.Model(&models.Chapter{}).Where("document_id = ?", doc.ID).Count(&count)

		if count == 0 {
			if plugin == "local" {
				services.ProcessLocalFile(doc.LocalPath)
			} else {
				s, ok := h.Plugins[plugin]
				if ok {
					fetched, err := s.GetDocument(url)
					if err != nil {
						fmt.Printf("[API] Error refetching chapters for %s: %v\n", doc.Title, err)
					}

					doc.CoverURL = fetched.CoverURL
					doc.Author = fetched.Author
					doc.Synopsis = fetched.Synopsis
					db.DB.Save(doc)

					for i := range fetched.Chapters {
						fetched.Chapters[i].DocumentID = doc.ID
						fetched.Chapters[i].ID = 0
						fetched.Chapters[i].Order = i + 1
					}
					if err := db.DB.Clauses(clause.OnConflict{
						Columns:   []clause.Column{{Name: "document_id"}, {Name: "url"}},
						DoUpdates: clause.AssignmentColumns([]string{"title", "order"}),
					}).CreateInBatches(fetched.Chapters, 100).Error; err != nil {
						fmt.Printf("[API] Error creating chapters for existing %s: %v\n", doc.Title, err)
					}
				}
			}
		}
	}

	db.DB.Preload("Chapters", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "document_id", "title", "url", "order", "is_read", "status").Order("CAST(\"order\" AS INTEGER) ASC")
	}).First(doc, doc.ID)

	var readCount int64
	db.DB.Model(&models.Chapter{}).Where("document_id = ? AND is_read = ?", doc.ID, true).Count(&readCount)
	doc.ReadChapters = int(readCount)
	doc.TotalChapters = len(doc.Chapters)

	return c.JSON(doc)
}

func (h *API) GetDocumentByID(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	doc, err := h.DocumentService.GetByID(uint(id))
	if err != nil {
		return c.Status(404).SendString("Document not found")
	}

	db.DB.Preload("Chapters", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "document_id", "title", "url", "order", "is_read", "status").Order("CAST(\"order\" AS INTEGER) ASC")
	}).First(doc, doc.ID)

	var readCount int64
	db.DB.Model(&models.Chapter{}).Where("document_id = ? AND is_read = ?", doc.ID, true).Count(&readCount)
	doc.ReadChapters = int(readCount)
	doc.TotalChapters = len(doc.Chapters)

	return c.JSON(doc)
}

func (h *API) ToggleLibrary(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	inLibrary := c.Query("is_in_library") == "true"
	groupID, _ := strconv.Atoi(c.Query("group_id", "0"))

	h.DocumentService.ToggleLibrary(uint(id), inLibrary, uint(groupID))
	return c.SendString("Updated")
}

func (h *API) GetDocumentProgress(c *fiber.Ctx) error {
	var p models.ReadingProgress
	db.DB.Where("document_id = ?", c.Params("id")).First(&p)
	return c.JSON(p)
}

func (h *API) GetHistory(c *fiber.Ctx) error {
	var docs []models.Document
	err := db.DB.Table("documents").
		Select("documents.*").
		Joins("JOIN reading_progresses ON reading_progresses.document_id = documents.id").
		Where("documents.is_archived = ?", false).
		Order("reading_progresses.updated_at DESC").
		Limit(50).
		Scan(&docs).Error

	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	for i := range docs {
		var count int64
		db.DB.Model(&models.Chapter{}).Where("document_id = ? AND is_read = ?", docs[i].ID, true).Count(&count)
		docs[i].ReadChapters = int(count)

		var totalCount int64
		db.DB.Model(&models.Chapter{}).Where("document_id = ?", docs[i].ID).Count(&totalCount)
		docs[i].TotalChapters = int(totalCount)
	}

	return c.JSON(docs)
}

func (h *API) UpdateMetadata(c *fiber.Ctx) error {
	var req map[string]interface{}
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Params("id"))
	h.DocumentService.UpdateMetadata(uint(id), req)
	return c.SendString("Updated")
}

func (h *API) UpdateCover(c *fiber.Ctx) error {
	file, err := c.FormFile("cover")
	if err != nil {
		return c.Status(400).SendString("No file uploaded")
	}

	ext := filepath.Ext(file.Filename)
	localCoverName := fmt.Sprintf("custom_%s%s", c.Params("id"), ext)
	localCoverPath := filepath.Join("uploads", localCoverName)

	if err := c.SaveFile(file, localCoverPath); err != nil {
		return c.Status(500).SendString("Failed to save cover")
	}

	db.DB.Model(&models.Document{}).Where("id = ?", c.Params("id")).Update("cover_url", "/uploads/"+localCoverName)
	return c.JSON(fiber.Map{"url": "/uploads/" + localCoverName})
}

func (h *API) MigrateDocument(c *fiber.Ctx) error {
	var req struct {
		URL    string `json:"url"`
		Source string `json:"source"`
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Params("id"))
	doc, _ := h.DocumentService.GetByID(uint(id))
	doc.URL = req.URL
	doc.Source = req.Source
	h.DocumentService.Save(doc)
	return c.SendString("Migrated")
}

func (h *API) GetArchiveImage(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	fileName := c.Query("file")
	if fileName == "" {
		return c.Status(400).SendString("Missing file parameter")
	}

	data, contentType, err := services.GetImageFromArchive(uint(id), fileName)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	if contentType == "" {
		contentType = "image/jpeg"
	}

	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "public, max-age=604800")
	return c.Send(data)
}
