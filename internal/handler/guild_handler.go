package handler

import (
	"errors"

	apperrors "fiberbackend/internal/apperror"
	"fiberbackend/internal/dto"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type GuildHandler struct {
	guildService service.GuildService
}

func NewGuildHandler(guildService service.GuildService) *GuildHandler {
	return &GuildHandler{guildService: guildService}
}

func (h *GuildHandler) CreateGuild(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User authentication required")
	}

	var req dto.CreateGuildRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", err)
	}
	if err := bindValidate(c, &req); err != nil {
		return err
	}

	guild, err := h.guildService.CreateGuild(c.Context(), userID, &req)
	if err != nil {
		if handled := handleGuildError(c, err); handled != nil {
			return handled
		}
		return response.InternalServerError(c, "Failed to create guild", err)
	}
	return response.Created(c, "Guild created successfully", guild)
}

// ListGuilds returns the public guild directory. It is open to anonymous
// callers, so no viewer-relative fields are filled in.
func (h *GuildHandler) ListGuilds(c fiber.Ctx) error {
	limit, offset := ParsePaginationParams(c, 20)
	search := c.Query("search")

	guilds, total, err := h.guildService.ListGuilds(c.Context(), search, limit, offset)
	if err != nil {
		if handled := handleGuildError(c, err); handled != nil {
			return handled
		}
		return response.InternalServerError(c, "Failed to fetch guilds", err)
	}

	meta := response.CalculatePaginationMeta(total, offset, limit)
	return response.SuccessWithMeta(c, "Guilds fetched successfully", guilds, meta)
}

// GetMyGuilds lists the guilds the caller belongs to, including private ones.
func (h *GuildHandler) GetMyGuilds(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User authentication required")
	}
	limit, offset := ParsePaginationParams(c, 20)

	guilds, total, err := h.guildService.ListMyGuilds(c.Context(), userID, limit, offset)
	if err != nil {
		if handled := handleGuildError(c, err); handled != nil {
			return handled
		}
		return response.InternalServerError(c, "Failed to fetch guilds", err)
	}

	meta := response.CalculatePaginationMeta(total, offset, limit)
	return response.SuccessWithMeta(c, "Guilds fetched successfully", guilds, meta)
}

func (h *GuildHandler) GetGuild(c fiber.Ctx) error {
	guildSlug := c.Params("slug")
	if guildSlug == "" {
		return response.BadRequest(c, "Guild slug is required", nil)
	}
	// OptionalAuth: an anonymous caller simply gets no membership fields.
	viewerID, _ := GetUserIDFromClaims(c)

	guild, err := h.guildService.GetGuildBySlug(c.Context(), guildSlug, viewerID)
	if err != nil {
		if handled := handleGuildError(c, err); handled != nil {
			return handled
		}
		return response.InternalServerError(c, "Failed to fetch guild", err)
	}
	return response.Success(c, "Guild fetched successfully", guild)
}

func (h *GuildHandler) UpdateGuild(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User authentication required")
	}
	guildSlug := c.Params("slug")
	if guildSlug == "" {
		return response.BadRequest(c, "Guild slug is required", nil)
	}

	var req dto.UpdateGuildRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", err)
	}
	if err := bindValidate(c, &req); err != nil {
		return err
	}

	guild, err := h.guildService.UpdateGuild(c.Context(), guildSlug, userID, &req)
	if err != nil {
		if handled := handleGuildError(c, err); handled != nil {
			return handled
		}
		return response.InternalServerError(c, "Failed to update guild", err)
	}
	return response.Success(c, "Guild updated successfully", guild)
}

func (h *GuildHandler) DeleteGuild(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User authentication required")
	}
	guildSlug := c.Params("slug")
	if guildSlug == "" {
		return response.BadRequest(c, "Guild slug is required", nil)
	}

	if err := h.guildService.DeleteGuild(c.Context(), guildSlug, userID); err != nil {
		if handled := handleGuildError(c, err); handled != nil {
			return handled
		}
		return response.InternalServerError(c, "Failed to delete guild", err)
	}
	return response.Success(c, "Guild deleted successfully", nil)
}

func (h *GuildHandler) JoinGuild(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User authentication required")
	}
	guildSlug := c.Params("slug")
	if guildSlug == "" {
		return response.BadRequest(c, "Guild slug is required", nil)
	}

	member, err := h.guildService.JoinGuild(c.Context(), guildSlug, userID)
	if err != nil {
		if handled := handleGuildError(c, err); handled != nil {
			return handled
		}
		return response.InternalServerError(c, "Failed to join guild", err)
	}
	return response.Created(c, "Joined guild successfully", member)
}

func (h *GuildHandler) LeaveGuild(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User authentication required")
	}
	guildSlug := c.Params("slug")
	if guildSlug == "" {
		return response.BadRequest(c, "Guild slug is required", nil)
	}

	if err := h.guildService.LeaveGuild(c.Context(), guildSlug, userID); err != nil {
		if handled := handleGuildError(c, err); handled != nil {
			return handled
		}
		return response.InternalServerError(c, "Failed to leave guild", err)
	}
	return response.Success(c, "Left guild successfully", nil)
}

func (h *GuildHandler) GetMembers(c fiber.Ctx) error {
	guildSlug := c.Params("slug")
	if guildSlug == "" {
		return response.BadRequest(c, "Guild slug is required", nil)
	}
	viewerID, _ := GetUserIDFromClaims(c)
	limit, offset := ParsePaginationParams(c, 50)

	members, total, err := h.guildService.ListMembers(c.Context(), guildSlug, viewerID, limit, offset)
	if err != nil {
		if handled := handleGuildError(c, err); handled != nil {
			return handled
		}
		return response.InternalServerError(c, "Failed to fetch guild members", err)
	}

	meta := response.CalculatePaginationMeta(total, offset, limit)
	return response.SuccessWithMeta(c, "Guild members fetched successfully", members, meta)
}

// handleGuildError maps guild domain errors to responses, returning nil when
// the error is not one of them so the caller can fall back to a 500.
func handleGuildError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, apperrors.ErrGuildNotFound):
		return response.NotFound(c, "Guild not found", err)
	case errors.Is(err, apperrors.ErrGuildNotOwned):
		return response.Forbidden(c, "Access forbidden")
	case errors.Is(err, apperrors.ErrGuildSlugExists):
		return response.Conflict(c, "Guild slug already taken", err.Error())
	case errors.Is(err, apperrors.ErrAlreadyGuildMember):
		return response.Conflict(c, "Already a member of this guild", err.Error())
	case errors.Is(err, apperrors.ErrNotGuildMember),
		errors.Is(err, apperrors.ErrGuildOwnerCannotLeave),
		errors.Is(err, apperrors.ErrGuildSlugInvalid),
		errors.Is(err, apperrors.ErrGuildSlugReserved):
		return response.BadRequest(c, err.Error(), err)
	case errors.Is(err, apperrors.ErrContentSpamRejected):
		return respondSpamRejected(c, "Content rejected by anti-spam policy", "name", err)
	default:
		return nil
	}
}
