package api

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func (h *API) ServePluginAsset(c *fiber.Ctx) error {
	name := strings.ToLower(c.Params("name"))
	wildcardPath := c.Params("*")

	if wildcardPath == "" {
		return c.Status(400).SendString("No asset path provided")
	}

	cleanPath := filepath.Clean(wildcardPath)
	if strings.HasPrefix(cleanPath, "..") {
		return c.Status(403).SendString("Invalid path")
	}

	pluginDir := getPluginDir()
	basePath := filepath.Join(pluginDir, name, "assets")
	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return c.Status(500).SendString("Internal server error")
	}

	assetPath := filepath.Join(basePath, cleanPath)
	absAssetPath, err := filepath.Abs(assetPath)
	if err != nil || !strings.HasPrefix(absAssetPath, absBasePath+string(filepath.Separator)) {
		return c.Status(403).SendString("Invalid path")
	}

	if _, err := os.Stat(absAssetPath); os.IsNotExist(err) {
		return c.Status(404).SendString("Asset not found")
	}

	return c.SendFile(absAssetPath)
}
