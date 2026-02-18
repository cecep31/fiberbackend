package repository

import (
	"context"
	"fmt"

	"fiberbackend/internal/model"

	"gorm.io/gorm"
)

type PostLikeRepository interface {
	CreateLike(ctx context.Context, like *model.PostLike) error
	DeleteLike(ctx context.Context, postID, userID string) error
	GetLikesByPostID(ctx context.Context, postID string, limit, offset int) ([]*model.PostLike, int64, error)
	GetLikeStats(ctx context.Context, postID string) (*model.PostLikeStats, error)
	HasUserLikedPost(ctx context.Context, postID, userID string) (bool, error)
	GetLikeByUserAndPost(ctx context.Context, postID, userID string) (*model.PostLike, error)
}

type postLikeRepository struct {
	db *gorm.DB
}

func NewPostLikeRepository(db *gorm.DB) PostLikeRepository {
	return &postLikeRepository{db: db}
}

func (r *postLikeRepository) CreateLike(ctx context.Context, like *model.PostLike) error {
	return r.db.WithContext(ctx).Create(like).Error
}

func (r *postLikeRepository) DeleteLike(ctx context.Context, postID, userID string) error {
	return r.db.WithContext(ctx).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Delete(&model.PostLike{}).Error
}

func (r *postLikeRepository) GetLikesByPostID(ctx context.Context, postID string, limit, offset int) ([]*model.PostLike, int64, error) {
	var likes []*model.PostLike
	var total int64

	// Count total likes
	if err := r.db.WithContext(ctx).Model(&model.PostLike{}).Where("post_id = ?", postID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count likes: %w", err)
	}

	// Get paginated likes with user information
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("post_id = ?", postID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&likes).Error

	return likes, total, err
}

func (r *postLikeRepository) GetLikeStats(ctx context.Context, postID string) (*model.PostLikeStats, error) {
	var stats model.PostLikeStats
	stats.PostID = postID

	// Count total likes
	err := r.db.WithContext(ctx).
		Model(&model.PostLike{}).
		Where("post_id = ?", postID).
		Count(&stats.TotalLikes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get like stats: %w", err)
	}

	return &stats, nil
}

func (r *postLikeRepository) HasUserLikedPost(ctx context.Context, postID, userID string) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).Raw(
		`SELECT EXISTS (
			SELECT 1
			FROM post_likes
			WHERE post_id = ? AND user_id = ? AND deleted_at IS NULL
		)`,
		postID,
		userID,
	).Scan(&exists).Error
	return exists, err
}

func (r *postLikeRepository) GetLikeByUserAndPost(ctx context.Context, postID, userID string) (*model.PostLike, error) {
	var like model.PostLike
	err := r.db.WithContext(ctx).Preload("User").Where("post_id = ? AND user_id = ?", postID, userID).First(&like).Error
	return &like, err
}
