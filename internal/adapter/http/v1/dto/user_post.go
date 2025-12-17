package dto

import (
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CreatePostRequest represents request to create a new post
type CreatePostRequest struct {
	Title           string            `json:"title" binding:"required,min=3,max=200"`
	Content         string            `json:"content" binding:"required,min=10"`
	Excerpt         string            `json:"excerpt" binding:"omitempty,max=500"`
	Tags            []string          `json:"tags" binding:"omitempty"`
	Categories      []string          `json:"categories" binding:"omitempty"`
	IsPublic        bool              `json:"is_public"`
	FeaturedImage   *FeaturedImageDTO `json:"featured_image,omitempty"`
	MetaTitle       string            `json:"meta_title" binding:"omitempty,max=100"`
	MetaDescription string            `json:"meta_description" binding:"omitempty,max=250"`
	MetaKeywords    []string          `json:"meta_keywords" binding:"omitempty"`
}

// UpdatePostRequest represents request to update a post
type UpdatePostRequest struct {
	Title           *string           `json:"title" binding:"omitempty,min=3,max=200"`
	Content         *string           `json:"content" binding:"omitempty,min=10"`
	Excerpt         *string           `json:"excerpt" binding:"omitempty,max=500"`
	Tags            []string          `json:"tags"`
	Categories      []string          `json:"categories"`
	IsPublic        *bool             `json:"is_public"`
	FeaturedImage   *FeaturedImageDTO `json:"featured_image"`
	MetaTitle       *string           `json:"meta_title" binding:"omitempty,max=100"`
	MetaDescription *string           `json:"meta_description" binding:"omitempty,max=250"`
	MetaKeywords    []string          `json:"meta_keywords"`
}

// PublishPostRequest represents request to publish a post
type PublishPostRequest struct {
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

// FeaturedImageDTO represents featured image data
type FeaturedImageDTO struct {
	URL          string `json:"url" binding:"required"`
	Alt          string `json:"alt,omitempty"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// PostResponse represents a post in API response
type PostResponse struct {
	ID                string            `json:"id"`
	UserID            string            `json:"user_id"`
	Title             string            `json:"title"`
	Slug              string            `json:"slug"`
	Excerpt           *string           `json:"excerpt,omitempty"`
	Content           string            `json:"content"`
	FeaturedImage     *FeaturedImageDTO `json:"featured_image,omitempty"`
	Status            string            `json:"status"`
	IsPublic          bool              `json:"is_public"`
	IsFeatured        bool              `json:"is_featured"`
	IsCommentsEnabled bool              `json:"is_comments_enabled"`
	PublishedAt       *time.Time        `json:"published_at,omitempty"`
	ScheduledAt       *time.Time        `json:"scheduled_at,omitempty"`
	Tags              []string          `json:"tags"`
	Categories        []string          `json:"categories"`
	MetaTitle         *string           `json:"meta_title,omitempty"`
	MetaDescription   *string           `json:"meta_description,omitempty"`
	MetaKeywords      []string          `json:"meta_keywords,omitempty"`
	ViewCount         int               `json:"view_count"`
	LikeCount         int               `json:"like_count"`
	CommentCount      int               `json:"comment_count"`
	ShareCount        int               `json:"share_count"`
	ReadingTime       int               `json:"reading_time_minutes"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// PostListResponse represents paginated posts response
type PostListResponse struct {
	Posts      []PostResponse         `json:"posts"`
	Pagination map[string]interface{} `json:"pagination"`
}

// ToPostResponse converts entity to response DTO
func ToPostResponse(post *entity.UserPost) *PostResponse {
	resp := &PostResponse{
		ID:                post.ID.String(),
		UserID:            post.UserID.String(),
		Title:             post.Title,
		Slug:              post.Slug,
		Excerpt:           post.Excerpt,
		Content:           post.Content,
		Status:            string(post.Status),
		IsPublic:          post.IsPublic,
		IsFeatured:        post.IsFeatured,
		IsCommentsEnabled: post.IsCommentsEnabled,
		PublishedAt:       post.PublishedAt,
		ScheduledAt:       post.ScheduledAt,
		Tags:              post.Tags,
		Categories:        post.Categories,
		MetaTitle:         post.MetaTitle,
		MetaDescription:   post.MetaDescription,
		MetaKeywords:      post.MetaKeywords,
		ViewCount:         post.ViewCount,
		LikeCount:         post.LikeCount,
		CommentCount:      post.CommentCount,
		ShareCount:        post.ShareCount,
		ReadingTime:       post.ReadingTimeMinutes,
		CreatedAt:         post.CreatedAt,
		UpdatedAt:         post.UpdatedAt,
	}

	if post.FeaturedImage != nil {
		resp.FeaturedImage = &FeaturedImageDTO{
			URL:          post.FeaturedImage.URL,
			Alt:          post.FeaturedImage.Alt,
			Width:        post.FeaturedImage.Width,
			Height:       post.FeaturedImage.Height,
			ThumbnailURL: post.FeaturedImage.ThumbnailURL,
		}
	}

	return resp
}

// ToPostResponses converts slice of entities to response DTOs
func ToPostResponses(posts []*entity.UserPost) []PostResponse {
	responses := make([]PostResponse, len(posts))
	for i, post := range posts {
		responses[i] = *ToPostResponse(post)
	}
	return responses
}

// ToPostListResponse converts posts with pagination metadata to response
func ToPostListResponse(posts []*entity.UserPost, meta *pagination.Metadata) *PostListResponse {
	return &PostListResponse{
		Posts: ToPostResponses(posts),
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

// ToFeaturedImage converts DTO to entity
func (dto *FeaturedImageDTO) ToFeaturedImage() *entity.FeaturedImage {
	if dto == nil {
		return nil
	}
	return &entity.FeaturedImage{
		URL:          dto.URL,
		Alt:          dto.Alt,
		Width:        dto.Width,
		Height:       dto.Height,
		ThumbnailURL: dto.ThumbnailURL,
	}
}

// ToUpdateMap converts update request to map for usecase
func (req *UpdatePostRequest) ToUpdateMap() map[string]interface{} {
	updates := make(map[string]interface{})

	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.Excerpt != nil {
		updates["excerpt"] = *req.Excerpt
	}
	if req.Tags != nil {
		updates["tags"] = req.Tags
	}
	if req.Categories != nil {
		updates["categories"] = req.Categories
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}
	if req.FeaturedImage != nil {
		updates["featured_image"] = req.FeaturedImage.ToFeaturedImage()
	}
	if req.MetaTitle != nil {
		updates["meta_title"] = *req.MetaTitle
	}
	if req.MetaDescription != nil {
		updates["meta_description"] = *req.MetaDescription
	}
	if req.MetaKeywords != nil {
		updates["meta_keywords"] = req.MetaKeywords
	}

	return updates
}

// ListPostsRequest represents request to list posts with filters
type ListPostsRequest struct {
	UserID      *uuidv7.UUID       `form:"user_id"`
	Status      *entity.PostStatus `form:"status"`
	IsPublic    *bool              `form:"is_public"`
	IsFeatured  *bool              `form:"is_featured"`
	Tag         *string            `form:"tag"`
	SearchQuery *string            `form:"q"`
	SortBy      string             `form:"sort_by" binding:"omitempty,oneof=created_at updated_at published_at view_count like_count"`
	SortOrder   string             `form:"sort_order" binding:"omitempty,oneof=asc desc ASC DESC"`
	Limit       int                `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset      int                `form:"offset" binding:"omitempty,min=0"`
}
