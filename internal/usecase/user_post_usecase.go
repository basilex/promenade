package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

var (
	ErrSlugAlreadyExists = errors.New("post with this slug already exists")
	ErrInvalidStatus     = errors.New("invalid post status")
)

type UserPostUseCase interface {
	CreatePost(ctx context.Context, userID uuidv7.UUID, title, content, excerpt string, tags, categories []string) (*entity.UserPost, error)
	GetPost(ctx context.Context, id uuidv7.UUID) (*entity.UserPost, error)
	GetPostBySlug(ctx context.Context, userID uuidv7.UUID, slug string) (*entity.UserPost, error)
	UpdatePost(ctx context.Context, userID, postID uuidv7.UUID, updates map[string]interface{}) (*entity.UserPost, error)
	DeletePost(ctx context.Context, userID, postID uuidv7.UUID) error
	SoftDeletePost(ctx context.Context, userID, postID uuidv7.UUID) error
	RestorePost(ctx context.Context, userID, postID uuidv7.UUID) error

	PublishPost(ctx context.Context, userID, postID uuidv7.UUID) error
	UnpublishPost(ctx context.Context, userID, postID uuidv7.UUID) error
	ArchivePost(ctx context.Context, userID, postID uuidv7.UUID) error
	SchedulePost(ctx context.Context, userID, postID uuidv7.UUID, scheduledAt time.Time) error

	ToggleFeatured(ctx context.Context, userID, postID uuidv7.UUID) error
	ToggleComments(ctx context.Context, userID, postID uuidv7.UUID) error

	ViewPost(ctx context.Context, postID uuidv7.UUID) error
	LikePost(ctx context.Context, postID uuidv7.UUID) error
	UnlikePost(ctx context.Context, postID uuidv7.UUID) error

	GetUserPosts(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.UserPost, error)
	GetPublishedPosts(ctx context.Context, limit, offset int) ([]*entity.UserPost, error)
	GetFeaturedPosts(ctx context.Context, limit int) ([]*entity.UserPost, error)
	SearchPosts(ctx context.Context, query string, limit, offset int) ([]*entity.UserPost, error)
	GetPostsByTag(ctx context.Context, tag string, limit, offset int) ([]*entity.UserPost, error)

	ListPosts(ctx context.Context, params repository.ListPostsParams) ([]*entity.UserPost, *pagination.Metadata, error)

	ProcessScheduledPosts(ctx context.Context) error
}

type userPostUseCase struct {
	postRepo repository.UserPostRepository
}

func NewUserPostUseCase(postRepo repository.UserPostRepository) UserPostUseCase {
	return &userPostUseCase{
		postRepo: postRepo,
	}
}

func (uc *userPostUseCase) CreatePost(ctx context.Context, userID uuidv7.UUID, title, content, excerpt string, tags, categories []string) (*entity.UserPost, error) {
	// Generate slug from title
	slug := generateSlug(title)

	// Check if slug already exists for this user
	existing, err := uc.postRepo.GetBySlug(ctx, userID, slug)
	if err != nil && !errors.Is(err, entity.ErrNotFound) {
		return nil, fmt.Errorf("failed to check slug uniqueness: %w", err)
	}
	if existing != nil {
		return nil, ErrSlugAlreadyExists
	}

	// Create new post
	post, err := entity.NewUserPost(userID, title, slug, content)
	if err != nil {
		return nil, fmt.Errorf("failed to create post entity: %w", err)
	}

	if excerpt != "" {
		post.Excerpt = &excerpt
	}
	post.Tags = tags
	post.Categories = categories

	// Calculate reading time
	post.ReadingTimeMinutes = calculateReadingTime(content)

	// Validate
	if err := post.Validate(); err != nil {
		return nil, fmt.Errorf("post validation failed: %w", err)
	}

	// Save to repository
	if err := uc.postRepo.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return post, nil
}

func (uc *userPostUseCase) GetPost(ctx context.Context, id uuidv7.UUID) (*entity.UserPost, error) {
	post, err := uc.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	return post, nil
}

func (uc *userPostUseCase) GetPostBySlug(ctx context.Context, userID uuidv7.UUID, slug string) (*entity.UserPost, error) {
	post, err := uc.postRepo.GetBySlug(ctx, userID, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get post by slug: %w", err)
	}

	return post, nil
}

func (uc *userPostUseCase) UpdatePost(ctx context.Context, userID, postID uuidv7.UUID, updates map[string]interface{}) (*entity.UserPost, error) {
	// Get existing post
	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	// Check ownership
	if post.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Apply updates
	if title, ok := updates["title"].(string); ok {
		post.Title = title
		post.Slug = generateSlug(title)
	}

	if content, ok := updates["content"].(string); ok {
		post.Content = content
		post.ReadingTimeMinutes = calculateReadingTime(content)
	}

	if excerpt, ok := updates["excerpt"].(string); ok {
		post.Excerpt = &excerpt
	}

	if tags, ok := updates["tags"].([]string); ok {
		post.Tags = tags
	}

	if categories, ok := updates["categories"].([]string); ok {
		post.Categories = categories
	}

	if featuredImage, ok := updates["featured_image"].(*entity.FeaturedImage); ok {
		post.FeaturedImage = featuredImage
	}

	if isPublic, ok := updates["is_public"].(bool); ok {
		post.IsPublic = isPublic
	}

	if metaTitle, ok := updates["meta_title"].(string); ok {
		post.MetaTitle = &metaTitle
	}

	if metaDescription, ok := updates["meta_description"].(string); ok {
		post.MetaDescription = &metaDescription
	}

	if metaKeywords, ok := updates["meta_keywords"].([]string); ok {
		post.MetaKeywords = metaKeywords
	}

	// Validate updated post
	if err := post.Validate(); err != nil {
		return nil, fmt.Errorf("post validation failed: %w", err)
	}

	// Save changes
	if err := uc.postRepo.Update(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return post, nil
}

func (uc *userPostUseCase) DeletePost(ctx context.Context, userID, postID uuidv7.UUID) error {
	// Get post to check ownership
	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}

	if post.UserID != userID {
		return ErrUnauthorized
	}

	if err := uc.postRepo.Delete(ctx, postID); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	return nil
}

func (uc *userPostUseCase) SoftDeletePost(ctx context.Context, userID, postID uuidv7.UUID) error {
	// Get post to check ownership
	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}

	if post.UserID != userID {
		return ErrUnauthorized
	}

	if err := uc.postRepo.SoftDelete(ctx, postID); err != nil {
		return fmt.Errorf("failed to soft delete post: %w", err)
	}

	return nil
}

func (uc *userPostUseCase) RestorePost(ctx context.Context, userID, postID uuidv7.UUID) error {
	// Note: Can't check ownership easily since post is soft deleted
	// In production, you might want to add GetByIDIncludingDeleted method
	if err := uc.postRepo.Restore(ctx, postID); err != nil {
		return fmt.Errorf("failed to restore post: %w", err)
	}

	return nil
}

func (uc *userPostUseCase) PublishPost(ctx context.Context, userID, postID uuidv7.UUID) error {
	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}

	if post.UserID != userID {
		return ErrUnauthorized
	}

	post.Publish()

	if err := uc.postRepo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	return nil
}

func (uc *userPostUseCase) UnpublishPost(ctx context.Context, userID, postID uuidv7.UUID) error {
	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}

	if post.UserID != userID {
		return ErrUnauthorized
	}

	post.Unpublish()

	if err := uc.postRepo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	return nil
}

func (uc *userPostUseCase) ArchivePost(ctx context.Context, userID, postID uuidv7.UUID) error {
	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}

	if post.UserID != userID {
		return ErrUnauthorized
	}

	post.Archive()

	if err := uc.postRepo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	return nil
}

func (uc *userPostUseCase) SchedulePost(ctx context.Context, userID, postID uuidv7.UUID, scheduledAt time.Time) error {
	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}

	if post.UserID != userID {
		return ErrUnauthorized
	}

	if err := post.Schedule(scheduledAt); err != nil {
		return fmt.Errorf("failed to schedule post: %w", err)
	}

	if err := uc.postRepo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	return nil
}

func (uc *userPostUseCase) ToggleFeatured(ctx context.Context, userID, postID uuidv7.UUID) error {
	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}

	if post.UserID != userID {
		return ErrUnauthorized
	}

	post.ToggleFeatured()

	if err := uc.postRepo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	return nil
}

func (uc *userPostUseCase) ToggleComments(ctx context.Context, userID, postID uuidv7.UUID) error {
	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}

	if post.UserID != userID {
		return ErrUnauthorized
	}

	post.ToggleComments()

	if err := uc.postRepo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	return nil
}

func (uc *userPostUseCase) ViewPost(ctx context.Context, postID uuidv7.UUID) error {
	if err := uc.postRepo.IncrementViews(ctx, postID); err != nil {
		return fmt.Errorf("failed to increment views: %w", err)
	}

	return nil
}

func (uc *userPostUseCase) LikePost(ctx context.Context, postID uuidv7.UUID) error {
	if err := uc.postRepo.IncrementLikes(ctx, postID); err != nil {
		return fmt.Errorf("failed to increment likes: %w", err)
	}

	return nil
}

func (uc *userPostUseCase) UnlikePost(ctx context.Context, postID uuidv7.UUID) error {
	if err := uc.postRepo.DecrementLikes(ctx, postID); err != nil {
		return fmt.Errorf("failed to decrement likes: %w", err)
	}

	return nil
}

func (uc *userPostUseCase) GetUserPosts(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.UserPost, error) {
	posts, err := uc.postRepo.GetUserPosts(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user posts: %w", err)
	}

	return posts, nil
}

func (uc *userPostUseCase) GetPublishedPosts(ctx context.Context, limit, offset int) ([]*entity.UserPost, error) {
	posts, err := uc.postRepo.GetPublishedPosts(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get published posts: %w", err)
	}

	return posts, nil
}

func (uc *userPostUseCase) GetFeaturedPosts(ctx context.Context, limit int) ([]*entity.UserPost, error) {
	posts, err := uc.postRepo.GetFeaturedPosts(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get featured posts: %w", err)
	}

	return posts, nil
}

func (uc *userPostUseCase) SearchPosts(ctx context.Context, query string, limit, offset int) ([]*entity.UserPost, error) {
	posts, err := uc.postRepo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search posts: %w", err)
	}

	return posts, nil
}

func (uc *userPostUseCase) GetPostsByTag(ctx context.Context, tag string, limit, offset int) ([]*entity.UserPost, error) {
	posts, err := uc.postRepo.GetByTag(ctx, tag, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts by tag: %w", err)
	}

	return posts, nil
}

func (uc *userPostUseCase) ListPosts(ctx context.Context, params repository.ListPostsParams) ([]*entity.UserPost, *pagination.Metadata, error) {
	posts, meta, err := uc.postRepo.List(ctx, params)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list posts: %w", err)
	}

	return posts, meta, nil
}

func (uc *userPostUseCase) ProcessScheduledPosts(ctx context.Context) error {
	scheduledPosts, err := uc.postRepo.GetScheduledPosts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get scheduled posts: %w", err)
	}

	for _, post := range scheduledPosts {
		if post.ShouldPublishNow() {
			if err := uc.postRepo.PublishScheduledPost(ctx, post.ID); err != nil {
				// Log error but continue processing other posts
				fmt.Printf("failed to publish scheduled post %s: %v\n", post.ID, err)
				continue
			}
		}
	}

	return nil
}

// Helper functions

func generateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters (keep only alphanumeric and hyphens)
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	slug = result.String()

	// Remove multiple consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	// Limit length to 100 characters
	if len(slug) > 100 {
		slug = slug[:100]
		slug = strings.TrimRight(slug, "-")
	}

	return slug
}

func calculateReadingTime(content string) int {
	// Average reading speed: 200-250 words per minute
	// Using 225 as average
	words := len(strings.Fields(content))
	minutes := words / 225

	if minutes == 0 {
		minutes = 1
	}

	return minutes
}
