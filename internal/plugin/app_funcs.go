package plugin

import (
	"context"
	"fmt"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func (s *LuaPlugin) registerAppFunctions() {
	app := s.L.NewTable()
	s.L.SetField(app, "register_manifest", s.L.NewFunction(s.registerManifest))
	s.L.SetField(app, "enable_capability", s.L.NewFunction(s.enableCapability))
	s.L.SetField(app, "add_tab", s.L.NewFunction(s.addTab))
	s.L.SetField(app, "add_section", s.L.NewFunction(s.addSection))
	s.L.SetField(app, "add_settings_group", s.L.NewFunction(s.addSettingsGroup))
	s.L.SetField(app, "add_permission", s.L.NewFunction(s.addPermission))
	s.L.SetField(app, "add_action", s.L.NewFunction(s.addAction))
	s.L.SetField(app, "set_id", s.L.NewFunction(s.setID))
	s.L.SetField(app, "rpc", s.L.NewFunction(s.appRPC))
	s.L.SetField(app, "log", s.L.NewFunction(s.appLog))
	s.L.SetField(app, "spawn", s.L.NewFunction(s.appSpawn))
	s.L.SetField(app, "sleep", s.L.NewFunction(s.appSleep))

	ui := s.L.NewTable()
	s.L.SetField(ui, "set_override", s.L.NewFunction(s.uiSetOverride))
	s.L.SetField(ui, "add_style", s.L.NewFunction(s.uiAddStyle))
	s.L.SetField(ui, "open_stream", s.L.NewFunction(s.uiOpenStream))
	s.L.SetField(ui, "open_gallery", s.L.NewFunction(s.uiOpenGallery))
	s.L.SetField(app, "ui", ui)

	store := s.L.NewTable()
	s.L.SetField(store, "set", s.L.NewFunction(s.storeSet))
	s.L.SetField(store, "get", s.L.NewFunction(s.storeGet))
	s.L.SetField(app, "store", store)
	net := s.L.NewTable()
	s.L.SetField(net, "request", s.L.NewFunction(s.netRequest))
	s.L.SetField(net, "fetch_retry", s.L.NewFunction(s.netRequestRetry))
	s.L.SetField(net, "set_profile", s.L.NewFunction(s.netSetProfile))
	s.L.SetField(app, "net", net)

	s.L.SetGlobal("app", app)
}

func (s *LuaPlugin) uiOpenStream(L *lua.LState) int {
	if !s.HasCapability("ui") {
		return 0
	}
	url := L.CheckString(1)
	_ = L.OptTable(2, L.NewTable())
	s.appLog(L)
	fmt.Printf("[UI] Open Stream requested for %s\n", url)
	return 0
}

func (s *LuaPlugin) uiOpenGallery(L *lua.LState) int {
	if !s.HasCapability("ui") {
		return 0
	}
	images := L.CheckTable(1)
	s.appLog(L)
	fmt.Printf("[UI] Open Gallery requested for %d images\n", images.Len())
	return 0
}

func (s *LuaPlugin) netRequest(L *lua.LState) int {
	if !s.HasCapability("network") {
		L.Push(lua.LNil)
		L.Push(lua.LString("Capability 'network' not enabled"))
		return 2
	}
	method := L.CheckString(1)
	u := L.CheckString(2)
	options := L.OptTable(3, L.NewTable())

	body := ""
	if b := options.RawGetString("body"); b.Type() == lua.LTString {
		body = b.String()
	}

	referer := ""
	if r := options.RawGetString("referer"); r.Type() == lua.LTString {
		referer = r.String()
	}

	isAjax := false
	if a := options.RawGetString("is_ajax"); a.Type() == lua.LTBool {
		isAjax = bool(a.(lua.LBool))
	}

	headers := make(map[string]string)
	if h := options.RawGetString("headers"); h.Type() == lua.LTTable {
		tbl := h.(*lua.LTable)
		tbl.ForEach(func(k, v lua.LValue) {
			headers[k.String()] = v.String()
		})
	}

	res := s.Fetch(method, u, body, referer, isAjax, headers)
	L.Push(lua.LString(res))
	return 1
}

func (s *LuaPlugin) netRequestRetry(L *lua.LState) int {
	if !s.HasCapability("network") {
		L.Push(lua.LNil)
		L.Push(lua.LString("Capability 'network' not enabled"))
		return 2
	}
	method := L.CheckString(1)
	u := L.CheckString(2)
	options := L.OptTable(3, L.NewTable())

	body := ""
	if b := options.RawGetString("body"); b.Type() == lua.LTString {
		body = b.String()
	}

	referer := ""
	if r := options.RawGetString("referer"); r.Type() == lua.LTString {
		referer = r.String()
	}

	isAjax := false
	if a := options.RawGetString("is_ajax"); a.Type() == lua.LTBool {
		isAjax = bool(a.(lua.LBool))
	}

	headers := make(map[string]string)
	if h := options.RawGetString("headers"); h.Type() == lua.LTTable {
		tbl := h.(*lua.LTable)
		tbl.ForEach(func(k, v lua.LValue) {
			headers[k.String()] = v.String()
		})
	}

	var res string
	backoff := 500 * time.Millisecond
	for i := 0; i < 3; i++ {
		res = s.Fetch(method, u, body, referer, isAjax, headers)
		if !strings.Contains(res, "ERROR:") {
			break
		}
		time.Sleep(backoff)
		backoff *= 2
	}

	L.Push(lua.LString(res))
	return 1
}

func (s *LuaPlugin) netSetProfile(L *lua.LState) int {
	if !s.HasCapability("network") {
		return 0
	}
	profile := L.CheckString(1)
	s.Mu.Lock()
	s.NetworkProfileName = profile
	s.Mu.Unlock()
	return 0
}

func (s *LuaPlugin) registerManifest(L *lua.LState) int {
	manifest := L.CheckTable(1)
	s.ManifestMu.Lock()
	defer s.ManifestMu.Unlock()

	typ := manifest.RawGetString("type")
	if typ.Type() == lua.LTString {
		s.Type = typ.String()
	}

	caps := manifest.RawGetString("capabilities")
	if caps.Type() == lua.LTTable {
		caps.(*lua.LTable).ForEach(func(k, v lua.LValue) {
			if v.Type() == lua.LTString {
				capStr := v.String()
				exists := false
				for _, c := range s.Capabilities {
					if c == capStr {
						exists = true
						break
					}
				}
				if !exists {
					s.Capabilities = append(s.Capabilities, capStr)
				}
			}
		})
	}
	return 0
}

func (s *LuaPlugin) enableCapability(L *lua.LState) int {
	capName := L.CheckString(1)
	s.ManifestMu.Lock()
	defer s.ManifestMu.Unlock()
	found := false
	for _, c := range s.Capabilities {
		if c == capName {
			found = true
			break
		}
	}
	if !found {
		s.Capabilities = append(s.Capabilities, capName)
	}
	return 0
}

func (s *LuaPlugin) HasCapability(name string) bool {
	s.ManifestMu.RLock()
	defer s.ManifestMu.RUnlock()
	for _, c := range s.Capabilities {
		if c == name {
			return true
		}
	}
	return false
}

func (s *LuaPlugin) addPermission(L *lua.LState) int {
	name := s.Name
	if !s.HasCapability("network") {
		L.RaiseError("[Security] [%s] Blocked %s (Capability '%s' not enabled)", name, "add_permission", "network")
		return 0
	}
	domain := L.CheckString(1)
	s.ManifestMu.Lock()
	s.Permissions = append(s.Permissions, domain)
	s.ManifestMu.Unlock()
	return 0
}

func (s *LuaPlugin) addAction(L *lua.LState) int {
	name := s.Name
	if !s.HasCapability("ui") {
		L.RaiseError("[Security] [%s] Blocked %s (Capability '%s' not enabled)", name, "add_action", "ui")
		return 0
	}
	context := L.CheckString(1)
	label := L.CheckString(2)
	method := L.CheckString(3)
	icon := L.OptString(4, "Zap")
	s.ManifestMu.Lock()
	s.Actions = append(s.Actions, Action{Context: context, Label: label, Method: method, Icon: icon})
	s.ManifestMu.Unlock()
	return 0
}

func (s *LuaPlugin) setID(L *lua.LState) int {
	s.Name = strings.ToLower(L.CheckString(1))
	return 0
}

func (s *LuaPlugin) uiAddStyle(L *lua.LState) int {
	name := s.Name
	if !s.HasCapability("theming") {
		L.RaiseError("[Security] [%s] Blocked %s (Capability '%s' not enabled)", name, "ui.add_style", "theming")
		return 0
	}
	css := L.CheckString(1)
	s.ManifestMu.Lock()
	s.CSS = css
	s.ManifestMu.Unlock()
	return 0
}

func (s *LuaPlugin) uiSetOverride(L *lua.LState) int {
	name := s.Name
	if !s.HasCapability("ui") {
		L.RaiseError("[Security] [%s] Blocked %s (Capability '%s' not enabled)", name, "ui.set_override", "ui")
		return 0
	}
	key := L.CheckString(1)
	tbl := L.CheckTable(2)

	override := make(map[string]string)
	tbl.ForEach(func(k, v lua.LValue) {
		override[k.String()] = v.String()
	})

	s.ManifestMu.Lock()
	s.UIOverrides[key] = override
	s.ManifestMu.Unlock()
	return 0
}

func (s *LuaPlugin) appRPC(L *lua.LState) int {
	name := s.Name
	if !s.HasCapability("interaction") {
		L.RaiseError("[Security] [%s] Blocked %s (Capability '%s' not enabled)", name, "app.rpc", "interaction")
		L.Push(lua.LNil)
		L.Push(lua.LString("Capability 'interaction' not enabled"))
		return 2
	}
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

	p.L.SetContext(L.Context())

	exports := p.L.GetGlobal("exports")
	if exports.Type() != lua.LTTable {
		L.Push(lua.LNil)
		L.Push(lua.LString("Plugin has no exports defined"))
		return 2
	}
	fn := p.L.GetField(exports, method)
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
	name := s.Name
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("%s [%s] [%s] %s\n", timestamp, strings.ToUpper(level), name, msg)
	return 0
}

func (s *LuaPlugin) addTab(L *lua.LState) int {
	name := s.Name
	if !s.HasCapability("ui") {
		L.RaiseError("[Security] [%s] Blocked %s (Capability '%s' not enabled)", name, "add_tab", "ui")
		return 0
	}
	id := L.CheckString(1)
	label := L.CheckString(2)
	icon := L.OptString(3, "Compass")
	sectionID := L.OptString(4, "")
	component := L.OptString(5, "")
	s.ManifestMu.Lock()
	exists := false
	for i, t := range s.Tabs {
		if t.ID == id {
			s.Tabs[i] = Tab{ID: id, Label: label, Icon: icon, SectionID: sectionID, Component: component}
			exists = true
			break
		}
	}
	if !exists {
		s.Tabs = append(s.Tabs, Tab{ID: id, Label: label, Icon: icon, SectionID: sectionID, Component: component})
	}
	s.ManifestMu.Unlock()
	return 0
}

func (s *LuaPlugin) addSection(L *lua.LState) int {
	name := s.Name
	if !s.HasCapability("ui") {
		L.RaiseError("[Security] [%s] Blocked %s (Capability '%s' not enabled)", name, "add_section", "ui")
		return 0
	}
	id := L.CheckString(1)
	label := L.CheckString(2)
	s.ManifestMu.Lock()
	exists := false
	for i, sect := range s.Sections {
		if sect.ID == id {
			s.Sections[i] = Section{ID: id, Label: label}
			exists = true
			break
		}
	}
	if !exists {
		s.Sections = append(s.Sections, Section{ID: id, Label: label})
	}
	s.ManifestMu.Unlock()
	return 0
}

func (s *LuaPlugin) addSettingsGroup(L *lua.LState) int {
	name := s.Name
	if !s.HasCapability("ui") {
		L.RaiseError("[Security] [%s] Blocked %s (Capability '%s' not enabled)", name, "add_settings_group", "ui")
		return 0
	}
	id := L.CheckString(1)
	label := L.CheckString(2)
	s.ManifestMu.Lock()
	s.SettingsGroups = append(s.SettingsGroups, SettingsGroup{ID: id, Label: label})
	s.ManifestMu.Unlock()
	return 0
}

func (s *LuaPlugin) appSpawn(L *lua.LState) int {
	name := s.Name
	if !s.HasCapability("background") {
		L.RaiseError("[Security] [%s] Blocked %s (Capability '%s' not enabled)", name, "app.spawn", "background")
		return 0
	}
	funcName := L.CheckString(1)
	argsStr := L.OptString(2, "{}")

	go func() {
		newL, err := NewLuaPlugin(s.Name, s.Path, s.Store)
		if err != nil {
			fmt.Printf("[Plugin] Spawn error: Failed to initialize isolated VM: %v\n", err)
			return
		}
		defer newL.L.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		newL.L.SetContext(ctx)

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
	name := s.Name
	if !s.HasCapability("background") {
		L.RaiseError("[Security] [%s] Blocked %s (Capability '%s' not enabled)", name, "app.sleep", "background")
		return 0
	}
	ms := L.CheckInt(1)
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return 0
}

func (s *LuaPlugin) storeSet(L *lua.LState) int {
	name := s.Name
	if !s.HasCapability("storage") {
		L.RaiseError("[Security] [%s] Blocked %s (Capability '%s' not enabled)", name, "store.set", "storage")
		return 0
	}
	key := L.CheckString(1)
	val := L.CheckString(2)
	fullKey := fmt.Sprintf("plugin_%s_%s", name, key)

	s.Store.SetCacheItem(fullKey, []byte(val))
	return 0
}

func (s *LuaPlugin) storeGet(L *lua.LState) int {
	name := s.Name
	if !s.HasCapability("storage") {
		L.RaiseError("[Security] [%s] Blocked %s (Capability '%s' not enabled)", name, "store.get", "storage")
		L.Push(lua.LNil)
		return 1
	}
	key := L.CheckString(1)
	fullKey := fmt.Sprintf("plugin_%s_%s", name, key)

	if val, ok := s.Store.GetCacheItem(fullKey); ok {
		L.Push(lua.LString(string(val)))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}
