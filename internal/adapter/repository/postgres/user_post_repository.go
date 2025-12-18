package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type userPostRepository struct {
	*BaseRepository
}

// NewUserPostRepository creates a new user post repository
func NewUserPostRepository(db *sqlx.DB) repository.UserPostRepository {
	return &userPostRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *userPostRepository) Create(ctx context.Context, post *entity.UserPost) error {
	featuredImage, err := entity.MarshalFeaturedImage(post.FeaturedImage)
	if err != nil {
		return fmt.Errorf("failed to marshal featured image: %w", err)
	}

	// Ensure nil slices are marshaled as empty arrays
	if post.Tags == nil {
		post.Tags = []string{}
	}
	if post.Categories == nil {
		post.Categories = []string{}
	}
	if post.MetaKeywords == nil {
		post.MetaKeywords = []string{}
	}

	tagsJSON, _ := json.Marshal(post.Tags)
	categoriesJSON, _ := json.Marshal(post.Categories)
	metaKeywords := pq.Array(post.MetaKeywords)

	query := `
		INSERT INTO user_posts (
			id, user_id, title, slug, excerpt, content,
			featured_image, status, is_public, is_featured, is_comments_enabled,
			published_at, scheduled_at, tags, categories,
			meta_title, meta_description, meta_keywords,
			view_count, like_count, comment_count, share_count, reading_time_minutes,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $17, $18,
			$19, $20, $21, $22, $23,
			$24, $25
		)
	`

	// Convert featuredImage to interface{} to ensure NULL is sent for nil
	var featuredImageParam interface{}
	if len(featuredImage) > 0 {
		featuredImageParam = featuredImage
	} else {
		featuredImageParam = nil
	}

	err = r.Exec(ctx, query,
		post.ID, post.UserID, post.Title, post.Slug, post.Excerpt, post.Content,
		featuredImageParam, post.Status, post.IsPublic, post.IsFeatured, post.IsCommentsEnabled,
		post.PublishedAt, post.ScheduledAt, tagsJSON, categoriesJSON,
		post.MetaTitle, post.MetaDescription, metaKeywords,
		post.ViewCount, post.LikeCount, post.CommentCount, post.ShareCount, post.ReadingTimeMinutes,
		post.CreatedAt, post.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}

	return nil
}

func (r *userPostRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.UserPost, error) {
	query := `
		SELECT 
			id, user_id, title, slug, excerpt, content,
			featured_image, status, is_public, is_featured, is_comments_enabled,
			published_at, scheduled_at, tags, categories,
			meta_title, meta_description, meta_keywords,
			view_count, like_count, comment_count, share_count, reading_time_minutes,
			deleted_at, created_at, updated_at
		FROM user_posts
		WHERE id = $1 AND deleted_at IS NULL
	`

	return r.scanPost(ctx, query, id)
}

func (r *userPostRepository) GetBySlug(ctx context.Context, userID uuidv7.UUID, slug string) (*entity.UserPost, error) {
	query := `
		SELECT 
			id, user_id, title, slug, excerpt, content,
			featured_image, status, is_public, is_featured, is_comments_enabled,
			published_at, scheduled_at, tags, categories,
			meta_title, meta_description, meta_keywords,
			view_count, like_count, comment_count, share_count, reading_time_minutes,
			deleted_at, created_at, updated_at
		FROM user_posts
		WHERE user_id = $1 AND slug = $2 AND deleted_at IS NULL
	`

	return r.scanPost(ctx, query, userID, slug)
}

func (r *userPostRepository) Update(ctx context.Context, post *entity.UserPost) error {
	featuredImage, err := entity.MarshalFeaturedImage(post.FeaturedImage)
	if err != nil {
		return fmt.Errorf("failed to marshal featured image: %w", err)
	}

	// Ensure nil slices are marshaled as empty arrays
	if post.Tags == nil {
		post.Tags = []string{}
	}
	if post.Categories == nil {
		post.Categories = []string{}
	}
	if post.MetaKeywords == nil {
		post.MetaKeywords = []string{}
	}

	tagsJSON, _ := json.Marshal(post.Tags)
	categoriesJSON, _ := json.Marshal(post.Categories)
	metaKeywords := pq.Array(post.MetaKeywords)

	// Convert featuredImage to interface{} to ensure NULL is sent for nil
	var featuredImageParam interface{}
	if len(featuredImage) > 0 {
		featuredImageParam = featuredImage
	} else {
		featuredImageParam = nil
	}

	query := `
		UPDATE user_posts SET
			title = $1, slug = $2, excerpt = $3, content = $4,
			featured_image = $5, status = $6, is_public = $7, is_featured = $8, is_comments_enabled = $9,
			published_at = $10, scheduled_at = $11, tags = $12, categories = $13,
			meta_title = $14, meta_description = $15, meta_keywords = $16,
			reading_time_minutes = $17, updated_at = $18
		WHERE id = $19 AND deleted_at IS NULL
	`

	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query,
		post.Title, post.Slug, post.Excerpt, post.Content,
		featuredImageParam, post.Status, post.IsPublic, post.IsFeatured, post.IsCommentsEnabled,
		post.PublishedAt, post.ScheduledAt, tagsJSON, categoriesJSON,
		post.MetaTitle, post.MetaDescription, metaKeywords,
		post.ReadingTimeMinutes, time.Now(), post.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return entity.ErrNotFound
	}

	return nil
}

func (r *userPostRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM user_posts WHERE id = $1`
	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return entity.ErrNotFound
	}

	return nil
}

func (r *userPostRepository) SoftDelete(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE user_posts SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete post: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return entity.ErrNotFound
	}

	return nil
}

func (r *userPostRepository) Restore(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE user_posts SET deleted_at = NULL, updated_at = NOW() WHERE id = $1 AND deleted_at IS NOT NULL`
	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to restore post: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return entity.ErrNotFound
	}

	return nil
}

func (r *userPostRepository) GetUserPosts(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.UserPost, error) {
	query := `
		SELECT 
			id, user_id, title, slug, excerpt, content,
			featured_image, status, is_public, is_featured, is_comments_enabled,
			published_at, scheduled_at, tags, categories,
			meta_title, meta_description, meta_keywords,
			view_count, like_count, comment_count, share_count, reading_time_minutes,
			deleted_at, created_at, updated_at
		FROM user_posts
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	return r.scanPosts(ctx, query, userID, limit, offset)
}

func (r *userPostRepository) GetPublishedPosts(ctx context.Context, limit, offset int) ([]*entity.UserPost, error) {
	query := `
		SELECT 
			id, user_id, title, slug, excerpt, content,
			featured_image, status, is_public, is_featured, is_comments_enabled,
			published_at, scheduled_at, tags, categories,
			meta_title, meta_description, meta_keywords,
			view_count, like_count, comment_count, share_count, reading_time_minutes,
			deleted_at, created_at, updated_at
		FROM user_posts
		WHERE status = 'published' AND is_public = true AND deleted_at IS NULL
		ORDER BY published_at DESC
		LIMIT $1 OFFSET $2
	`

	return r.scanPosts(ctx, query, limit, offset)
}

func (r *userPostRepository) GetFeaturedPosts(ctx context.Context, limit int) ([]*entity.UserPost, error) {
	query := `
		SELECT 
			id, user_id, title, slug, excerpt, content,
			featured_image, status, is_public, is_featured, is_comments_enabled,
			published_at, scheduled_at, tags, categories,
			meta_title, meta_description, meta_keywords,
			view_count, like_count, comment_count, share_count, reading_time_minutes,
			deleted_at, created_at, updated_at
		FROM user_posts
		WHERE is_featured = true AND status = 'published' AND is_public = true AND deleted_at IS NULL
		ORDER BY published_at DESC
		LIMIT $1
	`

	return r.scanPosts(ctx, query, limit)
}

func (r *userPostRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entity.UserPost, error) {
	sqlQuery := `
		SELECT 
			id, user_id, title, slug, excerpt, content,
			featured_image, status, is_public, is_featured, is_comments_enabled,
			published_at, scheduled_at, tags, categories,
			meta_title, meta_description, meta_keywords,
			view_count, like_count, comment_count, share_count, reading_time_minutes,
			deleted_at, created_at, updated_at
		FROM user_posts
		WHERE (
			title ILIKE $1 OR 
			content ILIKE $1 OR 
			excerpt ILIKE $1
		) AND status = 'published' AND is_public = true AND deleted_at IS NULL
		ORDER BY published_at DESC
		LIMIT $2 OFFSET $3
	`

	searchTerm := "%" + query + "%"
	return r.scanPosts(ctx, sqlQuery, searchTerm, limit, offset)
}

func (r *userPostRepository) GetByTag(ctx context.Context, tag string, limit, offset int) ([]*entity.UserPost, error) {
	query := `
		SELECT 
			id, user_id, title, slug, excerpt, content,
			featured_image, status, is_public, is_featured, is_comments_enabled,
			published_at, scheduled_at, tags, categories,
			meta_title, meta_description, meta_keywords,
			view_count, like_count, comment_count, share_count, reading_time_minutes,
			deleted_at, created_at, updated_at
		FROM user_posts
		WHERE tags @> $1::jsonb AND status = 'published' AND is_public = true AND deleted_at IS NULL
		ORDER BY published_at DESC
		LIMIT $2 OFFSET $3
	`

	tagJSON := fmt.Sprintf(`["%s"]`, tag)
	return r.scanPosts(ctx, query, tagJSON, limit, offset)
}

func (r *userPostRepository) GetScheduledPosts(ctx context.Context) ([]*entity.UserPost, error) {
	query := `
		SELECT 
			id, user_id, title, slug, excerpt, content,
			featured_image, status, is_public, is_featured, is_comments_enabled,
			published_at, scheduled_at, tags, categories,
			meta_title, meta_description, meta_keywords,
			view_count, like_count, comment_count, share_count, reading_time_minutes,
			deleted_at, created_at, updated_at
		FROM user_posts
		WHERE status = 'scheduled' AND scheduled_at <= NOW() AND deleted_at IS NULL
		ORDER BY scheduled_at ASC
	`

	return r.scanPosts(ctx, query)
}

func (r *userPostRepository) IncrementViews(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE user_posts SET view_count = view_count + 1 WHERE id = $1`
	return r.Exec(ctx, query, id)
}

func (r *userPostRepository) IncrementLikes(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE user_posts SET like_count = like_count + 1, updated_at = NOW() WHERE id = $1`
	return r.Exec(ctx, query, id)
}

func (r *userPostRepository) DecrementLikes(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE user_posts SET like_count = GREATEST(like_count - 1, 0), updated_at = NOW() WHERE id = $1`
	return r.Exec(ctx, query, id)
}

func (r *userPostRepository) IncrementComments(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE user_posts SET comment_count = comment_count + 1, updated_at = NOW() WHERE id = $1`
	return r.Exec(ctx, query, id)
}

func (r *userPostRepository) DecrementComments(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE user_posts SET comment_count = GREATEST(comment_count - 1, 0), updated_at = NOW() WHERE id = $1`
	return r.Exec(ctx, query, id)
}

func (r *userPostRepository) IncrementShares(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE user_posts SET share_count = share_count + 1, updated_at = NOW() WHERE id = $1`
	return r.Exec(ctx, query, id)
}

func (r *userPostRepository) UpdateStatus(ctx context.Context, id uuidv7.UUID, status entity.PostStatus) error {
	query := `UPDATE user_posts SET status = $1, updated_at = NOW() WHERE id = $2`
	return r.Exec(ctx, query, status, id)
}

func (r *userPostRepository) PublishScheduledPost(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE user_posts 
		SET status = 'published', published_at = NOW(), scheduled_at = NULL, updated_at = NOW()
		WHERE id = $1 AND status = 'scheduled'
	`
	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return entity.ErrNotFound
	}

	return nil
}

func (r *userPostRepository) List(ctx context.Context, params repository.ListPostsParams) ([]*entity.UserPost, *pagination.Metadata, error) {
	// Build dynamic query based on filters
	conditions := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	argCount := 1

	if params.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argCount))
		args = append(args, *params.UserID)
		argCount++
	}

	if params.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argCount))
		args = append(args, *params.Status)
		argCount++
	}

	if params.IsPublic != nil {
		conditions = append(conditions, fmt.Sprintf("is_public = $%d", argCount))
		args = append(args, *params.IsPublic)
		argCount++
	}

	if params.IsFeatured != nil {
		conditions = append(conditions, fmt.Sprintf("is_featured = $%d", argCount))
		args = append(args, *params.IsFeatured)
		argCount++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Count total
	countQuery := "SELECT COUNT(*) FROM user_posts " + whereClause
	var total int
	err := r.Get(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, nil, err
	}

	// Get posts
	sortBy := "created_at"
	if params.SortBy != "" {
		sortBy = params.SortBy
	}
	sortOrder := "DESC"
	if params.SortOrder != "" {
		sortOrder = strings.ToUpper(params.SortOrder)
	}

	queryTemplate := `
		SELECT 
			id, user_id, title, slug, excerpt, content,
			featured_image, status, is_public, is_featured, is_comments_enabled,
			published_at, scheduled_at, tags, categories,
			meta_title, meta_description, meta_keywords,
			view_count, like_count, comment_count, share_count, reading_time_minutes,
			deleted_at, created_at, updated_at
		FROM user_posts
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`

	query := fmt.Sprintf(queryTemplate, whereClause, sortBy, sortOrder, argCount, argCount+1)

	args = append(args, params.Limit, params.Offset)

	posts, err := r.scanPosts(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}

	meta := pagination.NewMetadata(total, params.Limit, params.Offset)

	return posts, meta, nil
}

// Helper methods

func (r *userPostRepository) scanPost(ctx context.Context, query string, args ...interface{}) (*entity.UserPost, error) {
	var post entity.UserPost
	var featuredImageJSON []byte
	var tagsJSON, categoriesJSON []byte

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&post.ID, &post.UserID, &post.Title, &post.Slug, &post.Excerpt, &post.Content,
		&featuredImageJSON, &post.Status, &post.IsPublic, &post.IsFeatured, &post.IsCommentsEnabled,
		&post.PublishedAt, &post.ScheduledAt, &tagsJSON, &categoriesJSON,
		&post.MetaTitle, &post.MetaDescription, pq.Array(&post.MetaKeywords),
		&post.ViewCount, &post.LikeCount, &post.CommentCount, &post.ShareCount, &post.ReadingTimeMinutes,
		&post.DeletedAt, &post.CreatedAt, &post.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan post: %w", err)
	}

	// Unmarshal JSONB fields
	if featuredImageJSON != nil {
		post.FeaturedImage, _ = entity.UnmarshalFeaturedImage(featuredImageJSON)
	}

	if tagsJSON != nil {
		_ = json.Unmarshal(tagsJSON, &post.Tags)
	}

	if categoriesJSON != nil {
		_ = json.Unmarshal(categoriesJSON, &post.Categories)
	}

	return &post, nil
}

func (r *userPostRepository) scanPosts(ctx context.Context, query string, args ...interface{}) ([]*entity.UserPost, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}
	defer rows.Close()

	var posts []*entity.UserPost

	for rows.Next() {
		var post entity.UserPost
		var featuredImageJSON []byte
		var tagsJSON, categoriesJSON []byte

		err := rows.Scan(
			&post.ID, &post.UserID, &post.Title, &post.Slug, &post.Excerpt, &post.Content,
			&featuredImageJSON, &post.Status, &post.IsPublic, &post.IsFeatured, &post.IsCommentsEnabled,
			&post.PublishedAt, &post.ScheduledAt, &tagsJSON, &categoriesJSON,
			&post.MetaTitle, &post.MetaDescription, pq.Array(&post.MetaKeywords),
			&post.ViewCount, &post.LikeCount, &post.CommentCount, &post.ShareCount, &post.ReadingTimeMinutes,
			&post.DeletedAt, &post.CreatedAt, &post.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}

		// Unmarshal JSONB fields
		if featuredImageJSON != nil {
			post.FeaturedImage, _ = entity.UnmarshalFeaturedImage(featuredImageJSON)
		}

		if tagsJSON != nil {
			_ = json.Unmarshal(tagsJSON, &post.Tags)
		}

		if categoriesJSON != nil {
			_ = json.Unmarshal(categoriesJSON, &post.Categories)
		}

		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return posts, nil
}
