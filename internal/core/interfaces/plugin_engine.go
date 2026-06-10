package interfaces

import (
	"github.com/user/lector/internal/models"
)

type LuaPluginInterface interface {
	HasCapability(name string) bool
	Search(query string) ([]models.SearchItem, error)
}

type PluginEngine interface {
	GetPlugins() map[string]LuaPluginInterface
	GetPlugin(name string) (LuaPluginInterface, bool)
	RemovePlugin(name string)
	AddPlugin(name string, plugin LuaPluginInterface)
}
