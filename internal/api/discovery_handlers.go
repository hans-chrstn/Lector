package api

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/plugin"
)

func (h *API) Search(c *fiber.Ctx) error {
	pluginName, query := c.Query("plugin"), c.Query("q")
	fmt.Printf("[Search] Query: %s, Plugin: %s\n", query, pluginName)
	if pluginName == "all" {
		var wg sync.WaitGroup
		var mu sync.Mutex
		var all []map[string]interface{}

		activePlugins := h.GetActivePluginNames()
		fmt.Printf("[Search] Active plugins: %v\n", activePlugins)

		for name, s := range h.Plugins {
			if !s.HasCapability("catalog") {
				fmt.Printf("[Search] Skipping %s (no catalog capability)\n", name)
				continue
			}

			wg.Add(1)
			go func(name string, p *plugin.LuaPlugin) {
				defer wg.Done()
				res, err := p.Search(query)
				if err != nil {
					fmt.Printf("[Search] Error in plugin %s: %v\n", name, err)
					return
				}
				fmt.Printf("[Search] Plugin %s returned %d results\n", name, len(res))

				mu.Lock()
				for _, item := range res {
					all = append(all, map[string]interface{}{
						"title":     item.Title,
						"url":       item.URL,
						"cover_url": item.CoverURL,
						"info":      item.Info,
						"source":    name,
					})
				}
				mu.Unlock()
			}(name, s)
		}

		wg.Wait()
		fmt.Printf("[Search] Returning %d total results\n", len(all))
		return c.JSON(all)
	}

	s, ok := h.Plugins[pluginName]
	if !ok {
		fmt.Printf("[Search] Plugin %s not found\n", pluginName)
		return c.Status(404).SendString("Not found")
	}

	if !s.HasCapability("catalog") {
		fmt.Printf("[Search] Plugin %s missing catalog capability\n", pluginName)
		return c.Status(403).SendString("Plugin does not have catalog capability")
	}

	res, err := s.Search(query)
	if err != nil {
		fmt.Printf("[Search] Error in plugin %s: %v\n", pluginName, err)
		return c.Status(500).SendString(err.Error())
	}
	fmt.Printf("[Search] Plugin %s returned %d results\n", pluginName, len(res))
	return c.JSON(res)
}

func (h *API) GetActivePlugins(c *fiber.Ctx) error {
	return c.JSON(h.GetActivePluginNames())
}
