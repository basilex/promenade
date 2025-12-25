package entity

import (
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Comment represents a user comment on a post
type Comment struct {
	ID        uuidv7.UUID  `json:"id" db:"id"`
	PostID    uuidv7.UUID  `json:"post_id" db:"post_id"`
	UserID    uuidv7.UUID  `json:"user_id" db:"user_id"`
	ParentID  *uuidv7.UUID `json:"parent_id,omitempty" db:"parent_id"` // For threaded comments
	Content   string       `json:"content" db:"content"`
	Depth     int          `json:"depth" db:"depth"`               // Nesting level (0 = top-level)
	Path      string       `json:"path" db:"path"`                 // Materialized path for tree traversal
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time   `json:"deleted_at,omitempty" db:"deleted_at"` // Soft delete

	// Counters
	LikesCount   int `json:"likes_count" db:"likes_count"`
	RepliesCount int `json:"replies_count" db:"replies_count"`
}

// NewComment creates a new comment
func NewComment(postID, userID uuidv7.UUID, content string, parentID *uuidv7.UUID) (*Comment, error) {
	now := time.Now()

	comment := &Comment{
		ID:           uuidv7.New(),
		PostID:       postID,
		UserID:       userID,
		ParentID:     parentID,
		Content:      content,
		Depth:        0,
		Path:         "",
		CreatedAt:    now,
		UpdatedAt:    now,
		LikesCount:   0,
		RepliesCount: 0,
	}

	if err := comment.Validate(); err != nil {
		return nil, err
	}

	return comment, nil
}

// Validate validates comment data
func (c *Comment) Validate() error {
	if c.Content == "" {
		return fmt.Errorf("content is required")
	}

	if len(c.Content) < 2 {
		return fmt.Errorf("content must be at least 2 characters")
	}

	if len(c.Content) > 2000 {
		return fmt.Errorf("content must be at most 2000 characters")
	}

	return nil
}

// IsTopLevel returns true if this is a top-level comment (not a reply)
func (c *Comment) IsTopLevel() bool {
	return c.ParentID == nil
}

// IsDeleted returns true if the comment is soft-deleted
func (c *Comment) IsDeleted() bool {
	return c.DeletedAt != nil
}

// SoftDelete marks the comment as deleted
func (c *Comment) SoftDelete() {
	now := time.Now()
	c.DeletedAt = &now
	c.UpdatedAt = now
}
