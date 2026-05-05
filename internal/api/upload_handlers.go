package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/services"
	"log"
	"path/filepath"
)

func (h *API) HandleUpload(c *fiber.Ctx) error {
	file, err := c.FormFile("book")
	if err != nil {
		log.Printf("[API] Upload error: %v", err)
		return c.Status(400).SendString("No file uploaded")
	}

	path := filepath.Join("uploads", file.Filename)
	log.Printf("[API] Saving uploaded file to: %s", path)
	if err := c.SaveFile(file, path); err != nil {
		log.Printf("[API] Save error: %v", err)
		return c.Status(500).SendString("Failed to save file")
	}

	document, err := services.ProcessLocalFile(path)
	if err != nil {
		log.Printf("[API] Process error: %v", err)
		return c.Status(500).SendString(err.Error())
	}

	log.Printf("[API] Successfully processed local book: %s (ID: %d)", document.Title, document.ID)
	return c.JSON(document)
}
