package plugin

import (
	"fmt"
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
	name := strings.TrimSuffix(filepath.Base(s.Path), ".lua")
	if !s.HasCapability("storage") {
		fmt.Printf("[Security] [%s] Blocked unauthorized file read (Capability 'storage' not enabled)\n", name)
		L.Push(lua.LNil)
		return 1
	}

	path := L.CheckString(1)
	sandboxDir, _ := filepath.Abs(filepath.Join("uploads", "plugins", name))
	sandboxDir = filepath.Clean(sandboxDir)
	fullPath, _ := filepath.Abs(filepath.Join(sandboxDir, path))
	fullPath = filepath.Clean(fullPath)

	if fullPath != sandboxDir && !strings.HasPrefix(fullPath, sandboxDir+string(filepath.Separator)) {
		fmt.Printf("[Security] [%s] Blocked directory traversal attempt: %s\n", name, path)
		L.Push(lua.LNil)
		return 1
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LString(string(content)))
	return 1
}

func (s *LuaPlugin) fsWriteFile(L *lua.LState) int {
	name := strings.TrimSuffix(filepath.Base(s.Path), ".lua")
	if !s.HasCapability("storage") {
		fmt.Printf("[Security] [%s] Blocked unauthorized file write (Capability 'storage' not enabled)\n", name)
		L.Push(lua.LBool(false))
		return 1
	}

	path := L.CheckString(1)
	content := L.CheckString(2)
	sandboxDir, _ := filepath.Abs(filepath.Join("uploads", "plugins", name))
	sandboxDir = filepath.Clean(sandboxDir)
	fullPath, _ := filepath.Abs(filepath.Join(sandboxDir, path))
	fullPath = filepath.Clean(fullPath)

	if fullPath != sandboxDir && !strings.HasPrefix(fullPath, sandboxDir+string(filepath.Separator)) {
		fmt.Printf("[Security] [%s] Blocked directory traversal attempt: %s\n", name, path)
		L.Push(lua.LBool(false))
		return 1
	}

	os.MkdirAll(sandboxDir, 0755)
	err := os.WriteFile(fullPath, []byte(content), 0644)
	L.Push(lua.LBool(err == nil))
	return 1
}
