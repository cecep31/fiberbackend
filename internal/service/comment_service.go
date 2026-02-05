package service

import (
	"context"
	"errors"
	"fmt"

	"fiberbackend/internal/model"
	"fiberbackend/internal/repository"
)

type CommentService interface {
	CreateComment(ctx context.Context, postID string, dto *model.CreatePostCommentDTO, createdBy string) (*model.PostComment, error)
	GetCommentsByPostID(ctx context.Context, postID string) ([]*model.PostCommentResponse, error)
	GetCommentByID(ctx context.Context, id string) (*model.PostCommentResponse, error)
	UpdateComment(ctx context.Context, id string, content string, userID string) (*model.PostComment, error)
	DeleteComment(ctx context.Context, id string, userID string) error
	IsCommentAuthor(ctx context.Context, commentID string, userID string) error
}

type commentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository
}

func NewCommentService(commentRepo repository.CommentRepository, postRepo repository.PostRepository) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
	}
}

func (s *commentService) CreateComment(ctx context.Context, postID string, dto *model.CreatePostCommentDTO, createdBy string) (*model.PostComment, error) {
	// Verify post exists
	_, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, errors.New("post not found")
	}

	comment := &model.PostComment{
		PostID:    postID,
		Text:      &dto.Text,
		CreatedBy: &createdBy,
	}

	if err := s.commentRepo.CreateComment(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return comment, nil
}

func (s *commentService) GetCommentsByPostID(ctx context.Context, postID string) ([]*model.PostCommentResponse, error) {
	// Verify post exists
	_, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, errors.New("post not found")
	}

	comments, err := s.commentRepo.GetCommentsByPostID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by post id: %w", err)
	}

	responses := make([]*model.PostCommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = comment.ToResponse()
	}

	return responses, nil
}

func (s *commentService) GetCommentByID(ctx context.Context, id string) (*model.PostCommentResponse, error) {
	comment, err := s.commentRepo.GetCommentByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment by id: %w", err)
	}
	return comment.ToResponse(), nil
}

func (s *commentService) UpdateComment(ctx context.Context, id string, text string, userID string) (*model.PostComment, error) {
	comment, err := s.commentRepo.GetCommentByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment for update: %w", err)
	}

	if comment.CreatedBy != &userID {
		return nil, errors.New("not authorized to update this comment")
	}

	comment.Text = &text
	if err := s.commentRepo.UpdateComment(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return comment, nil
}

func (s *commentService) DeleteComment(ctx context.Context, id string, userID string) error {
	comment, err := s.commentRepo.GetCommentByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get comment for deletion: %w", err)
	}

	if comment.CreatedBy != &userID {
		return errors.New("not authorized to delete this comment")
	}

	if err := s.commentRepo.DeleteComment(ctx, id); err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}
	return nil
}

func (s *commentService) IsCommentAuthor(ctx context.Context, commentID string, userID string) error {
	comment, err := s.commentRepo.GetCommentByID(ctx, commentID)
	if err != nil {
		return fmt.Errorf("failed to get comment for author check: %w", err)
	}
	if comment.CreatedBy != &userID {
		return errors.New("not author")
	}
	return nil
}
