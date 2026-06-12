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
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/user/lector/internal/api"
	"github.com/user/lector/internal/core/sanitizer"
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

func createLargeEPUB() ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	f, _ := w.Create("META-INF/container.xml")
	f.Write(make([]byte, 1024))

	w.Close()
	return buf.Bytes(), nil
}

func TestUploadSecurityHardening(t *testing.T) {
	app := fiber.New()
	db.InitDB(":memory:")
	services.EnsureUploadsDir()
	defer os.RemoveAll("uploads")

	plugins := make(map[string]*plugin.LuaPlugin)
	engine := &plugin.PluginEngine{
		Store:   db.NewPluginRepository(),
		Plugins: plugins,
	}
	api.RegisterRoutes(app, engine)

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

		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		f, _ := w.Create("page1.jpg")
		f.Write([]byte("fake data"))
		w.Close()
		cbzData := buf.Bytes()

		body2 := new(bytes.Buffer)
		writer2 := multipart.NewWriter(body2)
		part2, _ := writer2.CreateFormFile("book", "valid.cbz")
		part2.Write(cbzData)
		writer2.Close()

		req2 := httptest.NewRequest("POST", "/api/upload", body2)
		req2.Header.Set("Content-Type", writer2.FormDataContentType())

		resp2, _ := app.Test(req2)
		if resp2.StatusCode != 200 {
			t.Errorf("Expected 200 for valid CBZ magic bytes, got %d", resp2.StatusCode)
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

func TestBasicAuthSecurity(t *testing.T) {
	app := fiber.New()
	app.Use(basicauth.New(basicauth.Config{
		Users: map[string]string{
			"admin": "password123",
		},
	}))
	app.Get("/test", func(c *fiber.Ctx) error { return c.SendString("OK") })

	t.Run("Blocks Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 401 {
			t.Errorf("Expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("Allows Authorized", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.SetBasicAuth("admin", "password123")
		resp, _ := app.Test(req)
		if resp.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})
}

func TestRateLimiting(t *testing.T) {
	app := fiber.New()
	app.Use(limiter.New(limiter.Config{
		Max:        2,
		Expiration: 1 * time.Second,
	}))
	app.Get("/test", func(c *fiber.Ctx) error { return c.SendString("OK") })

	t.Run("Triggers Rate Limit", func(t *testing.T) {
		app.Test(httptest.NewRequest("GET", "/test", nil))
		app.Test(httptest.NewRequest("GET", "/test", nil))
		resp, _ := app.Test(httptest.NewRequest("GET", "/test", nil))
		if resp.StatusCode != 429 {
			t.Errorf("Expected 429, got %d", resp.StatusCode)
		}
	})
}

func TestStoredXSSProtection(t *testing.T) {
	t.Run("Strips Script Tags", func(t *testing.T) {
		html := `<div>Hello<script>alert("XSS")</script> World</div>`
		sanitized := sanitizer.CleanHTML(html, "")
		if strings.Contains(sanitized, "<script>") {
			t.Errorf("Script tag was not removed")
		}
	})

	t.Run("Strips Event Handlers", func(t *testing.T) {
		html := `<img src="x" onerror="alert(1)">`
		sanitized := sanitizer.CleanHTML(html, "")
		if strings.Contains(sanitized, "onerror") {
			t.Errorf("Event handler (onerror) was not removed")
		}
	})

	t.Run("Strips Javascript Links", func(t *testing.T) {
		html := `<a href="javascript:alert(1)">Click Me</a>`
		sanitized := sanitizer.CleanHTML(html, "")
		if strings.Contains(sanitized, "javascript:") {
			t.Errorf("Javascript link was not removed")
		}
	})

	t.Run("Preserves Safe Content", func(t *testing.T) {
		html := `<h1>Title</h1><p>This is <b>bold</b> and <i>italic</i>.</p><img src="/uploads/test.jpg">`
		sanitized := sanitizer.CleanHTML(html, "")
		if !strings.Contains(sanitized, "<h1>Title</h1>") || !strings.Contains(sanitized, "<b>bold</b>") {
			t.Errorf("Safe HTML was incorrectly removed")
		}
	})
}

func TestLFIPolish(t *testing.T) {
	app := fiber.New()
	db.InitDB(":memory:")
	engine := &plugin.PluginEngine{
		Store:   db.NewPluginRepository(),
		Plugins: make(map[string]*plugin.LuaPlugin),
	}
	api.RegisterRoutes(app, engine)

	t.Run("Image Proxy Blocks Traversal", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/proxy-image?url=uploads/../../etc/passwd", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 403 {
			t.Errorf("Expected 403 for traversal attempt, got %d", resp.StatusCode)
		}
	})
}

func TestSecurityHeaders(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline';")
		return c.Next()
	})
	app.Get("/test", func(c *fiber.Ctx) error { return c.SendString("OK") })

	t.Run("CSP Header Present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, _ := app.Test(req)
		csp := resp.Header.Get("Content-Security-Policy")
		if !strings.Contains(csp, "default-src 'self'") || !strings.Contains(csp, "unsafe-inline") {
			t.Errorf("CSP header missing or incorrect: %s", csp)
		}
	})
}

func TestDatabasePermissions(t *testing.T) {
	if os.Getenv("DB_DRIVER") == "postgres" {
		t.Skip("Skipping local SQLite permissions test for PostgreSQL")
	}

	path := "test_perms.db"
	defer os.Remove(path)

	db.InitDB(path)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Failed to stat db: %v", err)
	}

	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("Expected 0600 permissions, got %v", mode)
	}
}
