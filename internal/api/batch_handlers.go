package api

import (
	"github.com/gofiber/fiber/v2"
)

func (h *API) BatchDeleteDocuments(c *fiber.Ctx) error {
	var req struct {
		IDs []uint `json:"ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	h.DocumentService.DeleteBatch(req.IDs)
	return c.SendString("Deleted")
}

func (h *API) BatchMoveDocuments(c *fiber.Ctx) error {
	var req struct {
		IDs     []uint `json:"ids"`
		GroupID uint   `json:"group_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	h.DocumentService.MoveBatch(req.IDs, req.GroupID)
	return c.SendString("Moved")
}

func (h *API) BatchArchiveDocuments(c *fiber.Ctx) error {
	var req struct {
		IDs     []uint `json:"ids"`
		Archive bool   `json:"archive"`
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	h.DocumentService.ArchiveBatch(req.IDs, req.Archive)
	return c.SendString("Updated")
}

func (h *API) BatchMarkReadDocuments(c *fiber.Ctx) error {
	var req struct {
		IDs    []uint `json:"ids"`
		IsRead bool   `json:"is_read"`
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	h.DocumentService.MarkReadBatch(req.IDs, req.IsRead)
	return c.SendString("Updated")
}
