package repository

import (
	"context"
	"fmt"

	"fiberbackend/internal/model"

	"gorm.io/gorm"
)

type CommentRepository interface {
	CreateComment(ctx context.Context, comment *model.PostComment) error
	GetCommentsByPostID(ctx context.Context, postID string) ([]*model.PostComment, error)
	GetCommentByID(ctx context.Context, id string) (*model.PostComment, error)
	UpdateComment(ctx context.Context, comment *model.PostComment) error
	DeleteComment(ctx context.Context, id string) error
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{
		db: db,
	}
}

func (r *commentRepository) CreateComment(ctx context.Context, comment *model.PostComment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *commentRepository) GetCommentsByPostID(ctx context.Context, postID string) ([]*model.PostComment, error) {
	var comments []*model.PostComment

	// Get all comments with user information
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("post_id = ?", postID).
		Order("created_at DESC").
		Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("failed to get comments by post id: %w", err)
	}

	return comments, nil
}

func (r *commentRepository) GetCommentByID(ctx context.Context, id string) (*model.PostComment, error) {
	var comment model.PostComment
	if err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Posts").
		Where("id = ?", id).
		First(&comment).Error; err != nil {
		return nil, fmt.Errorf("failed to get comment by id: %w", err)
	}
	return &comment, nil
}

func (r *commentRepository) UpdateComment(ctx context.Context, comment *model.PostComment) error {
	return r.db.WithContext(ctx).Save(comment).Error
}

func (r *commentRepository) DeleteComment(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.PostComment{}).Error
}
