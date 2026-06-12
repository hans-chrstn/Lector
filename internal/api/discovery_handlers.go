package api

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/plugin"
)

func (h *API) Search(c *fiber.Ctx) error {
	pluginName, query := c.Query("plugin"), c.Query("q")
	fmt.Printf("[Search] Query: %s, Plugin: %s\n", query, pluginName)

	allResults := make([]map[string]interface{}, 0)
	allErrors := make([]string, 0)

	if pluginName == "all" {
		var wg sync.WaitGroup
		var mu sync.Mutex

		activePlugins := h.GetActivePluginNames()
		fmt.Printf("[Search] Active plugins: %v\n", activePlugins)

		for name, s := range h.Engine.Plugins {
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
					mu.Lock()
					allErrors = append(allErrors, fmt.Sprintf("%s is unreachable: %v", name, err))
					mu.Unlock()
					return
				}
				fmt.Printf("[Search] Plugin %s returned %d results\n", name, len(res))

				mu.Lock()
				for _, item := range res {
					allResults = append(allResults, map[string]interface{}{
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
		fmt.Printf("[Search] Returning %d total results, %d errors\n", len(allResults), len(allErrors))
		return c.JSON(fiber.Map{
			"results": allResults,
			"errors":  allErrors,
		})
	}

	s, ok := h.Engine.Plugins[pluginName]
	if !ok {
		fmt.Printf("[Search] Plugin %s not found\n", pluginName)
		return c.Status(404).JSON(fiber.Map{"error": "Plugin not found"})
	}

	if !s.HasCapability("catalog") {
		fmt.Printf("[Search] Plugin %s missing catalog capability\n", pluginName)
		return c.Status(403).JSON(fiber.Map{"error": "Plugin does not have catalog capability"})
	}

	res, err := s.Search(query)
	if err != nil {
		fmt.Printf("[Search] Error in plugin %s: %v\n", pluginName, err)
		return c.JSON(fiber.Map{
			"results": []interface{}{},
			"errors":  []string{fmt.Sprintf("%s is unreachable: %v", pluginName, err)},
		})
	}
	fmt.Printf("[Search] Plugin %s returned %d results\n", pluginName, len(res))

	results := make([]map[string]interface{}, 0)
	for _, item := range res {
		results = append(results, map[string]interface{}{
			"title":     item.Title,
			"url":       item.URL,
			"cover_url": item.CoverURL,
			"info":      item.Info,
			"source":    pluginName,
		})
	}

	return c.JSON(fiber.Map{
		"results": results,
		"errors":  []string{},
	})
}

func (h *API) GetActivePlugins(c *fiber.Ctx) error {
	return c.JSON(h.GetActivePluginNames())
}

func (h *API) GetActivePluginNames() []string {
	var n []string
	for name, p := range h.Engine.Plugins {
		if p.HasCapability("catalog") {
			n = append(n, name)
		}
	}
	return n
}

func (h *API) PluginDirectory(c *fiber.Ctx) error {
	pluginName := strings.ToLower(c.Params("name"))
	dirId := c.Params("id")
	page := c.QueryInt("page", 1)

	s, ok := h.Engine.Plugins[pluginName]
	if !ok {
		return c.Status(404).JSON(fiber.Map{"error": "Plugin not found"})
	}

	res, err := s.GetDirectory(dirId, page)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("%v", err)})
	}

	results := make([]map[string]interface{}, 0)
	for _, item := range res {
		results = append(results, map[string]interface{}{
			"title":     item.Title,
			"url":       item.URL,
			"cover_url": item.CoverURL,
			"info":      item.Info,
			"source":    pluginName,
		})
	}

	return c.JSON(results)
}
