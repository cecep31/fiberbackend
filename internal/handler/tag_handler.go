package handler

import (
	"fiberbackend/internal/model"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"
	"fiberbackend/pkg/validator"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type TagHandler struct {
	service service.TagService
}

func NewTagHandler(service service.TagService) *TagHandler {
	return &TagHandler{service: service}
}

func (h *TagHandler) CreateTag(c fiber.Ctx) error {
	tag := new(model.Tag)
	if err := c.Bind().Body(tag); err != nil {
		return response.BadRequest(c, "Invalid request payload", err)
	}

	if err := h.service.CreateTag(c.Context(), tag); err != nil {
		return response.InternalServerError(c, "Failed to create tag", err)
	}

	return response.Created(c, "Tag created successfully", tag)
}

func (h *TagHandler) GetTags(c fiber.Ctx) error {
	limit, offset, err := validator.ValidatePaginationWithDefaults(c.Query("limit"), c.Query("offset"))
	if err != nil {
		return response.BadRequest(c, "Invalid pagination parameters", err)
	}

	tags, total, err := h.service.GetTags(c.Context(), offset, limit)
	if err != nil {
		return response.InternalServerError(c, "Failed to get tags", err)
	}

	meta := response.CalculatePaginationMeta(total, offset, limit)
	return response.SuccessWithMeta(c, "Successfully retrieved tags", tags, meta)
}

func (h *TagHandler) GetTagByID(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return response.BadRequest(c, "Invalid tag ID", err)
	}

	tag, err := h.service.GetTagByID(c.Context(), uint(id))
	if err != nil {
		return response.NotFound(c, "Tag not found", err)
	}

	return response.Success(c, "Successfully retrieved tag", tag)
}

func (h *TagHandler) UpdateTag(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return response.BadRequest(c, "Invalid tag ID", err)
	}

	tag := new(model.Tag)
	if err := c.Bind().Body(tag); err != nil {
		return response.BadRequest(c, "Invalid request payload", err)
	}
	tag.ID = int(id)

	if err := h.service.UpdateTag(c.Context(), tag); err != nil {
		return response.InternalServerError(c, "Failed to update tag", err)
	}

	return response.Success(c, "Tag updated successfully", tag)
}

func (h *TagHandler) DeleteTag(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return response.BadRequest(c, "Invalid tag ID", err)
	}

	if err := h.service.DeleteTag(c.Context(), uint(id)); err != nil {
		return response.InternalServerError(c, "Failed to delete tag", err)
	}

	return response.Success(c, "Tag deleted successfully", nil)
}
