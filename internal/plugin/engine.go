package plugin

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"runtime"
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

func MonitorMemory(ctx context.Context, cancel context.CancelFunc, limit int64) context.CancelFunc {
	memCtx, memCancel := context.WithCancel(ctx)
	go func() {
		var initialStats runtime.MemStats
		runtime.ReadMemStats(&initialStats)
		initialAlloc := initialStats.Alloc

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-memCtx.Done():
				return
			case <-ticker.C:
				var currentStats runtime.MemStats
				runtime.ReadMemStats(&currentStats)
				used := int64(currentStats.Alloc) - int64(initialAlloc)
				if used > limit {
					cancel()
					return
				}
			}
		}
	}()
	return memCancel
}

func NewLuaPlugin(name, path string, store interfaces.PluginDataStore) (*LuaPlugin, error) {
	L := lua.NewState(lua.Options{
		SkipOpenLibs:        true,
		MinimizeStackMemory: true,
		RegistrySize:        128,
		RegistryMaxSize:     1024,
	})

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
	}

	s.IsVerified = s.verifySignature()

	jar, _ := cookiejar.New(nil)
	s.Client = &http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
	}

	s.registerFunctions()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	L.SetContext(ctx)

	stopMonitor := MonitorMemory(ctx, cancel, 50*1024*1024)
	defer stopMonitor()

	if err := L.DoFile(path); err != nil {
		return nil, err
	}

	if err := s.Validate(); err != nil {
		return nil, err
	}

	return s, nil
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
	s.ManifestMu.RUnlock()

	if capLen == 0 {
		return fmt.Errorf("plugin has no capabilities enabled (use app.enable_capability)")
	}

	hasSourceFuncs := true
	for _, fn := range []string{"search", "get_document", "get_chapter"} {
		s.Mu.Lock()
		f := s.L.GetGlobal(fn)
		s.Mu.Unlock()
		if f.Type() != lua.LTFunction {
			hasSourceFuncs = false
			break
		}
	}

	s.ManifestMu.RLock()
	hasUI := len(s.Tabs) > 0 || len(s.Actions) > 0 || len(s.SettingsGroups) > 0 || len(s.Sections) > 0 || len(s.UIOverrides) > 0
	s.ManifestMu.RUnlock()

	if !hasSourceFuncs && !hasUI {
		return fmt.Errorf("plugin is empty: no source functions and no UI elements defined")
	}

	return nil
}
