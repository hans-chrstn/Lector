package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
)

func (h *API) AddNote(c *fiber.Ctx) error {
	var n models.Note
	if err := c.BodyParser(&n); err != nil {
		return err
	}
	db.DB.Create(&n)
	return c.JSON(n)
}

func (h *API) GetNotes(c *fiber.Ctx) error {
	var n []models.Note
	db.DB.Where("document_id = ?", c.Params("documentId")).Order("created_at DESC").Find(&n)
	return c.JSON(n)
}

func (h *API) DeleteNote(c *fiber.Ctx) error {
	db.DB.Delete(&models.Note{}, c.Params("id"))
	return c.SendString("Deleted")
}
