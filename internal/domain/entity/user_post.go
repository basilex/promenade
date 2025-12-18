package entity

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// PostStatus represents the publication status of a post
type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
	PostStatusArchived  PostStatus = "archived"
	PostStatusScheduled PostStatus = "scheduled"
)

// IsValid checks if post status is valid
func (s PostStatus) IsValid() bool {
	switch s {
	case PostStatusDraft, PostStatusPublished, PostStatusArchived, PostStatusScheduled:
		return true
	default:
		return false
	}
}

// Validate validates the post status
func (s PostStatus) Validate() error {
	if !s.IsValid() {
		return fmt.Errorf("invalid post status: %s", s)
	}
	return nil
}

// FeaturedImage represents post image metadata
type FeaturedImage struct {
	URL          string `json:"url"`
	Alt          string `json:"alt,omitempty"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// UserPost represents a blog post or article created by a user
type UserPost struct {
	ID     uuidv7.UUID `db:"id" validate:"required"`
	UserID uuidv7.UUID `db:"user_id" validate:"required"`

	// Content
	Title   string  `db:"title" validate:"required,min=3,max=255"`
	Slug    string  `db:"slug" validate:"required,max=255"`
	Excerpt *string `db:"excerpt" validate:"omitempty,max=500"` // nullable
	Content string  `db:"content" validate:"required,min=10"`

	// Media (stored as JSONB in database)
	FeaturedImage *FeaturedImage `db:"featured_image" validate:"omitempty"`

	// Status & Visibility
	Status            PostStatus `db:"status" validate:"required,oneof=draft published archived scheduled"`
	IsPublic          bool       `db:"is_public" validate:"-"`
	IsFeatured        bool       `db:"is_featured" validate:"-"`
	IsCommentsEnabled bool       `db:"is_comments_enabled" validate:"-"`

	// Publishing
	PublishedAt *time.Time `db:"published_at" validate:"omitempty"`
	ScheduledAt *time.Time `db:"scheduled_at" validate:"omitempty"`

	// SEO & Organization (JSONB/arrays in database)
	Tags            []string `db:"tags" validate:"omitempty,dive,min=2,max=50"`
	Categories      []string `db:"categories" validate:"omitempty,dive,min=2,max=50"`
	MetaTitle       *string  `db:"meta_title" validate:"omitempty,max=70"`
	MetaDescription *string  `db:"meta_description" validate:"omitempty,max=160"`
	MetaKeywords    []string `db:"meta_keywords" validate:"omitempty,dive,min=2,max=50"`

	// Engagement Metrics
	ViewCount          int `db:"view_count" validate:"min=0"`
	LikeCount          int `db:"like_count" validate:"min=0"`
	CommentCount       int `db:"comment_count" validate:"min=0"`
	ShareCount         int `db:"share_count" validate:"min=0"`
	ReadingTimeMinutes int `db:"reading_time_minutes" validate:"min=0"`

	// Soft delete
	DeletedAt *time.Time `db:"deleted_at" validate:"omitempty"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" validate:"required"`
	UpdatedAt time.Time `db:"updated_at" validate:"required"`
}

// NewUserPost creates a new post with required fields
func NewUserPost(userID uuidv7.UUID, title, slug, content string) (*UserPost, error) {
	post := &UserPost{
		ID:                 uuidv7.New(),
		UserID:             userID,
		Title:              strings.TrimSpace(title),
		Slug:               strings.TrimSpace(slug),
		Content:            content,
		Status:             PostStatusDraft,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		Tags:               []string{},
		Categories:         []string{},
		MetaKeywords:       []string{},
		ReadingTimeMinutes: calculateReadingTime(content),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := post.Validate(); err != nil {
		return nil, err
	}

	return post, nil
}

// Validate validates post fields
func (p *UserPost) Validate() error {
	if p.UserID == uuidv7.Nil {
		return fmt.Errorf("user_id is required")
	}

	if err := p.ValidateTitle(); err != nil {
		return err
	}

	if err := p.ValidateSlug(); err != nil {
		return err
	}

	if err := p.ValidateContent(); err != nil {
		return err
	}

	if err := p.ValidateExcerpt(); err != nil {
		return err
	}

	if !p.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", p.Status)
	}

	if err := p.ValidateSEO(); err != nil {
		return err
	}

	return nil
}

// ValidateTitle validates post title
func (p *UserPost) ValidateTitle() error {
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return fmt.Errorf("title is required")
	}

	length := utf8.RuneCountInString(title)
	if length < 3 {
		return fmt.Errorf("title must be at least 3 characters")
	}
	if length > 255 {
		return fmt.Errorf("title must not exceed 255 characters")
	}

	return nil
}

// ValidateSlug validates URL-friendly slug
func (p *UserPost) ValidateSlug() error {
	slug := strings.TrimSpace(p.Slug)
	if slug == "" {
		return fmt.Errorf("slug is required")
	}

	if len(slug) > 255 {
		return fmt.Errorf("slug must not exceed 255 characters")
	}

	// Slug should be lowercase with hyphens
	if !isValidSlug(slug) {
		return fmt.Errorf("slug must contain only lowercase letters, numbers, and hyphens")
	}

	return nil
}

// ValidateContent validates post content
func (p *UserPost) ValidateContent() error {
	content := strings.TrimSpace(p.Content)
	if content == "" {
		return fmt.Errorf("content is required")
	}

	length := utf8.RuneCountInString(content)
	if length < 10 {
		return fmt.Errorf("content must be at least 10 characters")
	}

	// Max content size: 1MB
	if len(p.Content) > 1024*1024 {
		return fmt.Errorf("content size exceeds 1MB limit")
	}

	return nil
}

// ValidateExcerpt validates excerpt if provided
func (p *UserPost) ValidateExcerpt() error {
	if p.Excerpt == nil {
		return nil
	}

	length := utf8.RuneCountInString(*p.Excerpt)
	if length > 500 {
		return fmt.Errorf("excerpt must not exceed 500 characters")
	}

	return nil
}

// ValidateSEO validates SEO fields
func (p *UserPost) ValidateSEO() error {
	if p.MetaTitle != nil {
		length := utf8.RuneCountInString(*p.MetaTitle)
		if length > 70 {
			return fmt.Errorf("meta_title must not exceed 70 characters")
		}
	}

	if p.MetaDescription != nil {
		length := utf8.RuneCountInString(*p.MetaDescription)
		if length > 160 {
			return fmt.Errorf("meta_description must not exceed 160 characters")
		}
	}

	return nil
}

// Publish marks the post as published
func (p *UserPost) Publish() {
	p.Status = PostStatusPublished
	now := time.Now()
	p.PublishedAt = &now
	p.ScheduledAt = nil
	p.UpdatedAt = now
}

// Unpublish marks the post as draft
func (p *UserPost) Unpublish() {
	p.Status = PostStatusDraft
	p.PublishedAt = nil
	p.UpdatedAt = time.Now()
}

// Archive marks the post as archived
func (p *UserPost) Archive() {
	p.Status = PostStatusArchived
	p.UpdatedAt = time.Now()
}

// Schedule sets scheduled publication time
func (p *UserPost) Schedule(scheduledAt time.Time) error {
	if scheduledAt.Before(time.Now()) {
		return fmt.Errorf("scheduled time must be in the future")
	}

	p.Status = PostStatusScheduled
	p.ScheduledAt = &scheduledAt
	p.PublishedAt = nil
	p.UpdatedAt = time.Now()

	return nil
}

// ToggleFeatured toggles featured status
func (p *UserPost) ToggleFeatured() {
	p.IsFeatured = !p.IsFeatured
	p.UpdatedAt = time.Now()
}

// ToggleComments toggles comments enabled/disabled
func (p *UserPost) ToggleComments() {
	p.IsCommentsEnabled = !p.IsCommentsEnabled
	p.UpdatedAt = time.Now()
}

// IncrementViews increments view counter
func (p *UserPost) IncrementViews() {
	p.ViewCount++
	// Don't update updated_at for view increments
}

// IncrementLikes increments like counter
func (p *UserPost) IncrementLikes() {
	p.LikeCount++
}

// DecrementLikes decrements like counter
func (p *UserPost) DecrementLikes() {
	if p.LikeCount > 0 {
		p.LikeCount--
	}
}

// IncrementComments increments comment counter
func (p *UserPost) IncrementComments() {
	p.CommentCount++
}

// DecrementComments decrements comment counter
func (p *UserPost) DecrementComments() {
	if p.CommentCount > 0 {
		p.CommentCount--
	}
}

// IncrementShares increments share counter
func (p *UserPost) IncrementShares() {
	p.ShareCount++
}

// SoftDelete marks post as deleted
func (p *UserPost) SoftDelete() {
	now := time.Now()
	p.DeletedAt = &now
	p.UpdatedAt = now
}

// Restore restores soft-deleted post
func (p *UserPost) Restore() {
	p.DeletedAt = nil
	p.UpdatedAt = time.Now()
}

// IsDeleted checks if post is soft-deleted
func (p *UserPost) IsDeleted() bool {
	return p.DeletedAt != nil
}

// IsPublished checks if post is published
func (p *UserPost) IsPublished() bool {
	return p.Status == PostStatusPublished && p.PublishedAt != nil
}

// IsScheduled checks if post is scheduled for future publication
func (p *UserPost) IsScheduled() bool {
	return p.Status == PostStatusScheduled && p.ScheduledAt != nil && p.ScheduledAt.After(time.Now())
}

// ShouldPublishNow checks if scheduled post should be published now
func (p *UserPost) ShouldPublishNow() bool {
	return p.Status == PostStatusScheduled && p.ScheduledAt != nil && !p.ScheduledAt.After(time.Now())
}

// IsOwnedBy checks if post belongs to user
func (p *UserPost) IsOwnedBy(userID uuidv7.UUID) bool {
	return p.UserID == userID
}

// GenerateSlug generates URL-friendly slug from title
func GenerateSlug(title string) string {
	slug := strings.ToLower(strings.TrimSpace(title))

	// Replace spaces and special characters with hyphens
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		if r == ' ' || r == '-' || r == '_' {
			return '-'
		}
		return -1 // remove character
	}, slug)

	// Replace multiple hyphens with single hyphen
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	return slug
}

// GenerateSlug generates slug for this post from its title
func (p *UserPost) GenerateSlug() string {
	return GenerateSlug(p.Title)
}

// isValidSlug checks if slug is properly formatted
func isValidSlug(slug string) bool {
	if slug == "" {
		return false
	}

	for _, r := range slug {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}

	// Should not start or end with hyphen
	if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
		return false
	}

	return true
}

// calculateReadingTime estimates reading time in minutes (avg 200 words/min)
func calculateReadingTime(content string) int {
	words := len(strings.Fields(content))
	minutes := words / 200
	if minutes == 0 {
		minutes = 1
	}
	return minutes
}

// MarshalFeaturedImage marshals FeaturedImage to JSONB
func MarshalFeaturedImage(img *FeaturedImage) ([]byte, error) {
	if img == nil {
		return nil, nil
	}
	return json.Marshal(img)
}

// UnmarshalFeaturedImage unmarshals JSONB to FeaturedImage
func UnmarshalFeaturedImage(data []byte) (*FeaturedImage, error) {
	if data == nil || len(data) == 0 {
		return nil, nil
	}

	var img FeaturedImage
	if err := json.Unmarshal(data, &img); err != nil {
		return nil, err
	}

	return &img, nil
}
