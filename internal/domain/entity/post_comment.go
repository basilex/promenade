package entity

import (
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// PostComment represents a user comment on a blog post
type PostComment struct {
	ID       uuidv7.UUID  `json:"id" db:"id" validate:"required"`
	PostID   uuidv7.UUID  `json:"post_id" db:"post_id" validate:"required"`
	UserID   uuidv7.UUID  `json:"user_id" db:"user_id" validate:"required"`
	ParentID *uuidv7.UUID `json:"parent_id,omitempty" db:"parent_id" validate:"omitempty"`

	Content string `json:"content" db:"content" validate:"required,min=1,max=5000"`

	// Edit tracking
	IsEdited bool       `json:"is_edited" db:"is_edited" validate:"-"`
	EditedAt *time.Time `json:"edited_at,omitempty" db:"edited_at" validate:"omitempty"`

	// Engagement metrics
	LikeCount  int `json:"like_count" db:"like_count" validate:"min=0"`
	ReplyCount int `json:"reply_count" db:"reply_count" validate:"min=0"`

	// Soft delete
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at" validate:"omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" validate:"required"`
}

// NewPostComment creates a new comment
func NewPostComment(postID, userID uuidv7.UUID, content string, parentID *uuidv7.UUID) (*PostComment, error) {
	if err := ValidateCommentContent(content); err != nil {
		return nil, err
	}

	return &PostComment{
		ID:         uuidv7.New(),
		PostID:     postID,
		UserID:     userID,
		ParentID:   parentID,
		Content:    content,
		IsEdited:   false,
		LikeCount:  0,
		ReplyCount: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

// ValidateCommentContent validates comment content
func ValidateCommentContent(content string) error {
	if len(content) < 1 {
		return fmt.Errorf("%w: content must be at least 1 character", ErrInvalidInput)
	}
	if len(content) > 5000 {
		return fmt.Errorf("%w: content must not exceed 5000 characters", ErrInvalidInput)
	}
	return nil
}

// UpdateContent updates comment content and marks as edited
func (c *PostComment) UpdateContent(content string) error {
	if err := ValidateCommentContent(content); err != nil {
		return err
	}

	c.Content = content
	c.IsEdited = true
	now := time.Now()
	c.EditedAt = &now
	c.UpdatedAt = now
	return nil
}

// IncrementLikes increments like count
func (c *PostComment) IncrementLikes() {
	c.LikeCount++
	c.UpdatedAt = time.Now()
}

// DecrementLikes decrements like count
func (c *PostComment) DecrementLikes() {
	if c.LikeCount > 0 {
		c.LikeCount--
		c.UpdatedAt = time.Now()
	}
}

// IncrementReplies increments reply count
func (c *PostComment) IncrementReplies() {
	c.ReplyCount++
	c.UpdatedAt = time.Now()
}

// DecrementReplies decrements reply count
func (c *PostComment) DecrementReplies() {
	if c.ReplyCount > 0 {
		c.ReplyCount--
		c.UpdatedAt = time.Now()
	}
}

// SoftDelete marks comment as deleted
func (c *PostComment) SoftDelete() {
	now := time.Now()
	c.DeletedAt = &now
	c.UpdatedAt = now
}

// Restore restores a soft-deleted comment
func (c *PostComment) Restore() {
	c.DeletedAt = nil
	c.UpdatedAt = time.Now()
}

// IsDeleted checks if comment is soft-deleted
func (c *PostComment) IsDeleted() bool {
	return c.DeletedAt != nil
}

// IsReply checks if comment is a reply to another comment
func (c *PostComment) IsReply() bool {
	return c.ParentID != nil
}

// IsTopLevel checks if comment is top-level (not a reply)
func (c *PostComment) IsTopLevel() bool {
	return c.ParentID == nil
}

// Validate validates the comment
func (c *PostComment) Validate() error {
	if c.PostID == (uuidv7.UUID{}) {
		return fmt.Errorf("%w: post_id is required", ErrInvalidInput)
	}
	if c.UserID == (uuidv7.UUID{}) {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if err := ValidateCommentContent(c.Content); err != nil {
		return err
	}
	if c.LikeCount < 0 || c.ReplyCount < 0 {
		return fmt.Errorf("%w: counts cannot be negative", ErrInvalidInput)
	}
	// Ensure edited timestamp logic is correct
	if c.IsEdited && c.EditedAt == nil {
		return fmt.Errorf("%w: edited_at must be set when is_edited is true", ErrInvalidInput)
	}
	if !c.IsEdited && c.EditedAt != nil {
		return fmt.Errorf("%w: edited_at should be null when is_edited is false", ErrInvalidInput)
	}
	return nil
}
