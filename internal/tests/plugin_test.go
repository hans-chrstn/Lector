package tests

import (
	"os"
	"testing"

	"github.com/user/lector/internal/plugin"
)

func TestLuaPlugin(t *testing.T) {
	luaCode := `
		app.enable_capability("source")
		function search(q) return {{title="Test", url="mock://test", cover_url="", info="info"}} end
		function get_document(u) return {title="Test", url=u, chapters={{id=1, title="Ch1", url="mock://ch1"}}} end
		function get_chapter(u) return {title="Ch1", content="Hello"} end
	`
	err := os.WriteFile("mock_plugin.lua", []byte(luaCode), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("mock_plugin.lua")

	s, err := plugin.NewLuaPlugin("test", "mock_plugin.lua")
	if err != nil {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	t.Run("Search", func(t *testing.T) {
		res, err := s.Search("test")
		if err != nil || len(res) == 0 {
			t.Errorf("Search failed: %v", err)
		}
	})

	t.Run("GetDocument", func(t *testing.T) {
		doc, err := s.GetDocument("mock://test")
		if err != nil || doc.Title != "Test" {
			t.Errorf("GetDocument failed: %v", err)
		}
	})
}

func TestModularPluginSupport(t *testing.T) {
	t.Run("Lua Defined ID", func(t *testing.T) {
		luaCode := `
			app.enable_capability("ui")
			app.set_id("my_custom_id")
			app.add_action("selection", "Action", "func")
		`
		os.WriteFile("id_test.lua", []byte(luaCode), 0644)
		defer os.Remove("id_test.lua")

		p, err := plugin.NewLuaPlugin("initial_name", "id_test.lua")
		if err != nil {
			t.Fatal(err)
		}
		if p.Name != "my_custom_id" {
			t.Errorf("Expected ID override to 'my_custom_id', got %s", p.Name)
		}
	})

	t.Run("Selection Actions", func(t *testing.T) {
		luaCode := `
			app.enable_capability("ui")
			app.add_action("selection", "Define", "define")
		`
		os.WriteFile("action_test.lua", []byte(luaCode), 0644)
		defer os.Remove("action_test.lua")

		p, err := plugin.NewLuaPlugin("test", "action_test.lua")
		if err != nil {
			t.Fatal(err)
		}

		found := false
		for _, a := range p.Actions {
			if a.Context == "selection" && a.Label == "Define" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Selection action not registered")
		}
	})
}
