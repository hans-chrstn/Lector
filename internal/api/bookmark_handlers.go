package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
)

func (h *API) AddBookmark(c *fiber.Ctx) error {
	var b models.Bookmark
	if err := c.BodyParser(&b); err != nil {
		return err
	}
	db.DB.WithContext(c.UserContext()).Create(&b)
	return c.JSON(b)
}

func (h *API) GetBookmarks(c *fiber.Ctx) error {
	var b []models.Bookmark
	db.DB.WithContext(c.UserContext()).Where("document_id = ?", c.Params("documentId")).Order("created_at DESC").Find(&b)
	return c.JSON(b)
}

func (h *API) DeleteBookmark(c *fiber.Ctx) error {
	db.DB.WithContext(c.UserContext()).Delete(&models.Bookmark{}, c.Params("id"))
	return c.SendString("Deleted")
}
