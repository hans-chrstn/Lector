package api

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/plugin"
	"github.com/user/lector/internal/services"
)

type API struct {
	Engine          *plugin.PluginEngine
	DocumentService services.DocumentService
}

func RegisterRoutes(app *fiber.App, engine *plugin.PluginEngine) {
	docService := services.NewDocumentService()

	h := &API{
		Engine:          engine,
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
	api.Get("/plugins/:name/assets/*", h.ServePluginAsset)
	api.Get("/plugins/:name/directory/:id", h.PluginDirectory)

	api.Get("/search", h.Search)
	api.Get("/discovery/search", h.Search)

	api.Get("/documents", h.GetDocuments)
	api.Get("/documents/search", h.SearchLibrary)
	api.Post("/documents/ensure", h.EnsureDocument)
	api.Post("/documents/batch/refresh", h.BatchRefreshDocuments)
	api.Delete("/documents/batch", h.BatchDeleteDocuments)
	api.Post("/documents/batch/move", h.BatchMoveDocuments)
	api.Post("/documents/batch/archive", h.BatchArchiveDocuments)
	api.Post("/documents/batch/mark-read", h.BatchMarkReadDocuments)

	api.Get("/documents/:id", h.GetDocumentByID)
	api.Post("/documents/:id/library", h.ToggleLibrary)
	api.Put("/documents/:id/metadata", h.UpdateMetadata)
	api.Post("/documents/:id/cover", h.UpdateCover)
	api.Get("/documents/:id/progress", h.GetDocumentProgress)
	api.Post("/documents/:id/refresh", h.RefreshDocument)
	api.Post("/documents/:id/migrate", h.MigrateDocument)
	api.Get("/documents/:id/export", h.ExportDocument)
	api.Get("/documents/:id/archive-image", h.GetArchiveImage)

	api.Get("/history", h.GetHistory)
	api.Delete("/history", h.ClearHistory)
	api.Delete("/history/batch", h.BatchDeleteHistory)
	api.Delete("/history/:id", h.DeleteHistory)

	api.Post("/analytics/track", h.TrackAnalytics)
	api.Get("/analytics", h.GetAnalytics)

	api.Get("/library/paths", h.GetLibraryPaths)
	api.Post("/library/paths", h.AddLibraryPath)
	api.Delete("/library/paths/:id", h.DeleteLibraryPath)
	api.Post("/library/scan", h.ScanLibrary)
	api.Get("/library/scan/status", h.ScanStatus)

	api.Get("/chapters/:id", h.GetChapterByID)
	api.Post("/chapters/:id/read", h.ToggleChapterRead)
	api.Post("/chapters/batch", h.BatchUpdateChapters)
	api.Post("/progress", h.SyncProgress)

	api.Get("/groups", h.GetGroups)
	api.Post("/groups", h.CreateGroup)

	api.Post("/upload", h.HandleUpload)
	api.Get("/proxy-image", h.ProxyImage)
	api.Get("/proxy-stream", h.ProxyStream)
	api.Get("/proxy-segment/:b64url/:b64ref/*", h.ProxySegment)

	api.Get("/documents/:documentId/bookmarks", h.GetBookmarks)
	api.Post("/bookmarks", h.AddBookmark)
	api.Delete("/bookmarks/:id", h.DeleteBookmark)
	api.Get("/documents/:documentId/notes", h.GetNotes)
	api.Post("/notes", h.AddNote)
	api.Delete("/notes/:id", h.DeleteNote)

	api.Get("/opds", h.GetOPDSRoot)
	api.Get("/opds/all", h.GetOPDSAll)
}

func (h *API) IsLocalNetworkAuthorized(targetURL string) bool {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return false
	}
	host := parsed.Hostname()
	if h.Engine == nil {
		return false
	}
	for _, p := range h.Engine.Plugins {
		if p.HasCapability("local_network") {
			for _, perm := range p.Permissions {
				if perm == "*" {
					return true
				}
				if host == perm || strings.HasSuffix(host, "."+perm) {
					return true
				}
			}
		}
	}
	return false
}
