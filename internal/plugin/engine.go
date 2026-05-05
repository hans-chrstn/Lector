package plugin

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"

	"github.com/user/lector/internal/models"
	lua "github.com/yuin/gopher-lua"
)

type PluginSource interface {
	Search(query string) ([]models.SearchItem, error)
	GetDocument(url string) (models.Document, error)
	GetChapter(url string) (models.Chapter, error)
	GetPopular(page int) ([]models.SearchItem, error)
	GetLatest(page int) ([]models.SearchItem, error)
}

type LuaPlugin struct {
	L              *lua.LState
	Path           string
	LoadedAt       time.Time
	Client         *http.Client
	Mu             sync.Mutex
	Tabs           []Tab
	Sections       []Section
	SettingsGroups []SettingsGroup
	Actions        []Action
	UIOverrides    map[string]map[string]string
	Permissions    []string
}

type Action struct {
	Context string `json:"context"`
	Label   string `json:"label"`
	Method  string `json:"method"`
	Icon    string `json:"icon"`
}

type Tab struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Icon      string `json:"icon"`
	SectionID string `json:"section_id"`
	Component string `json:"component"`
}

type Section struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type SettingsGroup struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

var GlobalPlugins map[string]*LuaPlugin
var PluginsMu sync.Mutex

func NewLuaPlugin(path string) (*LuaPlugin, error) {
	L := lua.NewState()
	s := &LuaPlugin{
		L:              L,
		Path:           path,
		Tabs:           []Tab{},
		Sections:       []Section{},
		SettingsGroups: []SettingsGroup{},
		Actions:        []Action{},
		UIOverrides:    make(map[string]map[string]string),
		Permissions:    []string{},
	}

	jar, _ := cookiejar.New(nil)
	s.Client = &http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
	}

	s.registerFunctions()

	if err := L.DoFile(path); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *LuaPlugin) Validate() error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	hasSourceFuncs := true
	for _, fn := range []string{"search", "get_document", "get_chapter"} {
		if s.L.GetGlobal(fn).Type() != lua.LTFunction {
			hasSourceFuncs = false
			break
		}
	}

	hasUI := len(s.Tabs) > 0 || len(s.Actions) > 0 || len(s.SettingsGroups) > 0

	if !hasSourceFuncs && !hasUI {
		return fmt.Errorf("plugin is empty: no source functions and no UI elements defined")
	}

	return nil
}
