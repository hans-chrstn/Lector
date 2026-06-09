package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"gorm.io/gorm"
)

type AnalyticsTrackRequest struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

func (h *API) TrackAnalytics(c *fiber.Ctx) error {
	var req AnalyticsTrackRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	dateStr := time.Now().Format("2006-01-02")
	stat := models.ReadingStat{
		Date: dateStr,
	}

	err := db.DB.WithContext(c.UserContext()).Transaction(func(tx *gorm.DB) error {
		var existing models.ReadingStat
		if err := tx.Where("date = ?", dateStr).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := tx.Create(&stat).Error; err != nil {
					return err
				}
				existing = stat
			} else {
				return err
			}
		}

		updates := map[string]interface{}{}
		switch req.Type {
		case "time":
			updates["read_seconds"] = existing.ReadSeconds + req.Value
		case "chapter":
			updates["chapters_read"] = existing.ChaptersRead + req.Value
		case "document":
			updates["documents_read"] = existing.DocumentsRead + req.Value
		}

		if len(updates) > 0 {
			return tx.Model(&existing).Updates(updates).Error
		}
		return nil
	})

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to track analytics"})
	}

	return c.JSON(fiber.Map{"status": "success"})
}

func (h *API) GetAnalytics(c *fiber.Ctx) error {
	var stats []models.ReadingStat
	if err := db.DB.WithContext(c.UserContext()).Order("date DESC").Limit(30).Find(&stats).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch analytics"})
	}
	return c.JSON(stats)
}
