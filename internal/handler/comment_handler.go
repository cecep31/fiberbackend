package handler

import (
	"fiberbackend/internal/dto"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"
	"fiberbackend/pkg/validator"

	"github.com/gofiber/fiber/v3"
)

type CommentHandler struct {
	commentService service.CommentService
	userService    service.UserService
}

func NewCommentHandler(commentService service.CommentService, userService service.UserService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		userService:    userService,
	}
}

func (h *CommentHandler) CreateComment(c fiber.Ctx) error {
	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID is required", nil)
	}

	var commentDTO dto.CreateCommentRequest
	if err := c.Bind().Body(&commentDTO); err != nil {
		return response.BadRequest(c, "Invalid request payload", err)
	}

	if err := validate.Validate(&commentDTO); err != nil {
		return response.FromValidateError(c, err)
	}

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	comment, err := h.commentService.CreateComment(c.Context(), postID, &commentDTO, userID)
	if err != nil {
		return respondError(c, "Failed to create comment", err)
	}

	return response.Created(c, "Comment created successfully", comment)
}

func (h *CommentHandler) GetCommentsByPostID(c fiber.Ctx) error {
	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID is required", nil)
	}

	comments, err := h.commentService.GetCommentsByPostID(c.Context(), postID)
	if err != nil {
		return respondError(c, "Comments fetched failed", err)
	}

	return response.Success(c, "Comments fetched successfully", comments)
}

func (h *CommentHandler) UpdateComment(c fiber.Ctx) error {
	postID := c.Params("id")
	if !validator.IsValidUUID(postID) {
		return response.BadRequest(c, "Invalid post ID", nil)
	}
	commentID := c.Params("comment_id")
	if !validator.IsValidUUID(commentID) {
		return response.BadRequest(c, "Invalid comment ID", nil)
	}

	var commentDTO dto.CreateCommentRequest
	if err := c.Bind().Body(&commentDTO); err != nil {
		return response.BadRequest(c, "Invalid request payload", err)
	}

	if err := validate.Validate(&commentDTO); err != nil {
		return response.FromValidateError(c, err)
	}

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	comment, err := h.commentService.UpdateComment(c.Context(), postID, commentID, commentDTO.Text, userID)
	if err != nil {
		return respondError(c, "Failed to update comment", err)
	}

	return response.Success(c, "Comment updated successfully", comment)
}

func (h *CommentHandler) DeleteComment(c fiber.Ctx) error {
	postID := c.Params("id")
	if !validator.IsValidUUID(postID) {
		return response.BadRequest(c, "Invalid post ID", nil)
	}
	commentID := c.Params("comment_id")
	if !validator.IsValidUUID(commentID) {
		return response.BadRequest(c, "Invalid comment ID", nil)
	}

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	// Super admins may delete any comment for moderation purposes.
	isAdmin := isSuperAdmin(c.Context(), h.userService, userID)

	if err := h.commentService.DeleteComment(c.Context(), postID, commentID, userID, isAdmin); err != nil {
		return respondError(c, "Failed to delete comment", err)
	}

	return response.Success(c, "Comment deleted successfully", nil)
}
