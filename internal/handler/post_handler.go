package handler

import (
	"errors"
	"strconv"
	"strings"

	"fiberbackend/internal/model"
	"fiberbackend/internal/service"
	"fiberbackend/pkg/constants"
	"fiberbackend/pkg/response"
	"fiberbackend/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

type PostHandler struct {
	postService     service.PostService
	postViewService service.PostViewService
}

func NewPostHandler(postService service.PostService, postViewService service.PostViewService) *PostHandler {
	return &PostHandler{
		postService:     postService,
		postViewService: postViewService,
	}
}

func (h *PostHandler) GetPosts(c fiber.Ctx) error {
	// Parse query parameters into filter struct
	filter := &model.PostQueryFilter{
		Limit:     10, // Default limit
		Offset:    0,  // Default offset
		Search:    c.Query("search"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		CreatedBy: c.Query("created_by"),
	}

	// Parse limit and offset
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

	// Parse published filter
	if published := c.Query("published"); published != "" {
		if pubBool, err := strconv.ParseBool(published); err == nil {
			filter.Published = &pubBool
		}
	}

	// Parse tags filter
	if tags := c.Query("tags"); tags != "" {
		filter.Tags = strings.Split(tags, ",")
		// Trim whitespace from each tag
		for i, tag := range filter.Tags {
			filter.Tags[i] = strings.TrimSpace(tag)
		}
	}

	posts, total, err := h.postService.GetPostsFiltered(c.Context(), filter)
	if err != nil {
		return response.InternalServerError(c, "Failed to get posts", err)
	}

	for _, post := range posts {
		if post.Body != nil && len(*post.Body) > 250 {
			truncated := (*post.Body)[:250] + " ..."
			post.Body = &truncated
		}
	}

	meta := response.PaginationMeta{
		TotalItems: int(total),
		Offset:     filter.Offset,
		Limit:      filter.Limit,
		TotalPages: int(total)/filter.Limit + 1,
	}
	if int(total)%filter.Limit == 0 {
		meta.TotalPages = int(total) / filter.Limit
	}

	return response.SuccessWithMeta(c, "Successfully retrieved posts", posts, meta)
}

func (h *PostHandler) CreatePost(c fiber.Ctx) error {
	var postReq model.CreatePostDTO
	if err := c.Bind().Body(&postReq); err != nil {
		return response.HandleBindError(c, err)
	}

	userID, err := utils.RequireUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Invalid or missing user context")
	}

	newpost, err := h.postService.CreatePost(c.Context(), &postReq, userID)
	if err != nil {
		return response.InternalServerError(c, "Failed to create post", err)
	}
	return response.Created(c, "Successfully created post", map[string]any{
		"id": newpost.ID,
	})
}

func (h *PostHandler) UpdatePost(c fiber.Ctx) error {
	id := c.Params("id")
	var updateDTO model.UpdatePostDTO
	if err := c.Bind().Body(&updateDTO); err != nil {
		return response.HandleBindError(c, err)
	}

	userID, err := utils.RequireUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Invalid or missing user context")
	}

	// Check if the user is the author of the post
	err = h.postService.IsAuthor(c.Context(), id, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotAuthor) {
			return response.Forbidden(c, "You are not the author of this post")
		}
		return response.InternalServerError(c, "Failed to check post ownership", err)
	}

	updatedPost, err := h.postService.UpdatePost(c.Context(), id, &updateDTO)
	if err != nil {
		return response.InternalServerError(c, "Failed to update post", err)
	}

	return response.Success(c, "Post updated successfully", updatedPost)
}

func (h *PostHandler) GetPostBySlugAndUsername(c fiber.Ctx) error {
	slug := c.Params("slug")
	username := c.Params("username")
	post, err := h.postService.GetPostBySlugAndUsername(c.Context(), slug, username)
	if err != nil {
		return response.InternalServerError(c, "Failed to get post", err)
	}

	return response.Success(c, "Successfully retrieved post", post)
}

func (h *PostHandler) GetPost(c fiber.Ctx) error {
	id := c.Params("id")
	post, err := h.postService.GetPostByID(c.Context(), id)
	if err != nil {
		return response.InternalServerError(c, "Failed to get post", err)
	}

	return response.Success(c, "Successfully retrieved post", post)
}

func (h *PostHandler) DeletePost(c fiber.Ctx) error {
	id := c.Params("id")
	err := h.postService.DeletePostByID(c.Context(), id)
	if err != nil {
		return response.InternalServerError(c, "Failed to delete post", err)
	}

	return response.Success(c, "Successfully deleted post", nil)
}

func (h *PostHandler) GetPostsRandom(c fiber.Ctx) error {
	limitStr := c.Query("limit")
	if limitStr == "" {
		limitStr = "6"
	}
	limitInt, err := strconv.Atoi(limitStr)
	if err != nil || limitInt < 1 {
		limitInt = 6
	}
	if limitInt > 20 {
		limitInt = 20
	}
	posts, err := h.postService.GetPostsRandom(c.Context(), limitInt)
	if err != nil {
		return response.InternalServerError(c, "Failed to get posts", err)
	}

	for _, post := range posts {
		if post.Body != nil && len(*post.Body) > 250 {
			truncated := (*post.Body)[:250] + " ..."
			post.Body = &truncated
		}
	}

	return response.Success(c, "Successfully retrieved posts", posts)
}

func (h *PostHandler) GetPostsSitemap(c fiber.Ctx) error {
	offset := c.Query("offset")
	limit := c.Query("limit")
	offsetInt, err := strconv.Atoi(offset)
	if err != nil || offsetInt < 0 {
		offsetInt = constants.DefaultOffset
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		limitInt = constants.SitemapDefaultLimit
	}
	if limitInt > constants.SitemapMaxLimit {
		limitInt = constants.SitemapMaxLimit
	}

	posts, total, err := h.postService.GetPostsSitemap(c.Context(), limitInt, offsetInt)
	if err != nil {
		return response.InternalServerError(c, "Failed to get posts sitemap", err)
	}
	if posts == nil {
		posts = []model.PostSitemapEntry{}
	}

	totalInt := int(total)
	meta := response.PaginationMeta{
		TotalItems: totalInt,
		Offset:     offsetInt,
		Limit:      limitInt,
		TotalPages: totalInt/limitInt + 1,
	}
	if limitInt > 0 && totalInt%limitInt == 0 {
		meta.TotalPages = totalInt / limitInt
	}

	return response.SuccessWithMeta(c, "Successfully retrieved posts sitemap", struct {
		Posts []model.PostSitemapEntry `json:"posts"`
	}{Posts: posts}, meta)
}

func (h *PostHandler) GetMyPosts(c fiber.Ctx) error {
	offset := c.Query("offset")
	limit := c.Query("limit")
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		offsetInt = 0 // Default offset if not provided or invalid
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = 10 // Default limit if not provided or invalid
	}

	userID, err := utils.RequireUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Invalid or missing user context")
	}

	posts, total, err := h.postService.GetPostsByCreatedBy(c.Context(), userID, offsetInt, limitInt)
	if err != nil {
		return response.InternalServerError(c, "Failed to get posts", err)
	}

	for _, post := range posts {
		if post.Body != nil && len(*post.Body) > 250 {
			truncated := (*post.Body)[:250] + " ..."
			post.Body = &truncated
		}
	}

	meta := response.PaginationMeta{
		TotalItems: int(total),
		Offset:     offsetInt,
		Limit:      limitInt,
		TotalPages: int(total)/limitInt + 1,
	}
	if int(total)%limitInt == 0 {
		meta.TotalPages = int(total) / limitInt
	}

	return response.SuccessWithMeta(c, "success retrieving posts", posts, meta)
}

func (h *PostHandler) GetPostsByUsername(c fiber.Ctx) error {
	username := c.Params("username")
	offset := c.Query("offset")
	limit := c.Query("limit")
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		offsetInt = 0 // Default offset if not provided or invalid
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = 10 // Default limit if not provided or invalid
	}
	posts, total, err := h.postService.GetPostsByUsername(c.Context(), username, offsetInt, limitInt)
	if err != nil {
		return response.InternalServerError(c, "Failed to get posts", err)
	}

	for _, post := range posts {
		if post.Body != nil && len(*post.Body) > 250 {
			truncated := (*post.Body)[:250] + " ..."
			post.Body = &truncated
		}
	}

	meta := response.PaginationMeta{
		TotalItems: int(total),
		Offset:     offsetInt,
		Limit:      limitInt,
		TotalPages: int(total)/limitInt + 1,
	}
	if int(total)%limitInt == 0 {
		meta.TotalPages = int(total) / limitInt
	}

	return response.SuccessWithMeta(c, "success retrieving posts", posts, meta)
}

func (h *PostHandler) GetPostsByTag(c fiber.Ctx) error {
	tag := c.Params("tag")
	offset := c.Query("offset")
	limit := c.Query("limit")
	offsetInt, err := strconv.Atoi(offset)
	if err != nil || offsetInt < 0 {
		offsetInt = constants.DefaultOffset
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		limitInt = constants.DefaultLimit
	}
	if limitInt > constants.MaxLimit {
		limitInt = constants.MaxLimit
	}
	posts, total, err := h.postService.GetPostsByTag(c.Context(), tag, limitInt, offsetInt)
	if err != nil {
		return response.InternalServerError(c, "Failed to get posts by tag", err)
	}

	for _, post := range posts {
		if post.Body != nil && len(*post.Body) > constants.PostBodyTruncateLength {
			truncated := (*post.Body)[:constants.PostBodyTruncateLength] + " ..."
			post.Body = &truncated
		}
	}

	totalInt := int(total)
	meta := response.PaginationMeta{
		TotalItems: totalInt,
		Offset:     offsetInt,
		Limit:      limitInt,
		TotalPages: totalInt/limitInt + 1,
	}
	if limitInt > 0 && totalInt%limitInt == 0 {
		meta.TotalPages = totalInt / limitInt
	}

	return response.SuccessWithMeta(c, "success retrieving posts by tag", posts, meta)
}
