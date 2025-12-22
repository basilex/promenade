package dto

import (
	"time"

	"github.com/basilex/promenade/internal/modules/posts/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CreateCommentRequest represents a request to create a comment
type CreateCommentRequest struct {
	Content  string         `json:"content" validate:"required,min=1,max=2000"`
	ParentID *uuidv7.UUID   `json:"parent_id,omitempty"`
}

// UpdateCommentRequest represents a request to update a comment
type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=2000"`
}

// CommentResponse represents a comment in API responses
type CommentResponse struct {
	ID            string     `json:"id"`
	PostID        string     `json:"post_id"`
	UserID        string     `json:"user_id"`
	ParentID      *string    `json:"parent_id,omitempty"`
	Content       string     `json:"content"`
	Depth         int        `json:"depth"`
	Path          string     `json:"path"`
	LikesCount    int        `json:"likes_count"`
	RepliesCount  int        `json:"replies_count"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ToCommentResponse converts a comment entity to a response DTO
func ToCommentResponse(comment *entity.Comment) CommentResponse {
	resp := CommentResponse{
		ID:            comment.ID.String(),
		PostID:        comment.PostID.String(),
		UserID:        comment.UserID.String(),
		Content:       comment.Content,
		Depth:         comment.Depth,
		Path:          comment.Path,
		LikesCount:    comment.LikesCount,
		RepliesCount:  comment.RepliesCount,
		CreatedAt:     comment.CreatedAt,
		UpdatedAt:     comment.UpdatedAt,
	}

	if comment.ParentID != nil {
		parentIDStr := comment.ParentID.String()
		resp.ParentID = &parentIDStr
	}

	return resp
}

// ToCommentResponses converts a slice of comment entities to response DTOs
func ToCommentResponses(comments []*entity.Comment) []CommentResponse {
	responses := make([]CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = ToCommentResponse(comment)
	}
	return responses
}
