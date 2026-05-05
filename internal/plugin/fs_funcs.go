package plugin

import (
	"os"
	"path/filepath"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func (s *LuaPlugin) registerFSFunctions() {
	fs := s.L.NewTable()
	s.L.SetField(fs, "read_file", s.L.NewFunction(s.fsReadFile))
	s.L.SetField(fs, "write_file", s.L.NewFunction(s.fsWriteFile))
	s.L.SetGlobal("fs", fs)
}

func (s *LuaPlugin) fsReadFile(L *lua.LState) int {
	path := L.CheckString(1)
	name := strings.TrimSuffix(filepath.Base(s.Path), ".lua")
	fullPath := filepath.Join("uploads", "plugins", name, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LString(string(content)))
	return 1
}

func (s *LuaPlugin) fsWriteFile(L *lua.LState) int {
	path := L.CheckString(1)
	content := L.CheckString(2)
	name := strings.TrimSuffix(filepath.Base(s.Path), ".lua")
	dir := filepath.Join("uploads", "plugins", name)
	os.MkdirAll(dir, 0755)
	fullPath := filepath.Join(dir, path)
	err := os.WriteFile(fullPath, []byte(content), 0644)
	L.Push(lua.LBool(err == nil))
	return 1
}
