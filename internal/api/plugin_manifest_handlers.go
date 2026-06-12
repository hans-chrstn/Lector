package api

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/plugin"
)

func (h *API) GetPluginsManifest(c *fiber.Ctx) error {
	if h.Engine == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Plugin engine not initialized"})
	}

	eng := h.Engine

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
			info, err = os.Stat(sPath)
			if err != nil {
				sPath = ""
				info = nil
			}
		} else if fileInfo, err := os.Stat(filepath.Join(pluginDir, name+".lua")); err == nil {
			sPath = filepath.Join(pluginDir, name+".lua")
			info = fileInfo
		}

		if sPath == "" || info == nil {
			eng.Mu.Lock()
			if eng.Plugins != nil {
				delete(eng.Plugins, name)
			}
			eng.Mu.Unlock()
			continue
		}

		if p.IsEnabled {
			eng.Mu.Lock()
			if eng.Plugins == nil {
				eng.Plugins = make(map[string]*plugin.LuaPlugin)
			}
			s, exists := eng.Plugins[name]
			eng.Mu.Unlock()

			if !exists || (s == nil || info.ModTime().After(s.LoadedAt)) {
				modTime := info.ModTime()
				go func(n, path string, mTime time.Time) {
					newS, _ := plugin.NewLuaPlugin(n, path, eng.Store)
					if newS != nil {
						newS.LoadedAt = mTime
						eng.Mu.Lock()
						if eng.Plugins == nil {
							eng.Plugins = make(map[string]*plugin.LuaPlugin)
						}
						eng.Plugins[n] = newS
						eng.Mu.Unlock()
					}
				}(name, sPath, modTime)
			}
		} else {
			eng.Mu.Lock()
			if eng.Plugins != nil {
				delete(eng.Plugins, name)
			}
			eng.Mu.Unlock()
		}

		var s *plugin.LuaPlugin
		var isLoaded bool

		eng.Mu.Lock()
		if eng.Plugins != nil {
			s = eng.Plugins[name]
		}
		isLoaded = s != nil
		eng.Mu.Unlock()

		m := PluginManifest{
			Name:           name,
			IsEnabled:      p.IsEnabled,
			IsLoaded:       isLoaded,
			Tabs:           []plugin.Tab{},
			Sections:       []plugin.Section{},
			SettingsGroups: []plugin.SettingsGroup{},
			Actions:        []plugin.Action{},
			UIOverrides:    make(map[string]map[string]string),
			Permissions:    []string{},
			Capabilities:   []string{},
		}

		if isLoaded && s != nil {
			m.IsVerified = s.IsVerified
			m.Tabs = s.Tabs
			if m.Tabs == nil {
				m.Tabs = []plugin.Tab{}
			}
			m.Sections = s.Sections
			if m.Sections == nil {
				m.Sections = []plugin.Section{}
			}
			m.SettingsGroups = s.SettingsGroups
			if m.SettingsGroups == nil {
				m.SettingsGroups = []plugin.SettingsGroup{}
			}
			m.Actions = s.Actions
			if m.Actions == nil {
				m.Actions = []plugin.Action{}
			}
			m.UIOverrides = s.UIOverrides
			if m.UIOverrides == nil {
				m.UIOverrides = make(map[string]map[string]string)
			}
			m.Permissions = s.Permissions
			if m.Permissions == nil {
				m.Permissions = []string{}
			}
			m.Capabilities = s.Capabilities
			if m.Capabilities == nil {
				m.Capabilities = []string{}
			}
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

		s, _ := plugin.NewLuaPlugin(name, sPath, h.Engine.Store)
		if s != nil {
			h.Engine.Mu.Lock()
			if h.Engine.Plugins == nil {
				h.Engine.Plugins = make(map[string]*plugin.LuaPlugin)
			}
			h.Engine.Plugins[name] = s
			h.Engine.Mu.Unlock()
		}
	} else {
		h.Engine.Mu.Lock()
		if h.Engine.Plugins != nil {
			delete(h.Engine.Plugins, name)
		}
		h.Engine.Mu.Unlock()
	}

	return c.JSON(p)
}
