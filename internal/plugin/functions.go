package plugin

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func (s *LuaPlugin) registerFunctions() {
	s.registerAppFunctions()
	s.registerNetFunctions()
	s.registerDocFunctions()
	s.registerFSFunctions()

	s.L.SetGlobal("print", s.L.NewFunction(s.luaPrint))
	s.L.SetGlobal("url_encode", s.L.NewFunction(s.urlEncode))

	stringMeta := s.L.GetGlobal("string").(*lua.LTable)
	s.L.SetField(stringMeta, "contains", s.L.NewFunction(func(L *lua.LState) int {
		str, substr := L.CheckString(1), L.CheckString(2)
		L.Push(lua.LBool(strings.Contains(str, substr)))
		return 1
	}))
}

func (s *LuaPlugin) urlEncode(L *lua.LState) int {
	L.Push(lua.LString(url.QueryEscape(L.CheckString(1))))
	return 1
}

func (s *LuaPlugin) luaPrint(L *lua.LState) int {
	name := s.Name
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("[%s] [%s] %s\n", timestamp, name, L.CheckString(1))
	return 0
}
