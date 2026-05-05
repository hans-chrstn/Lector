package api

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/services"
)

func (h *API) Search(c *fiber.Ctx) error {
	plugin, query := c.Query("plugin"), c.Query("q")
	if plugin == "all" {
		var all []models.SearchItem
		for _, s := range h.Plugins {
			res, _ := s.Search(query)
			all = append(all, res...)
		}
		return c.JSON(all)
	}
	s, ok := h.Plugins[plugin]
	if !ok {
		return c.Status(404).SendString("Not found")
	}
	res, err := s.Search(query)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.JSON(res)
}

func (h *API) GetPopular(c *fiber.Ctx) error {
	plugin := c.Params("plugin")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	cacheKey := fmt.Sprintf("popular:%s:%d", plugin, page)
	var cached []models.SearchItem
	if ok, _ := services.GetCache(cacheKey, &cached); ok {
		return c.JSON(cached)
	}
	s, ok := h.Plugins[plugin]
	if !ok {
		return c.Status(404).SendString("Not found")
	}
	res, err := s.GetPopular(page)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	services.SetCache(cacheKey, res, 6*time.Hour)
	return c.JSON(res)
}

func (h *API) GetLatest(c *fiber.Ctx) error {
	plugin := c.Params("plugin")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	cacheKey := fmt.Sprintf("latest:%s:%d", plugin, page)
	var cached []models.SearchItem
	if ok, _ := services.GetCache(cacheKey, &cached); ok {
		return c.JSON(cached)
	}
	s, ok := h.Plugins[plugin]
	if !ok {
		return c.Status(404).SendString("Not found")
	}
	res, err := s.GetLatest(page)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	services.SetCache(cacheKey, res, 1*time.Hour)
	return c.JSON(res)
}
