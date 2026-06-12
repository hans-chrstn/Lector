package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/microcosm-cc/bluemonday"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/services"
)

func (h *API) GetChapterByID(c *fiber.Ctx) error {
	var ch models.Chapter
	db.DB.WithContext(c.UserContext()).Where("id = ?", c.Params("id")).Find(&ch)
	if ch.ID == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Chapter not found"})
	}

	if ch.Content == "" && (ch.Metadata == "" || ch.Metadata == "[]" || ch.Metadata == "null") {
		var document models.Document
		db.DB.WithContext(c.UserContext()).Find(&document, ch.DocumentID)

		if document.Source == "local" {
			content, err := services.ExtractLocalChapter(document.URL, ch.URL, ch.Title)
			if err == nil {
				ch.Content = content
			}
		} else {
			s, ok := h.Engine.Plugins[document.Source]
			if ok {
				res, err := s.GetChapter(ch.URL)
				if err == nil {
					ch.Content = s.CleanHTML(res.Content, ch.Title)
					ch.Metadata = res.Metadata
					ch.Status = "done"
					db.DB.WithContext(c.UserContext()).Save(&ch)
				}
			}
		}
	}

	p := bluemonday.UGCPolicy()
	p.AllowAttrs("style").OnElements("span", "div", "p")
	ch.Content = p.Sanitize(ch.Content)

	return c.JSON(ch)
}

func (h *API) ToggleChapterRead(c *fiber.Ctx) error {
	db.DB.WithContext(c.UserContext()).Model(&models.Chapter{}).Where("id = ?", c.Params("id")).Update("is_read", c.Query("read") == "true")
	return c.SendString("Updated")
}

func (h *API) BatchUpdateChapters(c *fiber.Ctx) error {
	var req struct {
		IDs    []uint `json:"ids"`
		IsRead bool   `json:"is_read"`
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	db.DB.WithContext(c.UserContext()).Model(&models.Chapter{}).Where("id IN ?", req.IDs).Update("is_read", req.IsRead)
	return c.SendString("Updated")
}

func (h *API) SyncProgress(c *fiber.Ctx) error {
	var p models.ReadingProgress
	if err := c.BodyParser(&p); err != nil {
		return err
	}

	var existing models.ReadingProgress
	db.DB.WithContext(c.UserContext()).Where("document_id = ?", p.DocumentID).First(&existing)

	if existing.ID != 0 {
		updates := map[string]interface{}{
			"updated_at": time.Now(),
		}
		if p.ClientUpdatedAt >= existing.ClientUpdatedAt {
			updates["chapter_id"] = p.ChapterID
			updates["scroll_pos"] = p.ScrollPos
			updates["client_updated_at"] = p.ClientUpdatedAt
		}
		db.DB.WithContext(c.UserContext()).Model(&existing).Updates(updates)
	} else {
		p.UpdatedAt = time.Now()
		db.DB.WithContext(c.UserContext()).Create(&p)
	}

	db.DB.WithContext(c.UserContext()).Model(&models.Chapter{}).Where("id = ?", p.ChapterID).Update("is_read", true)
	return c.SendString("Synced")
}
