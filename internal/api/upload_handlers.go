package api

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/user/lector/internal/services"
)

func (h *API) HandleUpload(c *fiber.Ctx) error {
	file, err := c.FormFile("book")
	if err != nil {
		log.Printf("[API] Upload error: %v", err)
		return c.Status(400).SendString("No file uploaded")
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(500).SendString("Failed to open uploaded file")
	}
	defer src.Close()

	head := make([]byte, 512)
	if _, err := src.Read(head); err != nil {
		return c.Status(500).SendString("Failed to read file header")
	}
	if _, err := src.Seek(0, 0); err != nil {
		return c.Status(500).SendString("Failed to reset file pointer")
	}

	contentType := http.DetectContentType(head)
	isPDF := contentType == "application/pdf"
	isZIP := contentType == "application/zip" || contentType == "application/x-zip-compressed"
	isRAR := false

	if !isPDF && !isZIP {
		if bytes.HasPrefix(head, []byte("Rar!\x1a\x07\x00")) || bytes.HasPrefix(head, []byte("Rar!\x1a\x07\x01\x00")) {
			isRAR = true
		}
	}

	if !isPDF && !isZIP && !isRAR {
		log.Printf("[API] Blocked invalid upload type: %s", contentType)
		return c.Status(400).SendString("Only EPUB, PDF, CBZ, and CBR files are allowed")
	}

	newID := uuid.New().String()
	ext := filepath.Ext(file.Filename)
	if isPDF && !strings.EqualFold(ext, ".pdf") {
		ext = ".pdf"
	} else if isZIP && !strings.EqualFold(ext, ".epub") && !strings.EqualFold(ext, ".cbz") {
		ext = ".epub"
	} else if isRAR && !strings.EqualFold(ext, ".cbr") {
		ext = ".cbr"
	}

	fileName := fmt.Sprintf("%s%s", newID, ext)
	path := filepath.Join("uploads", fileName)

	log.Printf("[API] Saving upload as: %s (Detected: %s)", path, contentType)
	if err := c.SaveFile(file, path); err != nil {
		log.Printf("[API] Save error: %v", err)
		return c.Status(500).SendString("Failed to save file")
	}

	document, err := services.ProcessLocalFile(path)
	if err != nil {
		log.Printf("[API] Process error: %v", err)
		return c.Status(500).SendString(err.Error())
	}

	log.Printf("[API] Successfully processed book: %s (ID: %d)", document.Title, document.ID)
	return c.JSON(document)
}
