package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
)

func (h *API) GetGroups(c *fiber.Ctx) error {
	var g []models.Group
	db.DB.WithContext(c.UserContext()).Find(&g)
	return c.JSON(g)
}

func (h *API) CreateGroup(c *fiber.Ctx) error {
	g := models.Group{Name: c.FormValue("name")}
	db.DB.WithContext(c.UserContext()).Create(&g)
	return c.JSON(g)
}
