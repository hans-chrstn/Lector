package api

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"gorm.io/gorm"
)

func (h *API) DeleteHistory(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := db.DB.Where("document_id = ?", id).Delete(&models.ReadingProgress{}).Error; err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendString("Deleted")
}

func (h *API) ClearHistory(c *fiber.Ctx) error {
	if err := db.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.ReadingProgress{}).Error; err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendString("Cleared")
}

func (h *API) BatchDeleteHistory(c *fiber.Ctx) error {
	var req struct {
		IDs []uint `json:"ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	if err := db.DB.Where("document_id IN ?", req.IDs).Delete(&models.ReadingProgress{}).Error; err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendString("Batch Deleted")
}
