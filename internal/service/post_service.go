package service

import (
	"context"
	"errors"
	"fmt"

	"fiberbackend/internal/model"
	"fiberbackend/internal/repository"
	"fiberbackend/pkg/constants"
)

type PostService interface {
	GetPosts(ctx context.Context, limit int, offset int) ([]*model.PostResponse, int64, error)
	GetPostsFiltered(ctx context.Context, filter *model.PostQueryFilter) ([]*model.PostResponse, int64, error)
	GetPostsByUsername(ctx context.Context, username string, offset int, limit int) ([]*model.PostResponse, int64, error)
	GetPostsRandom(ctx context.Context, limit int) ([]*model.PostResponse, error)
	GetPostByID(ctx context.Context, id string) (*model.PostResponse, error)
	GetPostBySlugAndUsername(ctx context.Context, slug string, username string) (*model.PostResponse, error)
	GetPostsByCreatedBy(ctx context.Context, createdBy string, offset int, limit int) ([]*model.PostResponse, int64, error)
	GetPostsByTag(ctx context.Context, tag string, limit int, offset int) ([]*model.PostResponse, int64, error)
	GetPostsSitemap(ctx context.Context, limit int) ([]model.PostSitemapEntry, int64, error)
	GetPostsTrending(ctx context.Context, limit int, offset int) ([]*model.PostResponse, int64, error)
	DeletePostByID(ctx context.Context, id string) error
	CreatePost(ctx context.Context, post *model.CreatePostDTO, userID string) (*model.Post, error)
	UpdatePost(ctx context.Context, id string, post *model.UpdatePostDTO) (*model.Post, error)
	IsAuthor(ctx context.Context, id string, userid string) error
}

// ErrNotAuthor is returned when the user is not the author of the resource.
var ErrNotAuthor = errors.New("not author")

type postService struct {
	postRepo   repository.PostRepository
	tagService TagService
}

func NewPostService(postRepo repository.PostRepository, tagService TagService) PostService {
	return &postService{postRepo: postRepo, tagService: tagService}
}

func (s *postService) IsAuthor(ctx context.Context, id string, userid string) error {
	post, err := s.postRepo.GetPostByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get post for author check: %w", err)
	}
	if post.CreatedBy == nil || *post.CreatedBy != userid {
		return ErrNotAuthor
	}
	return nil
}

func (s *postService) GetPostsByUsername(ctx context.Context, username string, offset int, limit int) ([]*model.PostResponse, int64, error) {
	// Input validation
	if limit < 0 {
		limit = 0
	}
	if offset < 0 {
		offset = 0
	}
	if username == "" {
		return []*model.PostResponse{}, 0, nil
	}

	posts, total, err := s.postRepo.GetPostByUsername(ctx, username, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	// Pre-allocate slice with known capacity to reduce memory allocations
	postsResponse := make([]*model.PostResponse, 0, len(posts))

	for _, post := range posts {
		postsResponse = append(postsResponse, post.ToResponse())
	}

	return postsResponse, total, nil
}

func (s *postService) CreatePost(ctx context.Context, post *model.CreatePostDTO, userID string) (*model.Post, error) {
	// Handle tags if they exist
	var tags []model.Tag
	if len(post.Tags) > 0 {
		for _, tagName := range post.Tags {
			if tagName == "" {
				continue // Skip empty tag names
			}

			// Try to find existing tag by name
			tag, err := s.findOrCreateTagByName(ctx, tagName)
			if err != nil {
				return nil, fmt.Errorf("failed to find or create tag '%s': %w", tagName, err)
			}
			tags = append(tags, *tag)
		}
	}

	// Create the post with tags
	createdPost, err := s.postRepo.CreatePostWithTags(ctx, post, userID, tags)
	if err != nil {
		return nil, fmt.Errorf("failed to create post with tags: %w", err)
	}
	return createdPost, nil
}

func (s *postService) GetPostBySlugAndUsername(ctx context.Context, slug string, username string) (*model.PostResponse, error) {
	post, err := s.postRepo.GetPostBySlugAndUsername(ctx, slug, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get post by slug and username: %w", err)
	}

	return post.ToResponse(), nil
}

func (s *postService) DeletePostByID(ctx context.Context, id string) error {
	return s.postRepo.DeletePostByID(ctx, id)
}

func (s *postService) UpdatePost(ctx context.Context, id string, post *model.UpdatePostDTO) (*model.Post, error) {
	return s.postRepo.UpdatePost(ctx, id, post)
}

func (s *postService) GetPostByID(ctx context.Context, id string) (*model.PostResponse, error) {
	post, err := s.postRepo.GetPostByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get post by id: %w", err)
	}

	return post.ToResponse(), nil
}

func (s *postService) GetPosts(ctx context.Context, limit int, offset int) ([]*model.PostResponse, int64, error) {
	// Input validation
	if limit < 0 {
		limit = 0
	}
	if offset < 0 {
		offset = 0
	}

	posts, total, err := s.postRepo.GetPosts(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Pre-allocate slice with known capacity to reduce memory allocations
	postsResponse := make([]*model.PostResponse, 0, len(posts))

	for _, post := range posts {
		postResponse := post.ToResponse()
		postsResponse = append(postsResponse, postResponse)
	}

	return postsResponse, total, nil
}

func (s *postService) GetPostsRandom(ctx context.Context, limit int) ([]*model.PostResponse, error) {
	// Input validation
	if limit < 0 {
		limit = 0
	}

	posts, err := s.postRepo.GetPostsRandom(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get random posts: %w", err)
	}

	// Pre-allocate slice with known capacity to reduce memory allocations
	postsResponse := make([]*model.PostResponse, 0, len(posts))

	for _, post := range posts {
		postResponse := post.ToResponse()
		postsResponse = append(postsResponse, postResponse)
	}

	return postsResponse, nil
}

func (s *postService) GetPostsByCreatedBy(ctx context.Context, createdBy string, offset int, limit int) ([]*model.PostResponse, int64, error) {
	// Input validation
	if limit < 0 {
		limit = 0
	}
	if offset < 0 {
		offset = 0
	}
	if createdBy == "" {
		return []*model.PostResponse{}, 0, nil
	}

	posts, total, err := s.postRepo.GetPostsByCreatedBy(ctx, createdBy, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	// Pre-allocate slice with known capacity to reduce memory allocations
	postsResponse := make([]*model.PostResponse, 0, len(posts))
	for _, post := range posts {
		postsResponse = append(postsResponse, post.ToResponse())
	}

	return postsResponse, total, nil
}

func (s *postService) GetPostsByTag(ctx context.Context, tag string, limit int, offset int) ([]*model.PostResponse, int64, error) {
	// Input validation: align with GetPostsFiltered defaults
	if limit <= 0 {
		limit = constants.DefaultLimit
	}
	if limit > constants.MaxLimit {
		limit = constants.MaxLimit
	}
	if offset < 0 {
		offset = 0
	}
	if tag == "" {
		return []*model.PostResponse{}, 0, nil
	}

	posts, total, err := s.postRepo.GetPostsByTag(ctx, tag, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Pre-allocate slice with known capacity to reduce memory allocations
	postsResponse := make([]*model.PostResponse, 0, len(posts))
	for _, post := range posts {
		postsResponse = append(postsResponse, post.ToResponse())
	}

	return postsResponse, total, nil
}

func (s *postService) GetPostsFiltered(ctx context.Context, filter *model.PostQueryFilter) ([]*model.PostResponse, int64, error) {
	// Input validation and defaults
	if filter.Limit < 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100 // Maximum limit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	posts, total, err := s.postRepo.GetPostsFiltered(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Pre-allocate slice with known capacity to reduce memory allocations
	postsResponse := make([]*model.PostResponse, 0, len(posts))

	for _, post := range posts {
		postResponse := post.ToResponse()
		postsResponse = append(postsResponse, postResponse)
	}

	return postsResponse, total, nil
}

func (s *postService) GetPostsSitemap(ctx context.Context, limit int) ([]model.PostSitemapEntry, int64, error) {
	if limit > 0 {
		if limit > constants.SitemapMaxLimit {
			limit = constants.SitemapMaxLimit
		}
	}

	entries, total, err := s.postRepo.GetPostsSitemap(ctx, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get posts sitemap: %w", err)
	}
	return entries, total, nil
}

func (s *postService) GetPostsTrending(ctx context.Context, limit int, offset int) ([]*model.PostResponse, int64, error) {
	if limit < 0 {
		limit = 0
	}
	if offset < 0 {
		offset = 0
	}

	posts, total, err := s.postRepo.GetPostsTrending(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	postsResponse := make([]*model.PostResponse, 0, len(posts))
	for _, post := range posts {
		postsResponse = append(postsResponse, post.ToResponse())
	}

	return postsResponse, total, nil
}

// findOrCreateTagByName finds an existing tag by name or creates a new one
func (s *postService) findOrCreateTagByName(ctx context.Context, tagName string) (*model.Tag, error) {
	return s.tagService.FindOrCreateByName(ctx, tagName)
}
