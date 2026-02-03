package handler

import (
	"net/http"

	"fiberbackend/internal/model"
	"fiberbackend/internal/service"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type PageHandler struct {
	pageService service.PageService
}

func NewPageHandler(pageService service.PageService) *PageHandler {
	return &PageHandler{pageService: pageService}
}

// CreatePage handles the creation of a new page
func (h *PageHandler) CreatePage(c fiber.Ctx) error {
	var page model.Page
	if err := c.Bind().Body(&page); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Invalid request format",
			"success": false,
		})
	}

	// Get user ID from context (set by auth middleware)
	userID, _ := c.Locals("user_id").(string)
	page.CreatedBy = userID

	if err := h.pageService.CreatePage(c.Context(), &page); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Failed to create page",
			"success": false,
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"data":    page,
		"message": "Page created successfully",
		"success": true,
	})
}

// GetPage retrieves a page by ID
func (h *PageHandler) GetPage(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Invalid page ID",
			"success": false,
		})
	}

	page, err := h.pageService.GetPageByID(c.Context(), id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Page not found",
			"success": false,
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"data":    page,
		"message": "Page retrieved successfully",
		"success": true,
	})
}

// GetWorkspacePages retrieves all pages in a workspace
func (h *PageHandler) GetWorkspacePages(c fiber.Ctx) error {
	workspaceID, err := uuid.Parse(c.Params("workspace_id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Invalid workspace ID",
			"success": false,
		})
	}

	pages, err := h.pageService.GetPagesByWorkspaceID(c.Context(), workspaceID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Failed to retrieve pages",
			"success": false,
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"data":    pages,
		"message": "Pages retrieved successfully",
		"success": true,
	})
}

// GetChildPages retrieves all child pages of a given page
func (h *PageHandler) GetChildPages(c fiber.Ctx) error {
	parentID, err := uuid.Parse(c.Params("parent_id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Invalid parent page ID",
			"success": false,
		})
	}

	pages, err := h.pageService.GetChildPages(c.Context(), parentID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Failed to retrieve child pages",
			"success": false,
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"data":    pages,
		"message": "Child pages retrieved successfully",
		"success": true,
	})
}

// UpdatePage updates an existing page
func (h *PageHandler) UpdatePage(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Invalid page ID",
			"success": false,
		})
	}

	var page model.Page
	if err := c.Bind().Body(&page); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Invalid request payload",
			"success": false,
		})
	}

	page.ID = id
	if err := h.pageService.UpdatePage(c.Context(), &page); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Failed to update page",
			"success": false,
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"data":    page,
		"message": "Page updated successfully",
		"success": true,
	})
}

// DeletePage deletes a page by ID
func (h *PageHandler) DeletePage(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Invalid page ID",
			"success": false,
		})
	}

	if err := h.pageService.DeletePage(c.Context(), id); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Failed to delete page",
			"success": false,
		})
	}

	return c.SendStatus(http.StatusNoContent)
}
