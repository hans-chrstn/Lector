package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/binder"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"gorm.io/gorm"
)

func sanitizeFilename(name string) string {
	badChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	res := name
	for _, c := range badChars {
		res = strings.ReplaceAll(res, c, "_")
	}
	return res
}

func (h *API) ExportDocument(c *fiber.Ctx) error {
	id := c.Params("id")
	format := c.Query("format", "epub")
	docID, _ := strconv.Atoi(id)

	var doc models.Document
	if err := db.DB.WithContext(c.UserContext()).Preload("Chapters", func(db *gorm.DB) *gorm.DB {
		return db.Order("order_val ASC")
	}).First(&doc, uint(docID)).Error; err != nil {
		fmt.Printf("[Export] Document %d not found: %v\n", docID, err)
		return c.Status(404).JSON(fiber.Map{"error": "Document not found"})
	}

	ext := "epub"
	if format == "pdf" {
		ext = "pdf"
	} else if format == "cbz" {
		ext = "cbz"
	}

	os.MkdirAll("exports", 0755)
	safeTitle := sanitizeFilename(doc.Title)
	path := filepath.Join("exports", fmt.Sprintf("%s.%s", safeTitle, ext))

	fmt.Printf("[Export] Binding %s to %s (Format: %s, Chapters: %d)\n", doc.Title, path, format, len(doc.Chapters))

	var binderErr error
	if format == "pdf" {
		binderErr = binder.BindPDF(&doc, path)
	} else if format == "cbz" {
		binderErr = binder.BindCBZ(&doc, path)
	} else {
		binderErr = binder.BindEPUB(&doc, path)
	}

	if binderErr != nil {
		fmt.Printf("[Export] Binder error for %s: %v\n", doc.Title, binderErr)
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to bind %s: %v", format, binderErr)})
	}

	go func() {
		time.Sleep(5 * time.Minute)
		os.Remove(path)
	}()

	return c.Download(path)
}
