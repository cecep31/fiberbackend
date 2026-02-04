package model

import (
	"time"

	"gorm.io/gorm"
)

// PostComment represents a comment on a post
type PostComment struct {
	ID               string         `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt        *time.Time     `json:"created_at"`
	UpdatedAt        *time.Time     `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
	Text             *string        `json:"text" gorm:"type:text"`
	PostID           string         `json:"post_id" gorm:"type:uuid"`
	ParrentCommentID *int64         `json:"parrent_comment_id" gorm:"type:bigint"`
	CreatedBy        *string        `json:"created_by" gorm:"type:uuid"`

	// Relationships
	User  *User `gorm:"foreignKey:CreatedBy" json:"user,omitempty"`
	Posts *Post `gorm:"foreignKey:PostID" json:"posts,omitempty"`
}

func (PostComment) TableName() string {
	return "post_comments"
}

type PostCommentResponse struct {
	ID        string       `json:"id"`
	PostID    string       `json:"post_id"`
	Text      *string      `json:"text"`
	User      *UserSummary `json:"user,omitempty"`
	CreatedAt *time.Time   `json:"created_at"`
	UpdatedAt *time.Time   `json:"updated_at"`
}

func (pc *PostComment) ToResponse() *PostCommentResponse {
	var userResp *UserSummary
	if pc.User != nil && pc.User.ID != "" {
		userResp = pc.User.ToSummary()
	}

	return &PostCommentResponse{
		ID:        pc.ID,
		PostID:    pc.PostID,
		Text:      pc.Text,
		User:      userResp,
		CreatedAt: pc.CreatedAt,
		UpdatedAt: pc.UpdatedAt,
	}
}

type CreatePostCommentDTO struct {
	Text string `json:"text" validate:"required,min=1,max=1000"`
}
