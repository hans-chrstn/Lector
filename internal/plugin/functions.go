package plugin

import (
	"encoding/json"
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
	s.L.SetGlobal("url_join", s.L.NewFunction(s.urlJoin))
	s.L.SetGlobal("json_decode", s.L.NewFunction(s.jsonDecode))
	s.L.SetGlobal("json_encode", s.L.NewFunction(s.jsonEncode))

	stringMeta := s.L.GetGlobal("string").(*lua.LTable)
	s.L.SetField(stringMeta, "contains", s.L.NewFunction(func(L *lua.LState) int {
		str, substr := L.CheckString(1), L.CheckString(2)
		L.Push(lua.LBool(strings.Contains(str, substr)))
		return 1
	}))
	s.L.SetField(stringMeta, "trim", s.L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(strings.TrimSpace(L.CheckString(1))))
		return 1
	}))
	s.L.SetField(stringMeta, "split", s.L.NewFunction(func(L *lua.LState) int {
		str, sep := L.CheckString(1), L.CheckString(2)
		parts := strings.Split(str, sep)
		res := L.NewTable()
		for _, p := range parts {
			res.Append(lua.LString(strings.TrimSpace(p)))
		}
		L.Push(res)
		return 1
	}))
}

func (s *LuaPlugin) urlEncode(L *lua.LState) int {
	L.Push(lua.LString(url.QueryEscape(L.CheckString(1))))
	return 1
}

func (s *LuaPlugin) urlJoin(L *lua.LState) int {
	base, rel := L.CheckString(1), L.CheckString(2)
	u, err := url.Parse(base)
	if err != nil {
		L.Push(lua.LString(rel))
		return 1
	}
	r, err := url.Parse(rel)
	if err != nil {
		L.Push(lua.LString(rel))
		return 1
	}
	L.Push(lua.LString(u.ResolveReference(r).String()))
	return 1
}

func (s *LuaPlugin) jsonDecode(L *lua.LState) int {
	str := L.CheckString(1)
	var data interface{}
	if err := json.Unmarshal([]byte(str), &data); err != nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(luaValueToInterfaceTable(L, data))
	return 1
}

func (s *LuaPlugin) jsonEncode(L *lua.LState) int {
	val := L.CheckAny(1)
	data := luaValueToInterfaceGo(val)
	b, err := json.Marshal(data)
	if err != nil {
		L.Push(lua.LString(""))
		return 1
	}
	L.Push(lua.LString(string(b)))
	return 1
}

func luaValueToInterfaceTable(L *lua.LState, val interface{}) lua.LValue {
	switch v := val.(type) {
	case string:
		return lua.LString(v)
	case float64:
		return lua.LNumber(v)
	case bool:
		return lua.LBool(v)
	case []interface{}:
		tbl := L.NewTable()
		for _, item := range v {
			tbl.Append(luaValueToInterfaceTable(L, item))
		}
		return tbl
	case map[string]interface{}:
		tbl := L.NewTable()
		for k, item := range v {
			tbl.RawSetString(k, luaValueToInterfaceTable(L, item))
		}
		return tbl
	default:
		return lua.LNil
	}
}

func luaValueToInterfaceGo(v lua.LValue) interface{} {
	switch v.Type() {
	case lua.LTString:
		return v.String()
	case lua.LTNumber:
		return float64(v.(lua.LNumber))
	case lua.LTBool:
		return bool(v.(lua.LBool))
	case lua.LTTable:
		tbl := v.(*lua.LTable)
		if tbl.MaxN() > 0 {
			arr := []interface{}{}
			tbl.ForEach(func(k, v lua.LValue) {
				arr = append(arr, luaValueToInterfaceGo(v))
			})
			return arr
		}
		res := make(map[string]interface{})
		tbl.ForEach(func(k, v lua.LValue) {
			res[k.String()] = luaValueToInterfaceGo(v)
		})
		return res
	default:
		return nil
	}
}

func (s *LuaPlugin) luaPrint(L *lua.LState) int {
	name := s.Name
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("[%s] [%s] %s\n", timestamp, name, L.CheckString(1))
	return 0
}
