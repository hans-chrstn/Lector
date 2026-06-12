package plugin

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"

	"strings"
	"sync"
	"time"

	"github.com/user/lector/internal/core/interfaces"
	"github.com/user/lector/internal/models"
	lua "github.com/yuin/gopher-lua"
)

const OfficialPublicKey = "7c9e1e79268f702672a969f69792688972a969f69792688972a969f697926889"

type PluginSource interface {
	Search(query string) ([]models.SearchItem, error)
	GetDirectory(id string, page int) ([]models.SearchItem, error)
	GetDocument(url string) (models.Document, error)
	GetChapter(url string) (models.Chapter, error)
	GetPopular(page int) ([]models.SearchItem, error)
	GetLatest(page int) ([]models.SearchItem, error)
}

type LuaPlugin struct {
	L                  *lua.LState
	Store              interfaces.PluginDataStore
	Name               string
	Path               string
	LoadedAt           time.Time
	Client             *http.Client
	Mu                 sync.Mutex
	ManifestMu         sync.RWMutex
	Tabs               []Tab
	Sections           []Section
	SettingsGroups     []SettingsGroup
	Actions            []Action
	UIOverrides        map[string]map[string]string
	Permissions        []string
	Capabilities       []string
	CSS                string
	IsVerified         bool
	NetworkProfileName string
	Type               string
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

type PluginEngine struct {
	Store   interfaces.PluginDataStore
	Plugins map[string]*LuaPlugin
	Mu      sync.Mutex
}

var GlobalPlugins map[string]*LuaPlugin
var PluginsMu sync.Mutex

func NewLuaPlugin(name, path string, store interfaces.PluginDataStore) (*LuaPlugin, error) {
	if err := ValidateManifest(path); err != nil {
		return nil, err
	}

	L := lua.NewState(lua.Options{
		SkipOpenLibs:        true,
		MinimizeStackMemory: true,
		RegistrySize:        128,
		RegistryMaxSize:     1024,
	})
	L.SetMx(256)

	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
		{lua.MathLibName, lua.OpenMath},
	} {
		if err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(pair.f),
			NRet:    0,
			Protect: true,
		}, lua.LString(pair.n)); err != nil {
			L.Close()
			return nil, err
		}
	}

	L.SetGlobal("package", lua.LNil)
	L.SetGlobal("require", lua.LNil)
	L.SetGlobal("io", lua.LNil)
	L.SetGlobal("debug", lua.LNil)
	L.SetGlobal("os", lua.LNil)
	L.SetGlobal("dofile", lua.LNil)
	L.SetGlobal("loadfile", lua.LNil)

	s := &LuaPlugin{
		L:              L,
		Store:          store,
		Name:           name,
		Path:           path,
		Tabs:           []Tab{},
		Sections:       []Section{},
		SettingsGroups: []SettingsGroup{},
		Actions:        []Action{},
		UIOverrides:    make(map[string]map[string]string),
		Permissions:    []string{},
		Capabilities:   []string{},
		Type:           "source",
	}

	s.IsVerified = s.verifySignature()

	jar, _ := cookiejar.New(nil)
	s.Client = &http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
	}

	s.registerFunctions()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	L.SetContext(ctx)

	loadErr := L.DoFile(path)

	if err := s.Validate(); err != nil {
		L.Close()
		return nil, fmt.Errorf("validation failed: %w (load error: %v)", err, loadErr)
	}

	if loadErr != nil {
		L.Close()
		s.L = nil
	}

	return s, loadErr
}

func (s *LuaPlugin) verifySignature() bool {
	sigPath := s.Path + ".sig"
	sigHex, err := os.ReadFile(sigPath)
	if err != nil {
		return false
	}

	sig, err := hex.DecodeString(strings.TrimSpace(string(sigHex)))
	if err != nil {
		return false
	}

	content, err := os.ReadFile(s.Path)
	if err != nil {
		return false
	}

	pubKey, _ := hex.DecodeString(OfficialPublicKey)
	return ed25519.Verify(pubKey, content, sig)
}

func (s *LuaPlugin) Validate() error {
	if s.Name == "probe" {
		return nil
	}
	s.ManifestMu.RLock()
	capLen := len(s.Capabilities)
	pluginType := s.Type
	s.ManifestMu.RUnlock()

	if capLen == 0 {
		return fmt.Errorf("plugin has no capabilities enabled (use app.enable_capability)")
	}

	validator := GetValidator(pluginType)
	if err := validator.Validate(s); err != nil {
		return fmt.Errorf("[%s validation failed] %w", pluginType, err)
	}

	return nil
}
