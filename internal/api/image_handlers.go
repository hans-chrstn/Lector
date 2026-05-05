package api

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/plugin"
	"github.com/user/lector/internal/services"
)

func (h *API) ProxyImage(c *fiber.Ctx) error {
	imgURL := c.Query("url")
	if imgURL == "" {
		return c.Status(400).SendString("Missing url")
	}

	if strings.HasPrefix(imgURL, "/uploads/") {
		return c.Redirect(imgURL)
	}

	cacheKey := "img:" + imgURL
	var cached []byte
	if ok, _ := services.GetCache(cacheKey, &cached); ok {
		c.Set("Content-Type", "image/jpeg")
		c.Set("Cache-Control", "public, max-age=604800")
		return c.Send(cached)
	}

	req, _ := http.NewRequest("GET", imgURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")

	plugin.HTTPMu.Lock()
	resp, err := plugin.HTTPClient.Do(req)
	plugin.HTTPMu.Unlock()
	if err != nil || resp.StatusCode != 200 {
		return c.Status(500).SendString("Failed to fetch image")
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(500).SendString("Failed to read image")
	}

	services.SetCache(cacheKey, data, 7*24*time.Hour)

	c.Set("Content-Type", resp.Header.Get("Content-Type"))
	c.Set("Cache-Control", "public, max-age=604800")
	return c.Send(data)
}
