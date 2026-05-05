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

func (h *API) GetActivePlugins(c *fiber.Ctx) error {
	var n []string
	for name := range h.Plugins {
		n = append(n, name)
	}
	return c.JSON(n)
}

func (h *API) GetDocuments(c *fiber.Ctx) error {
	showArchived := c.Query("archived") == "true"
	docs, err := h.DocumentService.GetAllInLibrary(showArchived)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	for i := range docs {
		var count int64
		db.DB.Model(&models.Chapter{}).Where("document_id = ? AND is_read = ?", docs[i].ID, true).Count(&count)
		docs[i].ReadChapters = int(count)
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
		if err := db.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&fetched).Error; err != nil {
			return c.Status(500).SendString("Failed to save document")
		}
		doc = &fetched
		for i := range chapters {
			chapters[i].DocumentID = doc.ID
			chapters[i].ID = 0
			chapters[i].Order = i + 1
		}
		db.DB.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(chapters, 250)
	} else {
		var count int64
		db.DB.Model(&models.Chapter{}).Where("document_id = ?", doc.ID).Count(&count)

		if count == 0 {
			if plugin == "local" {
				services.ProcessLocalFile(doc.LocalPath)
			} else {
				s, ok := h.Plugins[plugin]
				if ok {
					fetched, _ := s.GetDocument(url)
					for i := range fetched.Chapters {
						fetched.Chapters[i].DocumentID = doc.ID
						fetched.Chapters[i].ID = 0
						fetched.Chapters[i].Order = i + 1
					}
					db.DB.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(fetched.Chapters, 250)
				}
			}
		}
	}

	db.DB.Preload("Chapters", func(db *gorm.DB) *gorm.DB {
		return db.Order("`order` ASC")
	}).First(doc, doc.ID)

	var count int64
	db.DB.Model(&models.Chapter{}).Where("document_id = ? AND is_read = ?", doc.ID, true).Count(&count)
	doc.ReadChapters = int(count)

	return c.JSON(doc)
}

func (h *API) GetDocumentByID(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	doc, err := h.DocumentService.GetByID(uint(id))
	if err != nil {
		return c.Status(404).SendString("Document not found")
	}

	db.DB.Model(doc).Association("Chapters").Find(&doc.Chapters)

	var count int64
	db.DB.Model(&models.Chapter{}).Where("document_id = ? AND is_read = ?", doc.ID, true).Count(&count)
	doc.ReadChapters = int(count)

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
		db.DB.Model(&docs[i]).Association("Chapters").Find(&docs[i].Chapters)
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
