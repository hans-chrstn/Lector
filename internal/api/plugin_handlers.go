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
		fmt.Printf("[Plugin] Failed to create dir %s: %v\n", pluginDir, err)
		return c.Status(500).SendString("Failed to create plugins directory")
	}

	path := filepath.Join(pluginDir, file.Filename)
	fmt.Printf("[Plugin] Saving uploaded plugin to: %s\n", path)
	if err := c.SaveFile(file, path); err != nil {
		fmt.Printf("[Plugin] Failed to save file %s: %v\n", path, err)
		return c.Status(500).SendString("Failed to save file")
	}

	testPlugin, err := plugin.NewLuaPlugin(path)
	if err != nil || testPlugin.Validate() != nil {
		os.Remove(path)
		fmt.Printf("[Plugin] Invalid plugin %s: %v\n", file.Filename, err)
		return c.Status(400).SendString(fmt.Sprintf("Invalid plugin: %v", err))
	}

	name := strings.TrimSuffix(file.Filename, ".lua")
	h.Plugins[name] = testPlugin

	var p models.Plugin
	if err := db.DB.Where("name = ?", name).First(&p).Error; err != nil {
		p = models.Plugin{Name: name, IsEnabled: true}
		db.DB.Create(&p)
	}

	fmt.Printf("[Plugin] Successfully installed and loaded: %s\n", name)
	return c.JSON(fiber.Map{"status": "success", "name": name})
}

func (h *API) GetPluginsManifest(c *fiber.Ctx) error {
	type PluginManifest struct {
		Name           string                       `json:"name"`
		IsEnabled      bool                         `json:"is_enabled"`
		IsLoaded       bool                         `json:"is_loaded"`
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
		if filepath.Ext(file.Name()) == ".lua" {
			name := strings.TrimSuffix(file.Name(), ".lua")
			var p models.Plugin
			if err := db.DB.Where("name = ?", name).First(&p).Error; err != nil {
				p = models.Plugin{Name: name, IsEnabled: true}
				db.DB.Create(&p)
			}
		}
	}

	var dbPlugins []models.Plugin
	db.DB.Order("priority ASC, name ASC").Find(&dbPlugins)

	manifests := []PluginManifest{}
	for _, p := range dbPlugins {
		path := filepath.Join(pluginDir, p.Name+".lua")
		info, err := os.Stat(path)

		if os.IsNotExist(err) {
			if _, exists := h.Plugins[p.Name]; exists {
				delete(h.Plugins, p.Name)
			}
			continue
		}

		if p.IsEnabled {
			s, exists := h.Plugins[p.Name]
			if !exists || (info != nil && info.ModTime().After(s.LoadedAt)) {
				if newS, err := plugin.NewLuaPlugin(path); err == nil {
					newS.LoadedAt = info.ModTime()
					h.Plugins[p.Name] = newS
					fmt.Printf("[Plugin] %s: reloaded from disk (ModTime: %s)\n", p.Name, newS.LoadedAt.Format("15:04:05"))
				} else {
					fmt.Printf("[Plugin] %s: reload failed: %v\n", p.Name, err)
				}
			}
		} else {
			delete(h.Plugins, p.Name)
		}

		m := PluginManifest{
			Name:      p.Name,
			IsEnabled: p.IsEnabled,
			IsLoaded:  h.Plugins[p.Name] != nil,
		}

		if s, ok := h.Plugins[p.Name]; ok {
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

	fmt.Printf("[Plugins] Reordering plugins: %v\n", req)
	for i, name := range req {
		db.DB.Model(&models.Plugin{}).Where("name = ?", name).Update("priority", i)
	}

	return c.SendString("Reordered")
}

func (h *API) TogglePlugin(c *fiber.Ctx) error {
	name := c.Params("name")
	var p models.Plugin
	if err := db.DB.Where("name = ?", name).First(&p).Error; err != nil {
		return c.Status(404).SendString("Plugin not found")
	}

	p.IsEnabled = !p.IsEnabled
	db.DB.Save(&p)

	if p.IsEnabled {
		path := filepath.Join(getPluginDir(), name+".lua")
		s, err := plugin.NewLuaPlugin(path)
		if err == nil {
			h.Plugins[name] = s
		}
	} else {
		delete(h.Plugins, name)
	}

	return c.JSON(p)
}

func (h *API) DeletePlugin(c *fiber.Ctx) error {
	name := c.Params("name")

	path := filepath.Join(getPluginDir(), name+".lua")
	if err := os.Remove(path); err != nil {
		return c.Status(500).SendString("Failed to delete file")
	}

	delete(h.Plugins, name)
	return c.SendString("Deleted")
}

func (h *API) PluginRPC(c *fiber.Ctx) error {
	name := c.Params("name")
	method := c.Params("method")

	p, ok := h.Plugins[name]
	if !ok {
		fmt.Printf("[PluginRPC] Error: Plugin %s not found. Active: %v\n", name, h.GetActivePluginNames())
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

		fmt.Printf("[PluginRPC] Error: Method %s not found in plugin %s\n", method, name)
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
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Printf("[Security] [%s] RPC execution timed out in %s\n", name, method)
		}
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
