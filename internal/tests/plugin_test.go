package tests

import (
	"os"
	"testing"

	"github.com/user/lector/internal/plugin"
)

func TestLuaPlugin(t *testing.T) {
	luaCode := `
		function search(q) return {{title="Test", url="mock://test", cover_url="", info="info"}} end
		function get_document(u) return {title="Test", url=u, chapters={{id=1, title="Ch1", url="mock://ch1"}}} end
		function get_chapter(u) return {title="Ch1", content="Hello"} end
	`
	err := os.WriteFile("mock_plugin.lua", []byte(luaCode), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("mock_plugin.lua")

	s, err := plugin.NewLuaPlugin("mock_plugin.lua")
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
