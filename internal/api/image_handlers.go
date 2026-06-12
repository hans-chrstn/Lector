package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/core/httpclient"
	"github.com/user/lector/internal/services"
)

func (h *API) ProxyImage(c *fiber.Ctx) error {
	imgURL := c.Query("url")
	if imgURL == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing url"})
	}
	if strings.HasPrefix(imgURL, "/uploads/") || strings.HasPrefix(imgURL, "uploads/") || !strings.Contains(imgURL, "://") {
		absPath, _ := filepath.Abs(filepath.Clean(strings.TrimPrefix(imgURL, "/")))
		if !services.IsPathAuthorized(absPath) {
			return c.Status(403).JSON(fiber.Map{"error": "Security: Access denied"})
		}
		return c.SendFile(absPath)
	}
	if imgURL == "MISSING" {
		return c.SendStatus(204)
	}
	if !strings.HasPrefix(imgURL, "http://") && !strings.HasPrefix(imgURL, "https://") {
		if strings.HasPrefix(imgURL, "//") {
			imgURL = "https:" + imgURL
		} else {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid image URL"})
		}
	}
	hashBytes := sha256.Sum256([]byte(imgURL))
	hashStr := hex.EncodeToString(hashBytes[:])
	cacheDir := filepath.Join("uploads", "cache", "images")
	os.MkdirAll(cacheDir, 0755)

	matches, _ := filepath.Glob(filepath.Join(cacheDir, hashStr+".*"))
	if len(matches) > 0 {
		c.Set("Cache-Control", "public, max-age=604800")
		return c.SendFile(matches[0])
	}

	req, _ := http.NewRequest("GET", imgURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")

	client := httpclient.InternalClient
	if h.IsLocalNetworkAuthorized(imgURL) {
		client = httpclient.RelaxedClient
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[ImageProxy] Error fetching %s: %v\n", imgURL, err)
		return c.SendStatus(204)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		if resp.StatusCode != 404 {
			fmt.Printf("[ImageProxy] Remote server returned %d for %s\n", resp.StatusCode, imgURL)
		}
		return c.SendStatus(204)
	}

	contentType := resp.Header.Get("Content-Type")
	ext := ".jpg"
	switch contentType {
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	case "image/gif":
		ext = ".gif"
	case "image/avif":
		ext = ".avif"
	case "image/svg+xml":
		ext = ".svg"
	}

	width := c.QueryInt("width", 0)

	if width > 0 {
		hashStr = fmt.Sprintf("%s_%d", hashStr, width)
		ext = ".jpg"
	}

	cachedPath := filepath.Join(cacheDir, hashStr+ext)

	if width > 0 {
		if _, err := os.Stat(cachedPath); err == nil {
			c.Set("Cache-Control", "public, max-age=604800")
			return c.SendFile(cachedPath)
		}
	}

	out, err := os.Create(cachedPath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create cache file"})
	}
	defer out.Close()

	if width > 0 {
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			io.Copy(out, resp.Body)
		} else {
			bounds := img.Bounds()
			ratio := float64(bounds.Dx()) / float64(bounds.Dy())
			height := int(float64(width) / ratio)
			dst := image.NewRGBA(image.Rect(0, 0, width, height))
			draw.BiLinear.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
			jpeg.Encode(out, dst, &jpeg.Options{Quality: 80})
		}
	} else {
		io.Copy(out, resp.Body)
	}

	c.Set("Cache-Control", "public, max-age=604800")
	return c.SendFile(cachedPath)
}
