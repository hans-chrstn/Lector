package api

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/core/httpclient"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/services"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func populateChapterCounts(ctx context.Context, docs []models.Document) {
	if len(docs) == 0 {
		return
	}
	var docIDs []uint
	for _, d := range docs {
		docIDs = append(docIDs, d.ID)
	}

	type ChapterCount struct {
		DocumentID uint
		ReadCount  int
		TotalCount int
	}
	var counts []ChapterCount
	db.DB.WithContext(ctx).Model(&models.Chapter{}).
		Select("document_id, sum(case when is_read = 1 then 1 else 0 end) as read_count, count(*) as total_count").
		Where("document_id IN ?", docIDs).
		Group("document_id").Find(&counts)

	countMap := make(map[uint]ChapterCount)
	for _, c := range counts {
		countMap[c.DocumentID] = c
	}
	for i := range docs {
		docs[i].ReadChapters = countMap[docs[i].ID].ReadCount
		docs[i].TotalChapters = countMap[docs[i].ID].TotalCount
	}
}

func (h *API) GetDocuments(c *fiber.Ctx) error {
	showArchived := c.Query("archived") == "true"
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	query := db.DB.WithContext(c.UserContext()).Where("is_in_library = ? AND is_archived = ?", true, showArchived)

	if limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			query = query.Limit(limit)
		}
	}
	if offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset > 0 {
			query = query.Offset(offset)
		}
	}

	var docs []models.Document
	err := query.Find(&docs).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	populateChapterCounts(c.UserContext(), docs)
	return c.JSON(docs)
}

func (h *API) EnsureDocument(c *fiber.Ctx) error {
	var req struct {
		URL    string `json:"url"`
		Source string `json:"source"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	url, pluginName := req.URL, req.Source
	doc, _ := h.DocumentService.GetByURL(url)

	if doc == nil {
		if pluginName == "local" {
			return c.Status(404).JSON(fiber.Map{"error": "Local document not found"})
		}

		s, ok := h.Engine.Plugins[pluginName]
		if !ok {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid plugin"})
		}

		fetched, err := s.GetDocument(url)
		if err != nil || fetched.Title == "" {
			return c.Status(500).JSON(fiber.Map{"error": "Fetch failed"})
		}
		fetched.Source = pluginName
		chapters := fetched.Chapters
		fetched.Chapters = nil
		if err := db.DB.WithContext(c.UserContext()).Clauses(clause.OnConflict{UpdateAll: true}).Create(&fetched).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to save document"})
		}
		doc = &fetched
		for i := range chapters {
			chapters[i].DocumentID = doc.ID
			chapters[i].ID = 0
			chapters[i].Order = i + 1
		}
		if err := db.DB.WithContext(c.UserContext()).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "document_id"}, {Name: "url"}},
			DoUpdates: clause.AssignmentColumns([]string{"title", "order_val"}),
		}).CreateInBatches(chapters, 100).Error; err != nil {
			fmt.Printf("[API] Error creating chapters for %s: %v\n", doc.Title, err)
		}
	} else {
		doc.Source = pluginName
		db.DB.WithContext(c.UserContext()).Save(doc)

		var count int64
		db.DB.WithContext(c.UserContext()).Model(&models.Chapter{}).Where("document_id = ?", doc.ID).Count(&count)

		force := c.Query("force") == "true"

		if count == 0 || force {
			if pluginName == "local" {
				services.ProcessLocalFile(doc.LocalPath)
			} else {
				s, ok := h.Engine.Plugins[pluginName]
				if ok {
					fetched, err := s.GetDocument(url)
					if err != nil || fetched.Title == "" {
						fmt.Printf("[API] Error refetching chapters for %s: %v\n", doc.Title, err)
					} else {
						doc.CoverURL = fetched.CoverURL
						doc.Author = fetched.Author
						doc.Synopsis = fetched.Synopsis
						db.DB.WithContext(c.UserContext()).Save(doc)

						for i := range fetched.Chapters {
							fetched.Chapters[i].DocumentID = doc.ID
							fetched.Chapters[i].ID = 0
							fetched.Chapters[i].Order = i + 1
						}
						if err := db.DB.WithContext(c.UserContext()).Clauses(clause.OnConflict{
							Columns:   []clause.Column{{Name: "document_id"}, {Name: "url"}},
							DoUpdates: clause.AssignmentColumns([]string{"title", "order_val"}),
						}).CreateInBatches(fetched.Chapters, 100).Error; err != nil {
							fmt.Printf("[API] Error creating chapters for existing %s: %v\n", doc.Title, err)
						}
					}
				}
			}
		}
	}

	db.DB.WithContext(c.UserContext()).Preload("Chapters", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "document_id", "title", "url", "order_val", "is_read", "status", "metadata").Order("order_val ASC")
	}).First(doc, doc.ID)

	var readCount int64
	db.DB.WithContext(c.UserContext()).Model(&models.Chapter{}).Where("document_id = ? AND is_read = ?", doc.ID, true).Count(&readCount)
	doc.ReadChapters = int(readCount)
	doc.TotalChapters = len(doc.Chapters)

	return c.JSON(doc)
}

func (h *API) RefreshDocument(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	doc, err := h.DocumentService.GetByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Document not found"})
	}

	if doc.Source != "local" {
		s, ok := h.Engine.Plugins[doc.Source]
		if ok {
			fetched, err := s.GetDocument(doc.URL)
			if err != nil || fetched.Title == "" {
				fmt.Printf("[API] Error refreshing chapters for %s: %v\n", doc.Title, err)
			} else {
				doc.CoverURL = fetched.CoverURL
				doc.Author = fetched.Author
				doc.Synopsis = fetched.Synopsis
				db.DB.WithContext(c.UserContext()).Save(doc)

				for i := range fetched.Chapters {
					fetched.Chapters[i].DocumentID = doc.ID
					fetched.Chapters[i].ID = 0
					fetched.Chapters[i].Order = i + 1
				}
				if err := db.DB.WithContext(c.UserContext()).Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "document_id"}, {Name: "url"}},
					DoUpdates: clause.AssignmentColumns([]string{"title", "order_val"}),
				}).CreateInBatches(fetched.Chapters, 100).Error; err != nil {
					fmt.Printf("[API] Error creating chapters for refreshed %s: %v\n", doc.Title, err)
				}
			}
		}
	} else {
		services.ProcessLocalFile(doc.LocalPath)
	}

	return h.GetDocumentByID(c)
}

func (h *API) BatchRefreshDocuments(c *fiber.Ctx) error {
	var req struct {
		IDs []int `json:"ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	job, err := services.DefaultJobManager.Enqueue("batch_refresh", req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to enqueue job"})
	}

	return c.JSON(fiber.Map{"status": "success", "job_id": job.ID})
}

func (h *API) SearchLibrary(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.JSON([]models.Document{})
	}

	var documents []models.Document

	if db.DB.Dialector.Name() == "postgres" {
		db.DB.WithContext(c.UserContext()).
			Where("title ILIKE ? OR author ILIKE ? OR synopsis ILIKE ? OR genres ILIKE ?",
				"%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%").
			Find(&documents)
	} else {
		err := db.DB.WithContext(c.UserContext()).Raw(`
			SELECT d.* FROM documents d
			JOIN document_search ds ON d.id = ds.rowid
			WHERE document_search MATCH ?
			ORDER BY rank`, query+"*").Scan(&documents).Error

		if err != nil {
			db.DB.WithContext(c.UserContext()).Where("title LIKE ? OR author LIKE ? OR genres LIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%").Find(&documents)
		}
	}

	return c.JSON(documents)
}

func (h *API) GetDocumentByID(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	doc, err := h.DocumentService.GetByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Document not found"})
	}

	db.DB.WithContext(c.UserContext()).Preload("Chapters", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "document_id", "title", "url", "order_val", "is_read", "status", "metadata").Order("order_val ASC")
	}).First(doc, doc.ID)

	var readCount int64
	db.DB.WithContext(c.UserContext()).Model(&models.Chapter{}).Where("document_id = ? AND is_read = ?", doc.ID, true).Count(&readCount)
	doc.ReadChapters = int(readCount)
	doc.TotalChapters = len(doc.Chapters)

	return c.JSON(doc)
}

func (h *API) ToggleLibrary(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	inLibrary := c.Query("is_in_library") == "true"
	groupID, _ := strconv.Atoi(c.Query("group_id", "0"))

	h.DocumentService.ToggleLibrary(uint(id), inLibrary, uint(groupID))
	return c.SendString("Updated")
}

func (h *API) GetDocumentProgress(c *fiber.Ctx) error {
	var p models.ReadingProgress
	db.DB.WithContext(c.UserContext()).Where("document_id = ?", c.Params("id")).First(&p)
	return c.JSON(p)
}

func (h *API) GetHistory(c *fiber.Ctx) error {
	var docs []models.Document
	err := db.DB.WithContext(c.UserContext()).Table("documents").
		Select("documents.*").
		Joins("JOIN reading_progresses ON reading_progresses.document_id = documents.id").
		Where("documents.is_archived = ?", false).
		Order("reading_progresses.updated_at DESC").
		Limit(50).
		Scan(&docs).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	populateChapterCounts(c.UserContext(), docs)
	return c.JSON(docs)
}

func (h *API) UpdateMetadata(c *fiber.Ctx) error {
	var req map[string]interface{}
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Params("id"))
	h.DocumentService.UpdateMetadata(uint(id), req)
	return c.SendString("Updated")
}

func (h *API) UpdateCover(c *fiber.Ctx) error {
	file, err := c.FormFile("cover")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "No file uploaded"})
	}

	ext := filepath.Ext(file.Filename)
	localCoverName := fmt.Sprintf("custom_%s%s", c.Params("id"), ext)
	localCoverPath := filepath.Join("uploads", localCoverName)

	if err := c.SaveFile(file, localCoverPath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save cover"})
	}

	db.DB.WithContext(c.UserContext()).Model(&models.Document{}).Where("id = ?", c.Params("id")).Update("cover_url", "/uploads/"+localCoverName)
	return c.JSON(fiber.Map{"url": "/uploads/" + localCoverName})
}

func (h *API) MigrateDocument(c *fiber.Ctx) error {
	var req struct {
		URL    string `json:"url"`
		Source string `json:"source"`
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Params("id"))
	doc, _ := h.DocumentService.GetByID(uint(id))
	doc.URL = req.URL
	doc.Source = req.Source
	h.DocumentService.Save(doc)
	return c.SendString("Migrated")
}

func (h *API) ProxyStream(c *fiber.Ctx) error {
	u := c.Query("url")
	ref := c.Query("referer")
	if u == "" {
		return c.SendStatus(400)
	}
	return h.handleProxy(c, u, ref)
}

func (h *API) ProxySegment(c *fiber.Ctx) error {
	b64url := c.Params("b64url")
	b64ref := c.Params("b64ref")
	wildcardPath := c.Params("*")
	queryString := string(c.Request().URI().QueryString())
	if queryString != "" {
		wildcardPath += "?" + queryString
	}

	decodedURL, err := base64.RawURLEncoding.DecodeString(b64url)
	if err != nil {
		return c.Status(400).SendString("Invalid b64url")
	}
	decodedRef, err := base64.RawURLEncoding.DecodeString(b64ref)
	if err != nil {
		return c.Status(400).SendString("Invalid b64ref")
	}

	baseURL, err := url.Parse(string(decodedURL))
	if err != nil {
		return c.Status(400).SendString("Invalid url")
	}

	relURL, err := url.Parse(wildcardPath)
	if err != nil {
		return c.Status(400).SendString("Invalid path")
	}

	finalURL := baseURL.ResolveReference(relURL).String()
	return h.handleProxy(c, finalURL, string(decodedRef))
}

func (h *API) handleProxy(c *fiber.Ctx, u string, ref string) error {
	parsed, err := url.Parse(u)
	if err != nil {
		return c.Status(400).SendString("Invalid URL")
	}

	fetch := func(targetURL string, client *http.Client) (*http.Response, error) {
		req, _ := http.NewRequest("GET", targetURL, nil)
		if ref != "" {
			req.Header.Set("Referer", ref)
		} else {
			req.Header.Set("Referer", parsed.Scheme+"://"+parsed.Host+"/")
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		return client.Do(req)
	}

	client := httpclient.InternalClient
	if h.IsLocalNetworkAuthorized(u) {
		client = httpclient.RelaxedClient
	}

	resp, err := fetch(u, client)

	if err != nil && (strings.Contains(err.Error(), "handshake") || strings.Contains(err.Error(), "wrong version number") || strings.Contains(err.Error(), "connection reset") || strings.Contains(err.Error(), "record length")) {
		fmt.Printf("[ProxyStream] Primary TLS failed for %s, trying Relaxed Client\n", u)
		resp, err = fetch(u, httpclient.RelaxedClient)
	}

	if err != nil && parsed.Scheme == "https" {
		fmt.Printf("[ProxyStream] TLS failed for %s, falling back to HTTP\n", u)
		httpURL := "http://" + parsed.Host + parsed.RequestURI()
		resp, err = fetch(httpURL, client)
	}

	if err != nil {
		fmt.Printf("[ProxyStream] Request failed for %s: %v\n", u, err)
		return c.Status(500).SendString("Fetch error: " + err.Error())
	}
	contentType := resp.Header.Get("Content-Type")

	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return c.Status(resp.StatusCode).SendString("Remote server returned error")
	}

	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", resp.Header.Get("Cache-Control"))

	origin := c.Get("Origin")
	if !isAllowedOrigin(origin, c) {
		resp.Body.Close()
		return c.SendStatus(403)
	}
	c.Set("Access-Control-Allow-Origin", origin)
	c.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Set("Access-Control-Allow-Headers", "Range, Content-Type")
	c.Set("Access-Control-Allow-Credentials", "true")

	if c.Method() == "OPTIONS" {
		resp.Body.Close()
		return c.SendStatus(204)
	}

	if strings.Contains(contentType, "dash+xml") || strings.Contains(u, ".mpd") {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		rewritten := rewriteDASHManifest(string(body), u, ref, c.BaseURL())
		return c.Send([]byte(rewritten))
	}

	if strings.Contains(contentType, "mpegURL") || strings.Contains(u, ".m3u8") {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		rewritten := rewriteHLSManifest(string(body), u, ref, c.BaseURL())
		return c.Send([]byte(rewritten))
	}

	if resp.ContentLength > 0 {
		c.Set("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
	}
	if cr := resp.Header.Get("Content-Range"); cr != "" {
		c.Set("Content-Range", cr)
	}
	if ar := resp.Header.Get("Accept-Ranges"); ar != "" {
		c.Set("Accept-Ranges", ar)
	}
	c.Status(resp.StatusCode)
	return c.SendStream(resp.Body)
}

func isAllowedOrigin(origin string, c *fiber.Ctx) bool {
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := parsed.Hostname()
	return host == "localhost" || host == "127.0.0.1" || host == "::1" || host == c.Hostname()
}

func rewriteDASHManifest(manifest, originalURL, referer, baseURL string) string {
	b64url := base64.RawURLEncoding.EncodeToString([]byte(originalURL))
	b64ref := base64.RawURLEncoding.EncodeToString([]byte(referer))
	proxyBase := baseURL + "/api/proxy-segment/" + b64url + "/" + b64ref + "/"
	re := regexp.MustCompile(`(<MPD[^>]*>)`)
	return re.ReplaceAllString(manifest, "${1}\n\t<BaseURL>"+proxyBase+"</BaseURL>")
}

func rewriteHLSManifest(manifest, originalURL, referer, baseURL string) string {
	base, _ := url.Parse(originalURL)
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(manifest))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			result.WriteString(line + "\n")
			continue
		}
		u, err := url.Parse(line)
		if err == nil {
			abs := base.ResolveReference(u).String()
			result.WriteString(baseURL + "/api/proxy-stream?url=" + url.QueryEscape(abs) + "&referer=" + url.QueryEscape(referer) + "\n")
		} else {
			result.WriteString(line + "\n")
		}
	}
	return result.String()
}

func (h *API) GetArchiveImage(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	fileName := c.Query("file")
	if fileName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing file parameter"})
	}

	var doc models.Document
	if err := db.DB.WithContext(c.UserContext()).First(&doc, uint(id)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Document not found"})
	}

	absPath, _ := filepath.Abs(doc.LocalPath)
	if !services.IsPathAuthorized(absPath) {
		return c.Status(403).JSON(fiber.Map{"error": "Security: Access denied"})
	}

	data, contentType, err := services.GetImageFromArchive(uint(id), fileName)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if contentType == "" {
		contentType = "image/jpeg"
	}

	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "public, max-age=604800")
	return c.Send(data)
}
