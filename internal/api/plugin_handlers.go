package api

import (
	"archive/zip"
	"context"
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
	lua "github.com/yuin/gopher-lua"
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
			baseName := strings.ToLower(strings.TrimSuffix(file.Filename, ext))
			if baseName == "init" {
				return c.Status(409).JSON(fiber.Map{"error": "Generic filename detected. Please provide a plugin name."})
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
			destDir = filepath.Join(pluginDir, strings.TrimSuffix(file.Filename, ext))
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
func (h *API) GetPluginsManifest(c *fiber.Ctx) error {
	type PluginManifest struct {
		Name           string                       `json:"name"`
		IsEnabled      bool                         `json:"is_enabled"`
		IsLoaded       bool                         `json:"is_loaded"`
		IsVerified     bool                         `json:"is_verified"`
		Tabs           []plugin.Tab                 `json:"tabs"`
		Sections       []plugin.Section             `json:"sections"`
		SettingsGroups []plugin.SettingsGroup       `json:"settings_groups"`
		Actions        []plugin.Action              `json:"actions"`
		UIOverrides    map[string]map[string]string `json:"ui_overrides"`
		Permissions    []string                     `json:"permissions"`
		Capabilities   []string                     `json:"capabilities"`
		CSS            string                       `json:"css"`
	}

	pluginDir := getPluginDir()
	os.MkdirAll(pluginDir, 0755)

	files, _ := os.ReadDir(pluginDir)
	for _, file := range files {
		var name string
		if file.IsDir() {
			name = strings.ToLower(file.Name())
			if _, err := os.Stat(filepath.Join(pluginDir, file.Name(), "init.lua")); os.IsNotExist(err) {
				continue
			}
		} else if filepath.Ext(file.Name()) == ".lua" {
			name = strings.ToLower(strings.TrimSuffix(file.Name(), ".lua"))
		} else {
			continue
		}

		var p models.Plugin
		if err := db.DB.WithContext(c.UserContext()).Where("name = ?", name).First(&p).Error; err != nil {
			p = models.Plugin{Name: name, IsEnabled: true}
			db.DB.WithContext(c.UserContext()).Create(&p)
		}
	}

	var dbPlugins []models.Plugin
	db.DB.WithContext(c.UserContext()).Order("priority ASC, name ASC").Find(&dbPlugins)

	manifests := []PluginManifest{}
	for _, p := range dbPlugins {
		name := strings.ToLower(p.Name)
		var sPath string
		var info os.FileInfo

		if dirInfo, err := os.Stat(filepath.Join(pluginDir, name)); err == nil && dirInfo.IsDir() {
			sPath = filepath.Join(pluginDir, name, "init.lua")
			info, _ = os.Stat(sPath)
		} else if fileInfo, err := os.Stat(filepath.Join(pluginDir, name+".lua")); err == nil {
			sPath = filepath.Join(pluginDir, name+".lua")
			info = fileInfo
		}

		if sPath == "" {
			if _, exists := h.Engine.Plugins[name]; exists {
				delete(h.Engine.Plugins, name)
			}
			continue
		}

		if p.IsEnabled {
			s, exists := h.Engine.Plugins[name]
			if !exists || (info != nil && info.ModTime().After(s.LoadedAt)) {
				if newS, err := plugin.NewLuaPlugin(name, sPath, h.Engine.Store); err == nil {
					newS.LoadedAt = info.ModTime()
					h.Engine.Plugins[name] = newS
				}
			}
		} else {
			delete(h.Engine.Plugins, name)
		}

		m := PluginManifest{
			Name:      name,
			IsEnabled: p.IsEnabled,
			IsLoaded:  h.Engine.Plugins[name] != nil,
		}

		if s, ok := h.Engine.Plugins[name]; ok {
			m.IsVerified = s.IsVerified
			m.Tabs = s.Tabs
			m.Sections = s.Sections
			m.SettingsGroups = s.SettingsGroups
			m.Actions = s.Actions
			m.UIOverrides = s.UIOverrides
			m.Permissions = s.Permissions
			m.Capabilities = s.Capabilities
			m.CSS = s.CSS
		}
		manifests = append(manifests, m)
	}

	return c.JSON(manifests)
}

func (h *API) GetPlugins(c *fiber.Ctx) error {
	var plugins []models.Plugin
	db.DB.WithContext(c.UserContext()).Order("priority ASC, name ASC").Find(&plugins)
	return c.JSON(plugins)
}

func (h *API) ReorderPlugins(c *fiber.Ctx) error {
	var req []string
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	for i, name := range req {
		db.DB.WithContext(c.UserContext()).Model(&models.Plugin{}).Where("name = ?", strings.ToLower(name)).Update("priority", i)
	}

	return c.SendString("Reordered")
}

func (h *API) TogglePlugin(c *fiber.Ctx) error {
	name := strings.ToLower(c.Params("name"))
	var p models.Plugin
	if err := db.DB.WithContext(c.UserContext()).Where("name = ?", name).First(&p).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Plugin not found"})
	}

	p.IsEnabled = !p.IsEnabled
	db.DB.WithContext(c.UserContext()).Save(&p)

	if p.IsEnabled {
		pluginDir := getPluginDir()
		var sPath string
		if _, err := os.Stat(filepath.Join(pluginDir, name, "init.lua")); err == nil {
			sPath = filepath.Join(pluginDir, name, "init.lua")
		} else {
			sPath = filepath.Join(pluginDir, name+".lua")
		}

		s, err := plugin.NewLuaPlugin(name, sPath, h.Engine.Store)
		if err == nil {
			h.Engine.Plugins[name] = s
		}
	} else {
		delete(h.Engine.Plugins, name)
	}

	return c.JSON(p)
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

func (h *API) PluginRPC(c *fiber.Ctx) error {
	name := strings.ToLower(c.Params("name"))
	method := c.Params("method")

	p, ok := h.Engine.Plugins[name]
	if !ok {
		return c.Status(404).JSON(fiber.Map{"error": "Plugin not found"})
	}

	p.Mu.Lock()
	defer p.Mu.Unlock()

	exports := p.L.GetGlobal("exports")
	if exports.Type() != lua.LTTable {
		if method == "get_document_actions" {
			return c.JSON([]interface{}{})
		}
		return c.Status(403).JSON(fiber.Map{"error": fmt.Sprintf("Plugin %s does not export any functions (exports table not found)", name)})
	}

	fn := p.L.GetField(exports, method)

	if fn.Type() != lua.LTFunction {
		if method == "get_document_actions" {
			return c.JSON([]interface{}{})
		}
		return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("RPC method %s not found in plugin %s", method, name)})
	}

	body := string(c.Body())
	if body == "" {
		body = "{}"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	p.L.SetContext(ctx)

	if err := p.L.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true}, lua.LString(body)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("RPC error: %v", err)})
	}

	ret := p.L.Get(-1)
	p.L.Pop(1)

	if str, ok := ret.(lua.LString); ok {
		c.Set("Content-Type", "application/json")
		return c.SendString(string(str))
	}

	if tbl, ok := ret.(*lua.LTable); ok {
		data := tableToMap(tbl)
		return c.JSON(data)
	}

	return c.JSON(fiber.Map{"status": "success", "info": "RPC executed"})
}

func tableToMap(tbl *lua.LTable) interface{} {
	if tbl.MaxN() > 0 {
		arr := []interface{}{}
		tbl.ForEach(func(k, v lua.LValue) {
			arr = append(arr, luaValueToInterface(v))
		})
		return arr
	}
	res := make(map[string]interface{})
	tbl.ForEach(func(k, v lua.LValue) {
		res[k.String()] = luaValueToInterface(v)
	})
	return res
}

func luaValueToInterface(v lua.LValue) interface{} {
	switch v.Type() {
	case lua.LTString:
		return v.String()
	case lua.LTNumber:
		return float64(v.(lua.LNumber))
	case lua.LTBool:
		return bool(v.(lua.LBool))
	case lua.LTTable:
		return tableToMap(v.(*lua.LTable))
	default:
		return nil
	}
}

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
	assetPath := filepath.Join(pluginDir, name, "assets", cleanPath)

	if _, err := os.Stat(assetPath); os.IsNotExist(err) {
		return c.Status(404).SendString("Asset not found")
	}

	return c.SendFile(assetPath)
}
