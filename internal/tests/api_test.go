package tests

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/api"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/plugin"
	"github.com/user/lector/internal/repository"
)

func TestAPIRoutes(t *testing.T) {
	app := fiber.New()
	db.InitDB(":memory:")
	engine := &plugin.PluginEngine{
		Store:   repository.NewPluginRepository(),
		Plugins: make(map[string]*plugin.LuaPlugin),
	}
	api.RegisterRoutes(app, engine)

	t.Run("Get Plugins", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/plugins", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Get Documents", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/documents", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
		var documents []models.Document
		json.NewDecoder(resp.Body).Decode(&documents)
		if len(documents) != 0 {
			t.Errorf("Expected 0 documents, got %d", len(documents))
		}
	})
}
