package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/plugin"
	"github.com/user/lector/internal/repository"
	"github.com/user/lector/internal/services"
)

type API struct {
	Plugins         map[string]*plugin.LuaPlugin
	DocumentService services.DocumentService
}

func RegisterRoutes(app *fiber.App, plugins map[string]*plugin.LuaPlugin) {
	docRepo := repository.NewRepository[models.Document](db.DB)
	chapterRepo := repository.NewRepository[models.Chapter](db.DB)
	docService := services.NewDocumentService(docRepo, chapterRepo)

	h := &API{
		Plugins:         plugins,
		DocumentService: docService,
	}

	api := app.Group("/api")

	api.Get("/plugins", h.GetActivePlugins)
	api.Get("/plugins/all", h.GetPlugins)
	api.Get("/plugins/manifest", h.GetPluginsManifest)
	api.Post("/plugins/upload", h.UploadPlugin)
	api.Post("/plugins/reorder", h.ReorderPlugins)
	api.Post("/plugins/:name/toggle", h.TogglePlugin)
	api.Delete("/plugins/:name", h.DeletePlugin)
	api.Post("/plugins/:name/rpc/:method", h.PluginRPC)

	api.Get("/search", h.Search)
	api.Get("/discovery/search", h.Search)

	api.Get("/documents", h.GetDocuments)
	api.Post("/documents/ensure", h.EnsureDocument)
	api.Get("/documents/:id", h.GetDocumentByID)
	api.Post("/documents/:id/library", h.ToggleLibrary)
	api.Put("/documents/:id/metadata", h.UpdateMetadata)
	api.Post("/documents/:id/cover", h.UpdateCover)
	api.Get("/documents/:id/progress", h.GetDocumentProgress)
	api.Post("/documents/:id/migrate", h.MigrateDocument)
	api.Get("/documents/:id/export", h.ExportDocument)
	api.Get("/documents/:id/archive-image", h.GetArchiveImage)
	api.Delete("/documents/batch", h.BatchDeleteDocuments)
	api.Post("/documents/batch/move", h.BatchMoveDocuments)
	api.Post("/documents/batch/archive", h.BatchArchiveDocuments)
	api.Post("/documents/batch/mark-read", h.BatchMarkReadDocuments)
	api.Get("/history", h.GetHistory)
	api.Delete("/history", h.ClearHistory)
	api.Delete("/history/batch", h.BatchDeleteHistory)
	api.Delete("/history/:id", h.DeleteHistory)

	api.Get("/chapters/:id", h.GetChapterByID)
	api.Post("/chapters/:id/read", h.ToggleChapterRead)
	api.Post("/chapters/batch", h.BatchUpdateChapters)
	api.Post("/progress", h.SyncProgress)

	api.Get("/groups", h.GetGroups)
	api.Post("/groups", h.CreateGroup)

	api.Post("/upload", h.HandleUpload)
	api.Get("/proxy-image", h.ProxyImage)

	api.Get("/documents/:documentId/bookmarks", h.GetBookmarks)
	api.Post("/bookmarks", h.AddBookmark)
	api.Delete("/bookmarks/:id", h.DeleteBookmark)
	api.Get("/documents/:documentId/notes", h.GetNotes)
	api.Post("/notes", h.AddNote)
	api.Delete("/notes/:id", h.DeleteNote)
}
