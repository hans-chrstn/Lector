package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/binder"
	"github.com/user/lector/internal/db"
)

func (h *API) ExportDocument(c *fiber.Ctx) error {
	id := c.Params("id")
	format := c.Query("format", "epub")
	docID, _ := strconv.Atoi(id)
	doc, err := h.DocumentService.GetByID(uint(docID))
	if err != nil {
		return c.Status(404).SendString("Document not found")
	}

	db.DB.Model(doc).Association("Chapters").Find(&doc.Chapters)

	ext := "epub"
	if format == "pdf" {
		ext = "pdf"
	}

	os.MkdirAll("exports", 0755)
	path := filepath.Join("exports", fmt.Sprintf("%s.%s", doc.Title, ext))

	var binderErr error
	if format == "pdf" {
		binderErr = binder.BindPDF(doc, path)
	} else {
		binderErr = binder.BindEPUB(doc, path)
	}

	if binderErr != nil {
		return c.Status(500).SendString(fmt.Sprintf("Failed to bind %s: %v", format, binderErr))
	}

	go func() {
		time.Sleep(1 * time.Minute)
		os.Remove(path)
	}()

	return c.Download(path)
}
