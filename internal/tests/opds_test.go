package tests

import (
	"encoding/xml"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/api"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/plugin"
	"github.com/user/lector/internal/repository"
)

func TestOPDSServer(t *testing.T) {
	app := fiber.New()
	db.InitDB(":memory:")
	engine := &plugin.PluginEngine{
		Store:   repository.NewPluginRepository(),
		Plugins: make(map[string]*plugin.LuaPlugin),
	}
	api.RegisterRoutes(app, engine)

	t.Run("Root Feed Returns XML", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/opds", nil)
		resp, _ := app.Test(req)

		if resp.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		if !strings.Contains(resp.Header.Get("Content-Type"), "application/atom+xml") {
			t.Errorf("Expected atom+xml content type, got %s", resp.Header.Get("Content-Type"))
		}
	})

	t.Run("All Documents Feed with Data", func(t *testing.T) {
		doc := models.Document{
			Title:    "OPDS Test Book",
			Author:   "Tester",
			Synopsis: "Test synopsis content",
		}
		db.DB.Create(&doc)

		req := httptest.NewRequest("GET", "/api/opds/all", nil)
		resp, _ := app.Test(req)

		if resp.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var feed api.OPDSFeed
		xml.NewDecoder(resp.Body).Decode(&feed)

		found := false
		for _, entry := range feed.Entries {
			if entry.Title == "OPDS Test Book" {
				found = true
				if entry.Author.Name != "Tester" {
					t.Errorf("Incorrect author in OPDS entry")
				}

				hasAcquisition := false
				for _, link := range entry.Links {
					if strings.Contains(link.Rel, "acquisition") {
						hasAcquisition = true
						if !strings.Contains(link.Href, "/api/documents") {
							t.Errorf("Invalid acquisition link: %s", link.Href)
						}
					}
				}
				if !hasAcquisition {
					t.Errorf("Missing acquisition link in OPDS entry")
				}
			}
		}

		if !found {
			t.Errorf("Test book not found in OPDS feed")
		}
	})
}
