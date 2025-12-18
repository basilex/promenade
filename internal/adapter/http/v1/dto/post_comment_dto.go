package dto

import (
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/pagination"
)

// CreateCommentRequest represents request to create a comment
type CreateCommentRequest struct {
	PostID   string  `json:"post_id" binding:"required,uuid"`
	ParentID *string `json:"parent_id,omitempty" binding:"omitempty,uuid"`
	Content  string  `json:"content" binding:"required,min=1,max=5000"`
}

// UpdateCommentRequest represents request to update a comment
type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=5000"`
}

// CommentResponse represents a comment in API response
type CommentResponse struct {
	ID       string  `json:"id"`
	PostID   string  `json:"post_id"`
	UserID   string  `json:"user_id"`
	ParentID *string `json:"parent_id,omitempty"`
	Content  string  `json:"content"`

	IsEdited bool       `json:"is_edited"`
	EditedAt *time.Time `json:"edited_at,omitempty"`

	LikeCount  int `json:"like_count"`
	ReplyCount int `json:"reply_count"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CommentListResponse represents paginated comments response
type CommentListResponse struct {
	Comments   []CommentResponse `json:"comments"`
	Pagination map[string]any    `json:"pagination"`
}

// ToCommentResponse converts entity to response DTO
func ToCommentResponse(comment *entity.PostComment) *CommentResponse {
	resp := &CommentResponse{
		ID:         comment.ID.String(),
		PostID:     comment.PostID.String(),
		UserID:     comment.UserID.String(),
		Content:    comment.Content,
		IsEdited:   comment.IsEdited,
		EditedAt:   comment.EditedAt,
		LikeCount:  comment.LikeCount,
		ReplyCount: comment.ReplyCount,
		CreatedAt:  comment.CreatedAt,
		UpdatedAt:  comment.UpdatedAt,
	}

	if comment.ParentID != nil {
		parentIDStr := comment.ParentID.String()
		resp.ParentID = &parentIDStr
	}

	return resp
}

// ToCommentResponses converts slice of entities to response DTOs
func ToCommentResponses(comments []*entity.PostComment) []CommentResponse {
	responses := make([]CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = *ToCommentResponse(comment)
	}
	return responses
}

// ToCommentListResponse converts comments with pagination metadata to response
func ToCommentListResponse(comments []*entity.PostComment, meta *pagination.Metadata) *CommentListResponse {
	return &CommentListResponse{
		Comments: ToCommentResponses(comments),
		Pagination: map[string]interface{}{
			"total":        meta.Total,
			"limit":        meta.Limit,
			"offset":       meta.Offset,
			"total_pages":  meta.TotalPages,
			"current_page": meta.CurrentPage,
			"has_next":     meta.HasNext,
			"has_prev":     meta.HasPrev,
		},
	}
}
