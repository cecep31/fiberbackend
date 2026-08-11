package handler

import (
	"errors"
	"strconv"
	"strings"

	apperrors "fiberbackend/internal/apperror"
	"fiberbackend/internal/dto"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/response"
	"fiberbackend/pkg/validator"

	"github.com/gofiber/fiber/v3"
)

type PostHandler struct {
	postService     service.PostService
	postViewService service.PostViewService
}

func (h *PostHandler) respondPostError(c fiber.Ctx, message string, err error) error {
	switch {
	case errors.Is(err, apperrors.ErrPostNotFound):
		return response.NotFound(c, message, err)
	case errors.Is(err, apperrors.ErrNotAuthor), errors.Is(err, apperrors.ErrPostNotOwned):
		return response.Forbidden(c, message)
	case errors.Is(err, apperrors.ErrFileNil), errors.Is(err, apperrors.ErrFileTooLarge), errors.Is(err, apperrors.ErrInvalidFileType), errors.Is(err, apperrors.ErrStorageUnavailable):
		return response.BadRequest(c, message, err)
	default:
		return response.InternalServerError(c, message, err)
	}
}

func NewPostHandler(postService service.PostService, postViewService service.PostViewService) *PostHandler {
	return &PostHandler{
		postService:     postService,
		postViewService: postViewService,
	}
}

func (h *PostHandler) GetPosts(c fiber.Ctx) error {
	filter := &dto.PostQueryFilter{
		Limit:     10,
		Offset:    0,
		Search:    c.Query("search"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		CreatedBy: c.Query("created_by"),
	}

	if limit := c.Query("limit"); limit != "" {
		if limitInt, err := strconv.Atoi(limit); err == nil && limitInt > 0 {
			filter.Limit = limitInt
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if offsetInt, err := strconv.Atoi(offset); err == nil && offsetInt > 0 {
			filter.Offset = offsetInt
		}
	}

	if published := c.Query("published"); published != "" {
		if pubBool, err := strconv.ParseBool(published); err == nil {
			filter.Published = &pubBool
		}
	}

	if tags := c.Query("tags"); tags != "" {
		filter.Tags = strings.Split(tags, ",")
		for i, tag := range filter.Tags {
			filter.Tags[i] = strings.TrimSpace(tag)
		}
	}

	posts, total, err := h.postService.GetPostsFiltered(c.Context(), filter)
	if err != nil {
		return response.InternalServerError(c, "Failed to get posts", err)
	}

	dto.TruncatePostBodies(posts, 250)

	return response.SuccessWithMeta(c, "Successfully retrieved posts", posts,
		response.CalculatePaginationMeta(total, filter.Offset, filter.Limit))
}

func (h *PostHandler) CreatePost(c fiber.Ctx) error {
	var postReq dto.CreatePostRequest
	if err := c.Bind().Body(&postReq); err != nil {
		return response.BadRequest(c, "Failed to create post", err)
	}

	if err := bindValidate(c, &postReq); err != nil {
		return err
	}

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	newpost, err := h.postService.CreatePost(c.Context(), &postReq, userID)
	if err != nil {
		return h.respondPostError(c, "Failed to create post", err)
	}
	return response.Created(c, "Successfully created post", map[string]any{
		"id": newpost.ID,
	})
}

func (h *PostHandler) UpdatePost(c fiber.Ctx) error {
	id := c.Params("id")
	if !validator.IsValidUUID(id) {
		return response.BadRequest(c, "Invalid post ID", nil)
	}

	var updateDTO dto.UpdatePostRequest
	if err := c.Bind().Body(&updateDTO); err != nil {
		return response.BadRequest(c, "Failed to update post", err)
	}

	if err := bindValidate(c, &updateDTO); err != nil {
		return err
	}

	updatedPost, err := h.postService.UpdatePost(c.Context(), id, &updateDTO)
	if err != nil {
		return h.respondPostError(c, "Failed to update post", err)
	}

	return response.Success(c, "Post updated successfully", updatedPost)
}

func (h *PostHandler) UpdateMyPost(c fiber.Ctx) error {
	id := c.Params("id")
	if !validator.IsValidUUID(id) {
		return response.BadRequest(c, "Invalid post ID", nil)
	}

	var updateDTO dto.UpdatePostRequest
	if err := c.Bind().Body(&updateDTO); err != nil {
		return response.BadRequest(c, "Failed to update post", err)
	}

	if err := bindValidate(c, &updateDTO); err != nil {
		return err
	}

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	err := h.postService.IsAuthor(c.Context(), id, userID)
	if err != nil {
		return h.respondPostError(c, "Failed to check post ownership", err)
	}

	updatedPost, err := h.postService.UpdatePost(c.Context(), id, &updateDTO)
	if err != nil {
		return h.respondPostError(c, "Failed to update post", err)
	}

	return response.Success(c, "Post updated successfully", updatedPost)
}

func (h *PostHandler) GetPostBySlugAndUsername(c fiber.Ctx) error {
	slug := c.Params("slug")
	username := c.Params("username")
	post, err := h.postService.GetPostBySlugAndUsername(c.Context(), slug, username)
	if err != nil {
		return h.respondPostError(c, "Failed to get post", err)
	}

	return response.Success(c, "Successfully retrieved post", post)
}

func (h *PostHandler) GetPost(c fiber.Ctx) error {
	id := c.Params("id")
	if !validator.IsValidUUID(id) {
		return response.BadRequest(c, "Invalid post ID", nil)
	}

	post, err := h.postService.GetPostByID(c.Context(), id)
	if err != nil {
		return h.respondPostError(c, "Failed to get post", err)
	}

	return response.Success(c, "Successfully retrieved post", post)
}

func (h *PostHandler) GetMyPost(c fiber.Ctx) error {
	id := c.Params("id")
	if !validator.IsValidUUID(id) {
		return response.BadRequest(c, "Invalid post ID", nil)
	}

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	if err := h.postService.IsAuthor(c.Context(), id, userID); err != nil {
		return h.respondPostError(c, "Failed to check post ownership", err)
	}

	post, err := h.postService.GetPostByID(c.Context(), id)
	if err != nil {
		return h.respondPostError(c, "Failed to get post", err)
	}

	return response.Success(c, "Successfully retrieved post", post)
}

func (h *PostHandler) DeletePost(c fiber.Ctx) error {
	id := c.Params("id")
	if !validator.IsValidUUID(id) {
		return response.BadRequest(c, "Invalid post ID", nil)
	}

	err := h.postService.DeletePostByID(c.Context(), id)
	if err != nil {
		return h.respondPostError(c, "Failed to delete post", err)
	}

	return response.Success(c, "Successfully deleted post", nil)
}

func (h *PostHandler) DeleteMyPost(c fiber.Ctx) error {
	id := c.Params("id")
	if !validator.IsValidUUID(id) {
		return response.BadRequest(c, "Invalid post ID", nil)
	}

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	if err := h.postService.IsAuthor(c.Context(), id, userID); err != nil {
		return h.respondPostError(c, "Failed to check post ownership", err)
	}

	err := h.postService.DeletePostByID(c.Context(), id)
	if err != nil {
		return h.respondPostError(c, "Failed to delete post", err)
	}

	return response.Success(c, "Successfully deleted post", nil)
}

func (h *PostHandler) GetPostsRandom(c fiber.Ctx) error {
	limit, _ := ParsePaginationParams(c, 9)
	if limit > 20 {
		limit = 20
	}
	posts, err := h.postService.GetPostsRandom(c.Context(), limit)
	if err != nil {
		return response.InternalServerError(c, "Failed to get posts", err)
	}

	dto.TruncatePostBodies(posts, 250)

	return response.Success(c, "Successfully retrieved posts", posts)
}

func (h *PostHandler) GetPostsTrending(c fiber.Ctx) error {
	limit, _ := ParsePaginationParams(c, 10)

	posts, err := h.postService.GetPostsTrending(c.Context(), limit)
	if err != nil {
		return response.InternalServerError(c, "Failed to get trending posts", err)
	}

	dto.TruncatePostBodies(posts, 250)

	return response.Success(c, "Successfully retrieved trending posts", posts)
}

func (h *PostHandler) GetMyPosts(c fiber.Ctx) error {
	limit, offset := ParsePaginationParams(c, 10)

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	posts, total, err := h.postService.GetPostsByCreatedBy(c.Context(), userID, offset, limit)
	if err != nil {
		return response.InternalServerError(c, "Failed to get posts", err)
	}

	dto.TruncatePostBodies(posts, 250)

	return response.SuccessWithMeta(c, "Successfully retrieved posts", posts,
		response.CalculatePaginationMeta(total, offset, limit))
}

func (h *PostHandler) GetMyPostsAnalytics(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	q := &dto.MyPostsAnalyticsQuery{
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
	}

	analytics, err := h.postViewService.GetMyPostsAnalytics(c.Context(), userID, q)
	if err != nil {
		return response.InternalServerError(c, "Failed to get post analytics", err)
	}

	return response.Success(c, "Successfully retrieved post analytics", analytics)
}

func (h *PostHandler) GetMyPostsLikesByMonth(c fiber.Ctx) error {
	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	months := 12
	if m := c.Query("months"); m != "" {
		if v, err := strconv.Atoi(m); err == nil && v >= 1 && v <= 24 {
			months = v
		}
	}

	data, err := h.postViewService.GetMyPostsLikesByMonth(c.Context(), userID, &dto.MyPostsLikesByMonthQuery{
		Months: months,
	})
	if err != nil {
		return response.InternalServerError(c, "Failed to get likes by month", err)
	}

	return response.Success(c, "Successfully retrieved likes by month", data)
}

func (h *PostHandler) GetPostsForYou(c fiber.Ctx) error {
	limit, offset := ParsePaginationParams(c, 10)

	userID, ok := GetUserIDFromClaims(c)
	if !ok {
		return response.Unauthorized(c, "User not authenticated")
	}

	posts, total, err := h.postService.GetPostsForYou(c.Context(), userID, offset, limit)
	if err != nil {
		return response.InternalServerError(c, "Failed to get posts", err)
	}

	dto.TruncatePostBodies(posts, 250)

	return response.SuccessWithMeta(c, "Successfully retrieved for-you posts", posts,
		response.CalculatePaginationMeta(total, offset, limit))
}

func (h *PostHandler) GetPostsByUsername(c fiber.Ctx) error {
	username := c.Params("username")
	limit, offset := ParsePaginationParams(c, 10)

	posts, total, err := h.postService.GetPostsByUsername(c.Context(), username, offset, limit)
	if err != nil {
		return response.InternalServerError(c, "Failed to get posts", err)
	}

	dto.TruncatePostBodies(posts, 250)

	return response.SuccessWithMeta(c, "Successfully retrieved posts", posts,
		response.CalculatePaginationMeta(total, offset, limit))
}

func (h *PostHandler) GetPostsByTag(c fiber.Ctx) error {
	tag := c.Params("tag")
	limit, offset := ParsePaginationParams(c, 10)

	posts, total, err := h.postService.GetPostsByTag(c.Context(), tag, limit, offset)
	if err != nil {
		return response.InternalServerError(c, "Failed to get posts by tag", err)
	}

	dto.TruncatePostBodies(posts, 250)

	return response.SuccessWithMeta(c, "Successfully retrieved posts by tag", posts,
		response.CalculatePaginationMeta(total, offset, limit))
}

func (h *PostHandler) UploadImagePosts(c fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return response.BadRequest(c, "Failed to upload image", err)
	}

	if file == nil {
		return response.BadRequest(c, "No file uploaded", nil)
	}

	if err := h.postService.UploadImagePosts(c.Context(), file); err != nil {
		return h.respondPostError(c, "Failed to upload image", err)
	}
	return response.Success(c, "Successfully uploaded image", nil)
}

func (h *PostHandler) GetPostsForSitemap(c fiber.Ctx) error {
	posts, err := h.postService.GetPostsForSitemap(c.Context(), 1000)
	if err != nil {
		return response.InternalServerError(c, "Failed to get posts for sitemap", err)
	}

	return response.Success(c, "Successfully retrieved posts for sitemap", posts)
}
