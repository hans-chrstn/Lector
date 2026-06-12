package tests

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/api"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/plugin"

	"github.com/user/lector/internal/services"
)

func createMockEPUB() ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	f, _ := w.Create("mimetype")
	f.Write([]byte("application/epub+zip"))

	f, _ = w.Create("META-INF/container.xml")
	f.Write([]byte(`<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	f, _ = w.Create("OEBPS/content.opf")
	f.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Mock Book</dc:title>
    <dc:creator>Test Author</dc:creator>
  </metadata>
  <manifest>
    <item id="ch1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="ch1"/>
  </spine>
</package>`))

	f, _ = w.Create("OEBPS/chapter1.xhtml")
	f.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<html>
<head><title>Chapter 1</title></head>
<body><p>Hello World</p></body>
</html>`))

	w.Close()
	return buf.Bytes(), nil
}

func TestUploadAPI(t *testing.T) {
	app := fiber.New()
	db.InitDB(":memory:")
	services.EnsureUploadsDir()
	defer os.RemoveAll("uploads")

	engine := &plugin.PluginEngine{
		Store:   db.NewPluginRepository(),
		Plugins: make(map[string]*plugin.LuaPlugin),
	}
	api.RegisterRoutes(app, engine)

	epubData, _ := createMockEPUB()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("book", "test.epub")
	part.Write(epubData)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req, 10000)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	var document models.Document
	json.NewDecoder(resp.Body).Decode(&document)
	if document.Title != "Mock Book" {
		t.Errorf("Expected title 'Mock Book', got '%s'", document.Title)
	}

	if len(document.Chapters) != 1 {
		t.Errorf("Expected 1 chapter, got %d", len(document.Chapters))
	}
}
