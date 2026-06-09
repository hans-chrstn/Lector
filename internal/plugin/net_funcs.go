package plugin

import (
	"net/url"

	lua "github.com/yuin/gopher-lua"
)

func (s *LuaPlugin) registerNetFunctions() {
	s.L.SetGlobal("http_get", s.L.NewFunction(s.httpGet))
	s.L.SetGlobal("http_post", s.L.NewFunction(s.httpPost))

	net := s.L.NewTable()
	s.L.SetField(net, "fetch", s.L.NewFunction(s.netFetch))
	s.L.SetField(net, "url_encode", s.L.NewFunction(s.urlEncode))
	s.L.SetField(net, "url_decode", s.L.NewFunction(func(L *lua.LState) int {
		res, _ := url.QueryUnescape(L.CheckString(1))
		L.Push(lua.LString(res))
		return 1
	}))
	s.L.SetField(net, "download", s.L.NewFunction(s.netDownload))
	s.L.SetGlobal("net", net)
}

func (s *LuaPlugin) netFetch(L *lua.LState) int {
	u := L.CheckString(1)
	L.Push(lua.LString(s.Fetch("GET", u, "", "", false, nil)))
	return 1
}

func (s *LuaPlugin) httpGet(L *lua.LState) int {
	L.Push(lua.LString(s.Fetch("GET", L.CheckString(1), "", L.OptString(2, ""), L.OptBool(3, false), nil)))
	return 1
}

func (s *LuaPlugin) httpPost(L *lua.LState) int {
	L.Push(lua.LString(s.Fetch("POST", L.CheckString(1), L.CheckString(2), L.OptString(3, ""), L.OptBool(4, true), nil)))
	return 1
}

func (s *LuaPlugin) netDownload(L *lua.LState) int {
	if !s.HasCapability("network") || !s.HasCapability("storage") {
		L.Push(lua.LBool(false))
		return 1
	}

	u := L.CheckString(1)
	destPath := L.CheckString(2)
	options := L.OptTable(3, nil)

	var referer string
	headers := make(map[string]string)
	if options != nil {
		if val := options.RawGetString("referer"); val != lua.LNil {
			referer = val.String()
		}
		if val := options.RawGetString("headers"); val.Type() == lua.LTTable {
			val.(*lua.LTable).ForEach(func(k, v lua.LValue) {
				headers[k.String()] = v.String()
			})
		}
	}

	success := s.Download(u, destPath, referer, headers)
	L.Push(lua.LBool(success))
	return 1
}
