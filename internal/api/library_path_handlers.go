package api

import (
	"strconv"
	"sync/atomic"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/services"
)

var isScanning atomic.Bool

func (h *API) GetLibraryPaths(c *fiber.Ctx) error {
	var paths []models.LibraryPath
	db.DB.WithContext(c.UserContext()).Find(&paths)

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
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	if lp.Path == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Path is required"})
	}
	if lp.Pattern == "" {
		lp.Pattern = "None/Flat"
	}
	if err := db.DB.WithContext(c.UserContext()).Create(&lp).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(lp)
}

func (h *API) DeleteLibraryPath(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := db.DB.WithContext(c.UserContext()).Delete(&models.LibraryPath{}, uint(id)).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendString("Deleted")
}

func (h *API) ScanLibrary(c *fiber.Ctx) error {
	if isScanning.CompareAndSwap(false, true) {
		go func() {
			defer isScanning.Store(false)
			services.ScanLibraryPaths()
		}()
		return c.JSON(fiber.Map{"status": "Scan initiated"})
	}
	return c.Status(429).JSON(fiber.Map{"error": "Scan already in progress"})
}

func (h *API) ScanStatus(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"is_scanning": isScanning.Load(),
		"total":       services.ScanTotal.Load(),
		"done":        services.ScanDone.Load(),
	})
}
