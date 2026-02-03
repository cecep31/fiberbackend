package handler

import (
	"fiberbackend/internal/model"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/validator"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type WorkspaceHandler struct {
	workspaceService service.WorkspaceService
}

func NewWorkspaceHandler(workspaceService service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{workspaceService: workspaceService}
}

// CreateWorkspace handles the creation of a new workspace
func (h *WorkspaceHandler) CreateWorkspace(c fiber.Ctx) error {
	type CreateWorkspaceRequest struct {
		Name        string `json:"name" validate:"required"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	}

	var req CreateWorkspaceRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Invalid request format",
			"success": false,
		})
	}

	if err := validator.ValidateStruct(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Validation failed",
			"success": false,
		})
	}

	// Get user ID from context (assuming it's set by auth middleware)
	userID, _ := c.Locals("user_id").(string)

	workspace := &model.Workspace{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.workspaceService.Create(c.Context(), workspace); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Failed to create workspace",
			"success": false,
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"data":    workspace,
		"message": "Workspace created successfully",
		"success": true,
	})
}

// GetWorkspaceByID retrieves a workspace by its ID
func (h *WorkspaceHandler) GetWorkspaceByID(c fiber.Ctx) error {
	workspaceID := c.Params("id")

	workspace, err := h.workspaceService.GetByID(c.Context(), workspaceID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]any{
			"error":   err.Error(),
			"message": "Failed to retrieve workspace",
			"success": false,
		})
	}

	return c.Status(http.StatusOK).JSON(map[string]any{
		"data":    workspace,
		"success": true,
	})
}

// GetAllWorkspaces retrieves all workspaces with pagination
func (h *WorkspaceHandler) GetAllWorkspaces(c fiber.Ctx) error {
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		offset = 0
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		limit = 10
	}

	workspaces, total, err := h.workspaceService.GetAll(c.Context(), offset, limit)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]any{
			"error":   err.Error(),
			"message": "Failed to retrieve workspaces",
			"success": false,
		})
	}

	return c.Status(http.StatusOK).JSON(map[string]any{
		"data":    workspaces,
		"success": true,
		"metadata": map[string]any{
			"totalItems": total,
		},
	})
}

// GetUserWorkspaces retrieves all workspaces a user is a member of
func (h *WorkspaceHandler) GetUserWorkspaces(c fiber.Ctx) error {
	// Get user ID from context (assuming it's set by auth middleware)
	userID, _ := c.Locals("user_id").(string)

	workspaces, err := h.workspaceService.GetByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]any{
			"error":   err.Error(),
			"message": "Failed to retrieve user workspaces",
			"success": false,
		})
	}

	return c.Status(http.StatusOK).JSON(map[string]any{
		"data":    workspaces,
		"success": true,
	})
}

// UpdateWorkspace updates an existing workspace
func (h *WorkspaceHandler) UpdateWorkspace(c fiber.Ctx) error {
	workspaceID := c.Params("id")

	type UpdateWorkspaceRequest struct {
		Name        string `json:"name" validate:"required"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	}

	var req UpdateWorkspaceRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]any{
			"error":   err.Error(),
			"message": "Invalid request format",
			"success": false,
		})
	}

	if err := validator.ValidateStruct(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]any{
			"error":   err.Error(),
			"message": "Validation failed",
			"success": false,
		})
	}

	// Get the existing workspace
	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]any{
			"error":   err.Error(),
			"message": "Invalid workspace ID",
			"success": false,
		})
	}

	workspace := &model.Workspace{
		ID:          wsID,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		UpdatedAt:   time.Now(),
	}

	if err := h.workspaceService.Update(c.Context(), workspace); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]any{
			"error":   err.Error(),
			"message": "Failed to update workspace",
			"success": false,
		})
	}

	return c.Status(http.StatusOK).JSON(map[string]any{
		"message": "Workspace updated successfully",
		"success": true,
	})
}

// DeleteWorkspace soft deletes a workspace
func (h *WorkspaceHandler) DeleteWorkspace(c fiber.Ctx) error {
	workspaceID := c.Params("id")

	if err := h.workspaceService.Delete(c.Context(), workspaceID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]any{
			"error":   err.Error(),
			"message": "Failed to delete workspace",
			"success": false,
		})
	}

	return c.Status(http.StatusOK).JSON(map[string]any{
		"message": "Workspace deleted successfully",
		"success": true,
	})
}

// AddMember adds a new member to a workspace
func (h *WorkspaceHandler) AddMember(c fiber.Ctx) error {
	workspaceID := c.Params("id")

	type AddMemberRequest struct {
		UserID string `json:"user_id" validate:"required"`
		Role   string `json:"role" validate:"required,oneof=admin editor viewer"`
	}

	var req AddMemberRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]any{
			"error":   err.Error(),
			"message": "Invalid request format",
			"success": false,
		})
	}

	if err := validator.ValidateStruct(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]any{
			"error":   err.Error(),
			"message": "Validation failed",
			"success": false,
		})
	}

	if err := h.workspaceService.AddMember(c.Context(), workspaceID, req.UserID, req.Role); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]any{
			"error":   err.Error(),
			"message": "Failed to add member",
			"success": false,
		})
	}

	return c.Status(http.StatusOK).JSON(map[string]any{
		"message": "Member added successfully",
		"success": true,
	})
}

// GetMembers retrieves all members of a workspace
func (h *WorkspaceHandler) GetMembers(c fiber.Ctx) error {
	workspaceID := c.Params("id")

	members, err := h.workspaceService.GetMembers(c.Context(), workspaceID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]any{
			"error":   err.Error(),
			"message": "Failed to retrieve members",
			"success": false,
		})
	}

	return c.Status(http.StatusOK).JSON(map[string]any{
		"data":    members,
		"success": true,
	})
}

// UpdateMemberRole updates a member's role in a workspace
func (h *WorkspaceHandler) UpdateMemberRole(c fiber.Ctx) error {
	workspaceID := c.Params("id")
	userID := c.Params("user_id")

	type UpdateMemberRoleRequest struct {
		Role string `json:"role" validate:"required,oneof=admin editor viewer"`
	}

	var req UpdateMemberRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]any{
			"error":   err.Error(),
			"message": "Invalid request format",
			"success": false,
		})
	}

	if err := validator.ValidateStruct(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]any{
			"error":   err.Error(),
			"message": "Validation failed",
			"success": false,
		})
	}

	if err := h.workspaceService.UpdateMemberRole(c.Context(), workspaceID, userID, req.Role); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]any{
			"error":   err.Error(),
			"message": "Failed to update member role",
			"success": false,
		})
	}

	return c.Status(http.StatusOK).JSON(map[string]any{
		"message": "Member role updated successfully",
		"success": true,
	})
}

// RemoveMember removes a member from a workspace
func (h *WorkspaceHandler) RemoveMember(c fiber.Ctx) error {
	workspaceID := c.Params("id")
	userID := c.Params("user_id")

	if err := h.workspaceService.RemoveMember(c.Context(), workspaceID, userID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]any{
			"error":   err.Error(),
			"message": "Failed to remove member",
			"success": false,
		})
	}

	return c.Status(http.StatusOK).JSON(map[string]any{
		"message": "Member removed successfully",
		"success": true,
	})
}
