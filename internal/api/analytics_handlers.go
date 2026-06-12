package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	var readSeconds, chaptersRead, documentsRead int
	switch req.Type {
	case "time":
		readSeconds = req.Value
	case "chapter":
		chaptersRead = req.Value
	case "document":
		documentsRead = req.Value
	}

	err := db.DB.WithContext(c.UserContext()).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "date"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"read_seconds":   gorm.Expr("read_seconds + ?", readSeconds),
				"chapters_read":  gorm.Expr("chapters_read + ?", chaptersRead),
				"documents_read": gorm.Expr("documents_read + ?", documentsRead),
			}),
		}).
		Create(&models.ReadingStat{
			Date:          dateStr,
			ReadSeconds:   readSeconds,
			ChaptersRead:  chaptersRead,
			DocumentsRead: documentsRead,
		}).Error

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
