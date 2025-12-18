package entity

import (
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// PostComment represents a user comment on a blog post
type PostComment struct {
	ID       uuidv7.UUID  `json:"id" db:"id"`
	PostID   uuidv7.UUID  `json:"post_id" db:"post_id"`
	UserID   uuidv7.UUID  `json:"user_id" db:"user_id"`
	ParentID *uuidv7.UUID `json:"parent_id,omitempty" db:"parent_id"`

	Content string `json:"content" db:"content"`

	// Edit tracking
	IsEdited bool       `json:"is_edited" db:"is_edited"`
	EditedAt *time.Time `json:"edited_at,omitempty" db:"edited_at"`

	// Engagement metrics
	LikeCount  int `json:"like_count" db:"like_count"`
	ReplyCount int `json:"reply_count" db:"reply_count"`

	// Soft delete
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
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
		return ErrInvalidInput
	}
	if len(content) > 5000 {
		return ErrInvalidInput
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
	if c.ID == (uuidv7.UUID{}) {
		return ErrInvalidInput
	}
	if c.PostID == (uuidv7.UUID{}) {
		return ErrInvalidInput
	}
	if c.UserID == (uuidv7.UUID{}) {
		return ErrInvalidInput
	}
	if err := ValidateCommentContent(c.Content); err != nil {
		return err
	}
	if c.LikeCount < 0 {
		return ErrInvalidInput
	}
	if c.ReplyCount < 0 {
		return ErrInvalidInput
	}
	if c.IsEdited && c.EditedAt == nil {
		return ErrInvalidInput
	}
	if !c.IsEdited && c.EditedAt != nil {
		return ErrInvalidInput
	}
	return nil
}
