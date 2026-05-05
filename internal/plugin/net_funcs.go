package plugin

import (
	"fmt"
	"net/url"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func (s *LuaPlugin) registerNetFunctions() {
	s.L.SetGlobal("http_get", s.L.NewFunction(s.httpGet))
	s.L.SetGlobal("http_post", s.L.NewFunction(s.httpPost))
	s.L.SetGlobal("url_join", s.L.NewFunction(s.urlJoin))
	s.L.SetGlobal("url_encode", s.L.NewFunction(s.urlEncode))

	net := s.L.NewTable()
	s.L.SetField(net, "fetch", s.L.NewFunction(s.netFetch))
	s.L.SetField(net, "url_encode", s.L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(url.QueryEscape(L.CheckString(1))))
		return 1
	}))
	s.L.SetField(net, "url_decode", s.L.NewFunction(func(L *lua.LState) int {
		res, _ := url.QueryUnescape(L.CheckString(1))
		L.Push(lua.LString(res))
		return 1
	}))
	s.L.SetGlobal("net", net)
}

func (s *LuaPlugin) urlJoin(L *lua.LState) int {
	base, rel := L.CheckString(1), L.CheckString(2)
	bu, _ := url.Parse(base)
	ru, _ := url.Parse(rel)
	if bu == nil || ru == nil {
		L.Push(lua.LString(rel))
		return 1
	}
	L.Push(lua.LString(bu.ResolveReference(ru).String()))
	return 1
}

func (s *LuaPlugin) netFetch(L *lua.LState) int {
	u := L.CheckString(1)

	parsed, err := url.Parse(u)
	if err != nil {
		L.Push(lua.LString("Invalid URL"))
		return 1
	}

	allowed := false
	for _, domain := range s.Permissions {
		if domain == "*" || strings.HasSuffix(parsed.Host, domain) {
			allowed = true
			break
		}
	}

	if !allowed {
		fmt.Printf("[Security] Blocked unauthorized fetch to %s from plugin\n", parsed.Host)
		L.Push(lua.LString("UNAUTHORIZED: Domain not in plugin permissions"))
		return 1
	}

	L.Push(lua.LString(s.Fetch("GET", u, "", "", false)))
	return 1
}

func (s *LuaPlugin) httpGet(L *lua.LState) int {
	L.Push(lua.LString(s.Fetch("GET", L.CheckString(1), "", L.OptString(2, ""), L.OptBool(3, false))))
	return 1
}

func (s *LuaPlugin) httpPost(L *lua.LState) int {
	L.Push(lua.LString(s.Fetch("POST", L.CheckString(1), L.CheckString(2), L.OptString(3, ""), L.OptBool(4, true))))
	return 1
}
