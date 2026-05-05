package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
)

func (h *API) GetChapterByID(c *fiber.Ctx) error {
	var ch models.Chapter
	db.DB.Where("id = ?", c.Params("id")).Find(&ch)
	if ch.ID == 0 {
		return c.Status(404).SendString("Chapter not found")
	}
	if ch.Content == "" {
		var document models.Document
		db.DB.Find(&document, ch.DocumentID)
		s := h.Plugins[document.Source]
		res, _ := s.GetChapter(ch.URL)
		ch.Content = s.CleanHTML(res.Content, ch.Title)
		ch.Status = "done"
		db.DB.Save(&ch)
	}
	return c.JSON(ch)
}

func (h *API) ToggleChapterRead(c *fiber.Ctx) error {
	db.DB.Model(&models.Chapter{}).Where("id = ?", c.Params("id")).Update("is_read", c.Query("read") == "true")
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
	db.DB.Model(&models.Chapter{}).Where("id IN ?", req.IDs).Update("is_read", req.IsRead)
	return c.SendString("Updated")
}

func (h *API) SyncProgress(c *fiber.Ctx) error {
	var p models.ReadingProgress
	if err := c.BodyParser(&p); err != nil {
		return err
	}

	var existing models.ReadingProgress
	db.DB.Where("document_id = ?", p.DocumentID).First(&existing)

	if existing.ID != 0 {
		updates := map[string]interface{}{
			"updated_at": time.Now(),
		}
		if p.ClientUpdatedAt >= existing.ClientUpdatedAt {
			updates["chapter_id"] = p.ChapterID
			updates["scroll_pos"] = p.ScrollPos
			updates["client_updated_at"] = p.ClientUpdatedAt
		}
		db.DB.Model(&existing).Updates(updates)
	} else {
		p.UpdatedAt = time.Now()
		db.DB.Create(&p)
	}

	db.DB.Model(&models.Chapter{}).Where("id = ?", p.ChapterID).Update("is_read", true)
	return c.SendString("Synced")
}
