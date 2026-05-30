package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/plugin"
	"github.com/user/lector/internal/services"
)

func (h *API) ProxyImage(c *fiber.Ctx) error {
	imgURL := c.Query("url")
	if imgURL == "" {
		return c.Status(400).SendString("Missing url")
	}
	if strings.HasPrefix(imgURL, "/uploads/") || strings.HasPrefix(imgURL, "uploads/") || !strings.Contains(imgURL, "://") {
		absPath, _ := filepath.Abs(filepath.Clean(strings.TrimPrefix(imgURL, "/")))
		if !services.IsPathAuthorized(absPath) {
			return c.Status(403).SendString("Security: Access denied")
		}
		return c.SendFile(absPath)
	}
	if imgURL == "MISSING" {
		return c.Status(404).SendString("No image URL provided")
	}
	if !strings.HasPrefix(imgURL, "http://") && !strings.HasPrefix(imgURL, "https://") {
		if strings.HasPrefix(imgURL, "//") {
			imgURL = "https:" + imgURL
		} else {
			return c.Status(400).SendString("Invalid image URL")
		}
	}
	type CachedImage struct {
		Data        []byte `json:"data"`
		ContentType string `json:"content_type"`
	}
	cacheKey := "img:" + imgURL
	var cached CachedImage
	var item models.CacheItem
	if err := db.DB.Where("key = ? AND expires_at > ?", cacheKey, time.Now()).First(&item).Error; err == nil {
		if err := json.Unmarshal(item.Value, &cached); err == nil && len(cached.Data) > 0 {
			if cached.ContentType != "" {
				c.Set("Content-Type", cached.ContentType)
			} else {
				c.Set("Content-Type", "image/jpeg")
			}
			c.Set("Cache-Control", "public, max-age=604800")
			return c.Send(cached.Data)
		}
		var oldData []byte
		if err := json.Unmarshal(item.Value, &oldData); err == nil && len(oldData) > 0 {
			c.Set("Content-Type", "image/jpeg")
			c.Set("Cache-Control", "public, max-age=604800")
			return c.Send(oldData)
		}
	}
	req, _ := http.NewRequest("GET", imgURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
	resp, err := plugin.HTTPClient.Do(req)
	if err != nil {
		fmt.Printf("[ImageProxy] Error fetching %s: %v\n", imgURL, err)
		return c.Status(404).SendString("Image not found")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		if resp.StatusCode != 404 {
			fmt.Printf("[ImageProxy] Remote server returned %d for %s\n", resp.StatusCode, imgURL)
		}
		return c.Status(404).SendString("Image not found")
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(500).SendString("Failed to read image")
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" || !strings.HasPrefix(contentType, "image/") {
		contentType = "image/jpeg"
	}
	services.SetCache(cacheKey, CachedImage{Data: data, ContentType: contentType}, 7*24*time.Hour)
	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "public, max-age=604800")
	return c.Send(data)
}
