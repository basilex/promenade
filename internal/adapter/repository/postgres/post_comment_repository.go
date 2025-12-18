package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type postCommentRepository struct {
	*BaseRepository
}

// NewPostCommentRepository creates a new post comment repository
func NewPostCommentRepository(db *sqlx.DB) repository.PostCommentRepository {
	return &postCommentRepository{
		BaseRepository: &BaseRepository{db: db},
	}
}

func (r *postCommentRepository) Create(ctx context.Context, comment *entity.PostComment) error {
	query := `
		INSERT INTO post_comments (
			id, post_id, user_id, parent_id, content, 
			is_edited, like_count, reply_count, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	return r.Exec(ctx, query,
		comment.ID, comment.PostID, comment.UserID, comment.ParentID, comment.Content,
		comment.IsEdited, comment.LikeCount, comment.ReplyCount, comment.CreatedAt, comment.UpdatedAt,
	)
}

func (r *postCommentRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.PostComment, error) {
	query := `
		SELECT id, post_id, user_id, parent_id, content,
			is_edited, edited_at, like_count, reply_count,
			deleted_at, created_at, updated_at
		FROM post_comments
		WHERE id = $1
	`

	var comment entity.PostComment
	if err := r.Get(ctx, &comment, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	return &comment, nil
}

func (r *postCommentRepository) Update(ctx context.Context, comment *entity.PostComment) error {
	query := `
		UPDATE post_comments
		SET content = $1, is_edited = $2, edited_at = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`

	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query,
		comment.Content, comment.IsEdited, comment.EditedAt, comment.UpdatedAt, comment.ID,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return entity.ErrNotFound
	}

	return nil
}

func (r *postCommentRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM post_comments WHERE id = $1`
	return r.Exec(ctx, query, id)
}

func (r *postCommentRepository) SoftDelete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE post_comments
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
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

func (r *postCommentRepository) Restore(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE post_comments
		SET deleted_at = NULL, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NOT NULL
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

func (r *postCommentRepository) GetPostComments(ctx context.Context, postID uuidv7.UUID, limit, offset int) ([]*entity.PostComment, *pagination.Metadata, error) {
	// Count total comments
	countQuery := `SELECT COUNT(*) FROM post_comments WHERE post_id = $1 AND deleted_at IS NULL AND parent_id IS NULL`
	var total int
	if err := r.Get(ctx, &total, countQuery, postID); err != nil {
		return nil, nil, err
	}

	// Get paginated comments
	query := `
		SELECT id, post_id, user_id, parent_id, content,
			is_edited, edited_at, like_count, reply_count,
			deleted_at, created_at, updated_at
		FROM post_comments
		WHERE post_id = $1 AND deleted_at IS NULL AND parent_id IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var comments []*entity.PostComment
	if err := r.Select(ctx, &comments, query, postID, limit, offset); err != nil {
		return nil, nil, err
	}

	meta := pagination.NewMetadata(total, limit, offset)
	return comments, meta, nil
}

func (r *postCommentRepository) GetCommentReplies(ctx context.Context, parentID uuidv7.UUID, limit, offset int) ([]*entity.PostComment, *pagination.Metadata, error) {
	// Count total replies
	countQuery := `SELECT COUNT(*) FROM post_comments WHERE parent_id = $1 AND deleted_at IS NULL`
	var total int
	if err := r.Get(ctx, &total, countQuery, parentID); err != nil {
		return nil, nil, err
	}

	// Get paginated replies
	query := `
		SELECT id, post_id, user_id, parent_id, content,
			is_edited, edited_at, like_count, reply_count,
			deleted_at, created_at, updated_at
		FROM post_comments
		WHERE parent_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	var comments []*entity.PostComment
	if err := r.Select(ctx, &comments, query, parentID, limit, offset); err != nil {
		return nil, nil, err
	}

	meta := pagination.NewMetadata(total, limit, offset)
	return comments, meta, nil
}

func (r *postCommentRepository) GetUserComments(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.PostComment, *pagination.Metadata, error) {
	// Count total comments
	countQuery := `SELECT COUNT(*) FROM post_comments WHERE user_id = $1 AND deleted_at IS NULL`
	var total int
	if err := r.Get(ctx, &total, countQuery, userID); err != nil {
		return nil, nil, err
	}

	// Get paginated comments
	query := `
		SELECT id, post_id, user_id, parent_id, content,
			is_edited, edited_at, like_count, reply_count,
			deleted_at, created_at, updated_at
		FROM post_comments
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var comments []*entity.PostComment
	if err := r.Select(ctx, &comments, query, userID, limit, offset); err != nil {
		return nil, nil, err
	}

	meta := pagination.NewMetadata(total, limit, offset)
	return comments, meta, nil
}

func (r *postCommentRepository) CountPostComments(ctx context.Context, postID uuidv7.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM post_comments WHERE post_id = $1 AND deleted_at IS NULL`
	var count int
	if err := r.Get(ctx, &count, query, postID); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *postCommentRepository) CountUserComments(ctx context.Context, userID uuidv7.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM post_comments WHERE user_id = $1 AND deleted_at IS NULL`
	var count int
	if err := r.Get(ctx, &count, query, userID); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *postCommentRepository) IncrementLikes(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE post_comments SET like_count = like_count + 1, updated_at = NOW() WHERE id = $1`
	return r.Exec(ctx, query, id)
}

func (r *postCommentRepository) DecrementLikes(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE post_comments SET like_count = GREATEST(like_count - 1, 0), updated_at = NOW() WHERE id = $1`
	return r.Exec(ctx, query, id)
}

func (r *postCommentRepository) IncrementReplies(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE post_comments SET reply_count = reply_count + 1, updated_at = NOW() WHERE id = $1`
	return r.Exec(ctx, query, id)
}

func (r *postCommentRepository) DecrementReplies(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE post_comments SET reply_count = GREATEST(reply_count - 1, 0), updated_at = NOW() WHERE id = $1`
	return r.Exec(ctx, query, id)
}
