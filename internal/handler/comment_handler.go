package handler

import (
	"fiberbackend/internal/model"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"
	"fiberbackend/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

type CommentHandler struct {
	commentService service.CommentService
}

func NewCommentHandler(commentService service.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// CreateComment handles creating a new comment on a post
func (h *CommentHandler) CreateComment(c fiber.Ctx) error {
	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID is required", nil)
	}

	var dto model.CreatePostCommentDTO
	if err := c.Bind().Body(&dto); err != nil {
		return response.HandleBindError(c, err)
	}

	userID, err := utils.RequireUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Invalid or missing user context")
	}

	comment, err := h.commentService.CreateComment(c.Context(), postID, &dto, userID)
	if err != nil {
		return response.InternalServerError(c, "Failed to create comment", err)
	}

	return response.Created(c, "Comment created successfully", comment)
}

// GetCommentsByPostID handles getting all comments for a specific post
func (h *CommentHandler) GetCommentsByPostID(c fiber.Ctx) error {
	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID is required", nil)
	}

	comments, err := h.commentService.GetCommentsByPostID(c.Context(), postID)
	if err != nil {
		return response.InternalServerError(c, "Comments fetched failed", err)
	}

	return response.Success(c, "Comments fetched successfully", comments)
}

// UpdateComment handles updating a comment
func (h *CommentHandler) UpdateComment(c fiber.Ctx) error {
	commentID := c.Params("comment_id")
	if commentID == "" {
		return response.BadRequest(c, "Comment ID is required", nil)
	}

	var dto model.CreatePostCommentDTO
	if err := c.Bind().Body(&dto); err != nil {
		return response.HandleBindError(c, err)
	}

	userID, err := utils.RequireUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Invalid or missing user context")
	}

	comment, err := h.commentService.UpdateComment(c.Context(), commentID, dto.Text, userID)
	if err != nil {
		return response.InternalServerError(c, "Failed to update comment", err)
	}

	return response.Success(c, "Comment updated successfully", comment)
}

// DeleteComment handles deleting a comment
func (h *CommentHandler) DeleteComment(c fiber.Ctx) error {
	commentID := c.Params("comment_id")
	if commentID == "" {
		return response.BadRequest(c, "Comment ID is required", nil)
	}

	userID, err := utils.RequireUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Invalid or missing user context")
	}

	if err := h.commentService.DeleteComment(c.Context(), commentID, userID); err != nil {
		return response.InternalServerError(c, "Failed to delete comment", err)
	}

	return response.Success(c, "Comment deleted successfully", nil)
}
