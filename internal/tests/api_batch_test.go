package tests

import (
	"bytes"
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

func TestBatchAPI(t *testing.T) {
	app := fiber.New()
	db.InitDB(":memory:")
	engine := &plugin.PluginEngine{
		Store:   repository.NewPluginRepository(),
		Plugins: make(map[string]*plugin.LuaPlugin),
	}
	api.RegisterRoutes(app, engine)

	doc1 := models.Document{Title: "Doc 1", URL: "url1", IsInLibrary: true}
	doc2 := models.Document{Title: "Doc 2", URL: "url2", IsInLibrary: true}
	db.DB.Create(&doc1)
	db.DB.Create(&doc2)

	t.Run("Batch Archive", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"ids":     []uint{doc1.ID, doc2.ID},
			"archive": true,
		})
		req := httptest.NewRequest("POST", "/api/documents/batch/archive", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req)

		if resp.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var updatedDoc1 models.Document
		db.DB.First(&updatedDoc1, doc1.ID)
		if !updatedDoc1.IsArchived {
			t.Errorf("Expected doc1 to be archived")
		}
	})

	t.Run("Batch Delete", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"ids": []uint{doc1.ID},
		})
		req := httptest.NewRequest("DELETE", "/api/documents/batch", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req)

		if resp.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var count int64
		db.DB.Model(&models.Document{}).Where("id = ?", doc1.ID).Count(&count)
		if count != 0 {
			t.Errorf("Expected doc1 to be deleted")
		}
	})
}
