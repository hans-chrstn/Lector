package api

import (
	"context"
	"fmt"
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
		return c.Status(400).SendString("No file uploaded")
	}

	if !strings.HasSuffix(file.Filename, ".lua") {
		return c.Status(400).SendString("Only .lua files are allowed")
	}

	pluginDir := getPluginDir()
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return c.Status(500).SendString("Failed to create plugins directory")
	}

	name := strings.ToLower(strings.TrimSuffix(file.Filename, ".lua"))
	path := filepath.Join(pluginDir, file.Filename)

	if err := c.SaveFile(file, path); err != nil {
		return c.Status(500).SendString("Failed to save file")
	}

	testPlugin, err := plugin.NewLuaPlugin(name, path)
	if err != nil || testPlugin.Validate() != nil {
		os.Remove(path)
		return c.Status(400).SendString(fmt.Sprintf("Invalid plugin: %v", err))
	}

	h.Plugins[name] = testPlugin

	var p models.Plugin
	if err := db.DB.Where("name = ?", name).First(&p).Error; err != nil {
		p = models.Plugin{Name: name, IsEnabled: true}
		db.DB.Create(&p)
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
		if err := db.DB.Where("name = ?", name).First(&p).Error; err != nil {
			p = models.Plugin{Name: name, IsEnabled: true}
			db.DB.Create(&p)
		}
	}

	var dbPlugins []models.Plugin
	db.DB.Order("priority ASC, name ASC").Find(&dbPlugins)

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
			if _, exists := h.Plugins[name]; exists {
				delete(h.Plugins, name)
			}
			continue
		}

		if p.IsEnabled {
			s, exists := h.Plugins[name]
			if !exists || (info != nil && info.ModTime().After(s.LoadedAt)) {
				if newS, err := plugin.NewLuaPlugin(name, sPath); err == nil {
					newS.LoadedAt = info.ModTime()
					h.Plugins[name] = newS
				}
			}
		} else {
			delete(h.Plugins, name)
		}

		m := PluginManifest{
			Name:      name,
			IsEnabled: p.IsEnabled,
			IsLoaded:  h.Plugins[name] != nil,
		}

		if s, ok := h.Plugins[name]; ok {
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
	db.DB.Order("priority ASC, name ASC").Find(&plugins)
	return c.JSON(plugins)
}

func (h *API) ReorderPlugins(c *fiber.Ctx) error {
	var req []string
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).SendString("Invalid request")
	}

	for i, name := range req {
		db.DB.Model(&models.Plugin{}).Where("name = ?", strings.ToLower(name)).Update("priority", i)
	}

	return c.SendString("Reordered")
}

func (h *API) TogglePlugin(c *fiber.Ctx) error {
	name := strings.ToLower(c.Params("name"))
	var p models.Plugin
	if err := db.DB.Where("name = ?", name).First(&p).Error; err != nil {
		return c.Status(404).SendString("Plugin not found")
	}

	p.IsEnabled = !p.IsEnabled
	db.DB.Save(&p)

	if p.IsEnabled {
		pluginDir := getPluginDir()
		var sPath string
		if _, err := os.Stat(filepath.Join(pluginDir, name, "init.lua")); err == nil {
			sPath = filepath.Join(pluginDir, name, "init.lua")
		} else {
			sPath = filepath.Join(pluginDir, name+".lua")
		}

		s, err := plugin.NewLuaPlugin(name, sPath)
		if err == nil {
			h.Plugins[name] = s
		}
	} else {
		delete(h.Plugins, name)
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

	delete(h.Plugins, name)
	db.DB.Where("name = ?", name).Delete(&models.Plugin{})
	return c.SendString("Deleted")
}

func (h *API) PluginRPC(c *fiber.Ctx) error {
	name := strings.ToLower(c.Params("name"))
	method := c.Params("method")

	p, ok := h.Plugins[name]
	if !ok {
		return c.Status(404).SendString("Plugin not found")
	}

	p.Mu.Lock()
	defer p.Mu.Unlock()

	fn := p.L.GetGlobal(method)
	if fn.Type() != lua.LTFunction {
		globals := p.L.Get(lua.GlobalsIndex).(*lua.LTable)
		globals.ForEach(func(k, v lua.LValue) {
			if strings.EqualFold(k.String(), method) && v.Type() == lua.LTFunction {
				fn = v
			}
		})
	}

	if fn.Type() != lua.LTFunction {
		if method == "get_document_actions" {
			return c.JSON([]interface{}{})
		}
		return c.Status(404).SendString(fmt.Sprintf("RPC method %s not found in plugin %s", method, name))
	}

	body := string(c.Body())
	if body == "" {
		body = "{}"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	p.L.SetContext(ctx)

	if err := p.L.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true}, lua.LString(body)); err != nil {
		return c.Status(500).SendString(fmt.Sprintf("RPC error: %v", err))
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
