package tests

import (
	"archive/zip"
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/api"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/plugin"
	"github.com/user/lector/internal/services"
)

func createMaliciousEPUB() ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	f, _ := w.Create("../../../etc/passwd")
	f.Write([]byte("malicious content"))

	w.Close()
	return buf.Bytes(), nil
}

func TestUploadSecurityHardening(t *testing.T) {
	app := fiber.New()
	db.InitDB(":memory:")
	services.EnsureUploadsDir()
	defer os.RemoveAll("uploads")

	plugins := make(map[string]*plugin.LuaPlugin)
	api.RegisterRoutes(app, plugins)

	t.Run("UUID Renaming", func(t *testing.T) {
		epubData, _ := createMockEPUB()
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("book", "my_secret_novel.epub")
		part.Write(epubData)
		writer.Close()

		req := httptest.NewRequest("POST", "/api/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := app.Test(req)
		if err != nil || resp.StatusCode != 200 {
			t.Fatalf("Upload failed: %v, status: %d", err, resp.StatusCode)
		}

		files, _ := os.ReadDir("uploads")
		foundUUID := false
		uuidRegex := regexp.MustCompile(`^[a-fA-F0-9-]{36}\.epub$`)

		for _, f := range files {
			if uuidRegex.MatchString(f.Name()) {
				foundUUID = true
				break
			}
			if f.Name() == "my_secret_novel.epub" {
				t.Errorf("Original filename was preserved, security failure")
			}
		}

		if !foundUUID {
			t.Errorf("No UUID-named file found in uploads")
		}
	})

	t.Run("Magic Byte Validation", func(t *testing.T) {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("book", "fake.epub")
		part.Write([]byte("this is just a text file, not a zip"))
		writer.Close()

		req := httptest.NewRequest("POST", "/api/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, _ := app.Test(req)
		if resp.StatusCode != 400 {
			t.Errorf("Expected 400 for invalid magic bytes, got %d", resp.StatusCode)
		}
	})
}

func TestZipSlipProtection(t *testing.T) {
	services.EnsureUploadsDir()
	defer os.RemoveAll("uploads")

	t.Run("Rejects Traversal Paths", func(t *testing.T) {
		maliciousData, _ := createMaliciousEPUB()
		path := filepath.Join("uploads", "malicious.epub")
		os.WriteFile(path, maliciousData, 0644)

		_, err := services.ProcessLocalFile(path)
		if err == nil || !strings.Contains(err.Error(), "security: invalid file path") {
			t.Errorf("Expected security error for Zip Slip, got: %v", err)
		}
	})
}
