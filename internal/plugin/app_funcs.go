package plugin

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	lua "github.com/yuin/gopher-lua"
)

func (s *LuaPlugin) registerAppFunctions() {
	app := s.L.NewTable()
	s.L.SetField(app, "add_tab", s.L.NewFunction(s.addTab))
	s.L.SetField(app, "add_section", s.L.NewFunction(s.addSection))
	s.L.SetField(app, "add_settings_group", s.L.NewFunction(s.addSettingsGroup))
	s.L.SetField(app, "add_permission", s.L.NewFunction(s.addPermission))
	s.L.SetField(app, "add_action", s.L.NewFunction(s.addAction))
	s.L.SetField(app, "rpc", s.L.NewFunction(s.appRPC))
	s.L.SetField(app, "log", s.L.NewFunction(s.appLog))
	s.L.SetField(app, "spawn", s.L.NewFunction(s.appSpawn))
	s.L.SetField(app, "sleep", s.L.NewFunction(s.appSleep))

	ui := s.L.NewTable()
	s.L.SetField(ui, "set_override", s.L.NewFunction(s.uiSetOverride))
	s.L.SetField(app, "ui", ui)

	store := s.L.NewTable()
	s.L.SetField(store, "set", s.L.NewFunction(s.storeSet))
	s.L.SetField(store, "get", s.L.NewFunction(s.storeGet))
	s.L.SetField(app, "store", store)
	s.L.SetGlobal("app", app)
}

func (s *LuaPlugin) addPermission(L *lua.LState) int {
	domain := L.CheckString(1)
	s.Mu.Lock()
	s.Permissions = append(s.Permissions, domain)
	s.Mu.Unlock()
	return 0
}

func (s *LuaPlugin) addAction(L *lua.LState) int {
	context := L.CheckString(1)
	label := L.CheckString(2)
	method := L.CheckString(3)
	icon := L.OptString(4, "Zap")
	s.Mu.Lock()
	s.Actions = append(s.Actions, Action{Context: context, Label: label, Method: method, Icon: icon})
	s.Mu.Unlock()
	return 0
}

func (s *LuaPlugin) uiSetOverride(L *lua.LState) int {
	key := L.CheckString(1)
	tbl := L.CheckTable(2)

	override := make(map[string]string)
	tbl.ForEach(func(k, v lua.LValue) {
		override[k.String()] = v.String()
	})

	s.Mu.Lock()
	s.UIOverrides[key] = override
	s.Mu.Unlock()
	return 0
}

func (s *LuaPlugin) appRPC(L *lua.LState) int {
	target := L.CheckString(1)
	method := L.CheckString(2)
	args := L.OptString(3, "{}")

	PluginsMu.Lock()
	p, ok := GlobalPlugins[target]
	PluginsMu.Unlock()

	if !ok {
		L.Push(lua.LNil)
		L.Push(lua.LString("Plugin not found"))
		return 2
	}

	p.Mu.Lock()
	defer p.Mu.Unlock()
	fn := p.L.GetGlobal(method)
	if fn.Type() != lua.LTFunction {
		L.Push(lua.LNil)
		L.Push(lua.LString("Method not found"))
		return 2
	}

	if err := p.L.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true}, lua.LString(args)); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	ret := p.L.Get(-1)
	p.L.Pop(1)
	L.Push(ret)
	return 1
}

func (s *LuaPlugin) appLog(L *lua.LState) int {
	msg := L.CheckString(1)
	level := L.OptString(2, "INFO")
	name := strings.TrimSuffix(filepath.Base(s.Path), ".lua")
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("%s [%s] [%s] %s\n", timestamp, strings.ToUpper(level), name, msg)
	return 0
}

func (s *LuaPlugin) addTab(L *lua.LState) int {
	id := L.CheckString(1)
	label := L.CheckString(2)
	icon := L.OptString(3, "Compass")
	sectionID := L.OptString(4, "")
	component := L.OptString(5, "")
	s.Mu.Lock()
	s.Tabs = append(s.Tabs, Tab{ID: id, Label: label, Icon: icon, SectionID: sectionID, Component: component})
	s.Mu.Unlock()
	return 0
}

func (s *LuaPlugin) addSection(L *lua.LState) int {
	id := L.CheckString(1)
	label := L.CheckString(2)
	s.Mu.Lock()
	s.Sections = append(s.Sections, Section{ID: id, Label: label})
	s.Mu.Unlock()
	return 0
}

func (s *LuaPlugin) addSettingsGroup(L *lua.LState) int {
	id := L.CheckString(1)
	label := L.CheckString(2)
	s.Mu.Lock()
	s.SettingsGroups = append(s.SettingsGroups, SettingsGroup{ID: id, Label: label})
	s.Mu.Unlock()
	return 0
}

func (s *LuaPlugin) appSpawn(L *lua.LState) int {
	funcName := L.CheckString(1)
	argsStr := L.OptString(2, "{}")

	go func() {
		newL, err := NewLuaPlugin(s.Path)
		if err != nil {
			fmt.Printf("[Plugin] Spawn error: Failed to initialize isolated VM: %v\n", err)
			return
		}
		defer newL.L.Close()

		fn := newL.L.GetGlobal(funcName)
		if fn.Type() != lua.LTFunction {
			fmt.Printf("[Plugin] Spawn error: Function %s not found\n", funcName)
			return
		}

		if err := newL.L.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, lua.LString(argsStr)); err != nil {
			fmt.Printf("[Plugin] Spawn error in %s: %v\n", funcName, err)
		}
	}()

	return 0
}

func (s *LuaPlugin) appSleep(L *lua.LState) int {
	ms := L.CheckInt(1)
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return 0
}

func (s *LuaPlugin) storeSet(L *lua.LState) int {
	key := L.CheckString(1)
	val := L.CheckString(2)
	name := strings.TrimSuffix(filepath.Base(s.Path), ".lua")
	fullKey := fmt.Sprintf("plugin_%s_%s", name, key)

	db.DB.Save(&models.CacheItem{
		Key:   fullKey,
		Value: []byte(val),
	})
	return 0
}

func (s *LuaPlugin) storeGet(L *lua.LState) int {
	key := L.CheckString(1)
	name := strings.TrimSuffix(filepath.Base(s.Path), ".lua")
	fullKey := fmt.Sprintf("plugin_%s_%s", name, key)

	var item models.CacheItem
	if err := db.DB.Where("key = ?", fullKey).First(&item).Error; err == nil {
		L.Push(lua.LString(string(item.Value)))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}
