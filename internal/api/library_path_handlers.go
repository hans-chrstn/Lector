package api

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/services"
)

func (h *API) GetLibraryPaths(c *fiber.Ctx) error {
	var paths []models.LibraryPath
	db.DB.Find(&paths)

	uploadsPath := "uploads"

	type LibraryPathResponse struct {
		ID       uint   `json:"id"`
		Path     string `json:"path"`
		Pattern  string `json:"pattern"`
		IsSystem bool   `json:"is_system"`
	}

	var res []LibraryPathResponse
	res = append(res, LibraryPathResponse{
		ID:       0,
		Path:     uploadsPath,
		Pattern:  "None/Flat",
		IsSystem: true,
	})

	for _, p := range paths {
		res = append(res, LibraryPathResponse{
			ID:       p.ID,
			Path:     p.Path,
			Pattern:  p.Pattern,
			IsSystem: false,
		})
	}

	return c.JSON(res)
}

func (h *API) AddLibraryPath(c *fiber.Ctx) error {
	var lp models.LibraryPath
	if err := c.BodyParser(&lp); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	if lp.Path == "" {
		return c.Status(400).SendString("Path is required")
	}
	if lp.Pattern == "" {
		lp.Pattern = "None/Flat"
	}
	if err := db.DB.Create(&lp).Error; err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.JSON(lp)
}

func (h *API) DeleteLibraryPath(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := db.DB.Delete(&models.LibraryPath{}, uint(id)).Error; err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.SendString("Deleted")
}

func (h *API) ScanLibrary(c *fiber.Ctx) error {
	go services.ScanLibraryPaths()
	return c.SendString("Scan initiated")
}
