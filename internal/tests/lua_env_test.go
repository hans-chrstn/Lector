package tests

import (
	"os"
	"testing"

	"github.com/user/lector/internal/plugin"
)

func TestLuaEnvironment(t *testing.T) {
	luaCode := `
		assert(type(app) == "table", "app global should be a table")
		assert(type(net) == "table", "net global should be a table")
		assert(type(doc) == "table", "doc global should be a table")
		assert(type(fs) == "table", "fs global should be a table")
		assert(type(http_get) == "function", "http_get global should be a function")
	`
	err := os.WriteFile("test_env.lua", []byte(luaCode), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("test_env.lua")

	_, err = plugin.NewLuaPlugin("test_env.lua")
	if err != nil {
		t.Fatalf("Lua environment validation failed: %v", err)
	}
}
