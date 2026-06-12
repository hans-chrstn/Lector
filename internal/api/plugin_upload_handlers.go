package api

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/plugin"
)

func getPluginDir() string {
	return "plugins"
}

func (h *API) UploadPlugin(c *fiber.Ctx) error {
	file, err := c.FormFile("plugin")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "No file uploaded"})
	}

	name := strings.ToLower(strings.TrimSpace(c.FormValue("name")))
	name = filepath.Base(filepath.Clean(name))
	if name == "." || name == "/" {
		name = ""
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))

	if ext != ".zip" && ext != ".lua" {
		return c.Status(400).JSON(fiber.Map{"error": "Only .zip or .lua files are allowed"})
	}

	pluginDir := getPluginDir()
	os.MkdirAll(pluginDir, 0755)

	tempName := fmt.Sprintf("temp_%d", time.Now().UnixNano())
	tempPath := filepath.Join(pluginDir, tempName+ext)
	if err := c.SaveFile(file, tempPath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save upload"})
	}
	defer os.RemoveAll(tempPath)

	var detectedID string
	var destDir string

	if ext == ".lua" {
		p, err := plugin.NewLuaPlugin("probe", tempPath, h.Engine.Store)
		if err == nil {
			if p.Name != "probe" {
				detectedID = p.Name
			}
			p.L.Close()
		}

		finalID := name
		if finalID == "" {
			finalID = detectedID
		}
		if finalID == "" {
			cleanFileName := filepath.Base(filepath.Clean(file.Filename))
			baseName := strings.ToLower(strings.TrimSuffix(cleanFileName, ext))
			if baseName == "init" || baseName == "" || baseName == "." || baseName == "/" {
				return c.Status(409).JSON(fiber.Map{"error": "Generic or invalid filename detected. Please provide a plugin name."})
			}
			finalID = baseName
		}

		destDir = filepath.Join(pluginDir, finalID)
		os.MkdirAll(destDir, 0755)
		if err := os.Rename(tempPath, filepath.Join(destDir, "init.lua")); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to move plugin"})
		}
		name = finalID
	} else {
		destDir = filepath.Join(pluginDir, name)
		if name == "" {
			cleanFileName := filepath.Base(filepath.Clean(file.Filename))
			fallback := strings.TrimSuffix(cleanFileName, ext)
			if fallback == "" || fallback == "." || fallback == "/" {
				fallback = tempName
			}
			destDir = filepath.Join(pluginDir, fallback)
		}
		os.MkdirAll(destDir, 0755)

		r, err := zip.OpenReader(tempPath)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid zip file"})
		}
		defer r.Close()

		for _, f := range r.File {
			fpath := filepath.Join(destDir, f.Name)
			if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
				return c.Status(400).JSON(fiber.Map{"error": "Invalid file path in zip"})
			}
			if f.FileInfo().IsDir() {
				os.MkdirAll(fpath, 0755)
				continue
			}
			os.MkdirAll(filepath.Dir(fpath), 0755)
			outFile, _ := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			rc, _ := f.Open()
			io.Copy(outFile, rc)
			outFile.Close()
			rc.Close()
		}

		entryPoint := filepath.Join(destDir, "init.lua")
		p, err := plugin.NewLuaPlugin("probe", entryPoint, h.Engine.Store)
		if err == nil {
			detectedID = p.Name
			p.L.Close()
		}

		if detectedID != "" && detectedID != "probe" && detectedID != name {
			newDest := filepath.Join(pluginDir, detectedID)
			os.RemoveAll(newDest)
			os.Rename(destDir, newDest)
			name = detectedID
		} else if name == "" {
			name = strings.ToLower(strings.TrimSuffix(file.Filename, ext))
		}
	}

	finalEntryPoint := filepath.Join(pluginDir, name, "init.lua")
	testPlugin, err := plugin.NewLuaPlugin(name, finalEntryPoint, h.Engine.Store)
	if err != nil {
		os.RemoveAll(filepath.Join(pluginDir, name))
		return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("Invalid plugin: %v", err)})
	}

	h.Engine.Plugins[name] = testPlugin
	var dbP models.Plugin
	if err := db.DB.WithContext(c.UserContext()).Where("name = ?", name).First(&dbP).Error; err != nil {
		dbP = models.Plugin{Name: name, IsEnabled: true}
		db.DB.WithContext(c.UserContext()).Create(&dbP)
	}

	return c.JSON(fiber.Map{"status": "success", "name": name})
}

func (h *API) DeletePlugin(c *fiber.Ctx) error {
	name := strings.ToLower(c.Params("name"))

	pluginDir := getPluginDir()
	dirPath := filepath.Join(pluginDir, name)
	filePath := filepath.Join(pluginDir, name+".lua")

	if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
		os.RemoveAll(dirPath)
	} else {
		os.Remove(filePath)
	}

	delete(h.Engine.Plugins, name)
	db.DB.WithContext(c.UserContext()).Where("name = ?", name).Delete(&models.Plugin{})
	return c.SendString("Deleted")
}
