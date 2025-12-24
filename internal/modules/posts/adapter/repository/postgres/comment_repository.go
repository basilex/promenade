package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/modules/posts/domain/entity"
	"github.com/basilex/promenade/internal/modules/posts/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type commentRepository struct {
	db *sqlx.DB
}

// NewCommentRepository creates a new comment repository
func NewCommentRepository(db *sqlx.DB) repository.ICommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *entity.Comment) error {
	query := `
		INSERT INTO posts_comments (
			id, post_id, user_id, parent_id, content, depth, path,
			created_at, updated_at, likes_count, replies_count
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		comment.ID,
		comment.PostID,
		comment.UserID,
		comment.ParentID,
		comment.Content,
		comment.Depth,
		comment.Path,
		comment.CreatedAt,
		comment.UpdatedAt,
		comment.LikesCount,
		comment.RepliesCount,
	)

	return err
}

func (r *commentRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Comment, error) {
	query := `
		SELECT 
			id, post_id, user_id, parent_id, content, depth, path,
			created_at, updated_at, deleted_at, likes_count, replies_count
		FROM posts_comments
		WHERE id = $1 AND deleted_at IS NULL
	`

	var comment entity.Comment
	err := r.db.GetContext(ctx, &comment, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("comment not found")
	}
	if err != nil {
		return nil, err
	}

	return &comment, nil
}

func (r *commentRepository) Update(ctx context.Context, comment *entity.Comment) error {
	query := `
		UPDATE posts_comments
		SET content = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		comment.Content,
		comment.UpdatedAt,
		comment.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

func (r *commentRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE posts_comments
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}

func (r *commentRepository) GetPostComments(ctx context.Context, postID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error) {
	// Count total
	var total int64
	countQuery := `SELECT COUNT(*) FROM posts_comments WHERE post_id = $1 AND deleted_at IS NULL AND parent_id IS NULL`
	if err := r.db.GetContext(ctx, &total, countQuery, postID); err != nil {
		return nil, nil, err
	}

	// Get comments
	query := `
		SELECT 
			id, post_id, user_id, parent_id, content, depth, path,
			created_at, updated_at, deleted_at, likes_count, replies_count
		FROM posts_comments
		WHERE post_id = $1 AND deleted_at IS NULL AND parent_id IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var comments []*entity.Comment
	if err := r.db.SelectContext(ctx, &comments, query, postID, limit, offset); err != nil {
		return nil, nil, err
	}

	meta := pagination.NewMetadata(int(total), limit, offset)
	return comments, meta, nil
}

func (r *commentRepository) GetReplies(ctx context.Context, parentID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error) {
	// Count total
	var total int64
	countQuery := `SELECT COUNT(*) FROM posts_comments WHERE parent_id = $1 AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &total, countQuery, parentID); err != nil {
		return nil, nil, err
	}

	// Get replies
	query := `
		SELECT 
			id, post_id, user_id, parent_id, content, depth, path,
			created_at, updated_at, deleted_at, likes_count, replies_count
		FROM posts_comments
		WHERE parent_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	var comments []*entity.Comment
	if err := r.db.SelectContext(ctx, &comments, query, parentID, limit, offset); err != nil {
		return nil, nil, err
	}

	meta := pagination.NewMetadata(int(total), limit, offset)
	return comments, meta, nil
}

func (r *commentRepository) GetUserComments(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error) {
	// Count total
	var total int64
	countQuery := `SELECT COUNT(*) FROM posts_comments WHERE user_id = $1 AND deleted_at IS NULL`
	if err := r.db.GetContext(ctx, &total, countQuery, userID); err != nil {
		return nil, nil, err
	}

	// Get comments
	query := `
		SELECT 
			id, post_id, user_id, parent_id, content, depth, path,
			created_at, updated_at, deleted_at, likes_count, replies_count
		FROM posts_comments
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var comments []*entity.Comment
	if err := r.db.SelectContext(ctx, &comments, query, userID, limit, offset); err != nil {
		return nil, nil, err
	}

	meta := pagination.NewMetadata(int(total), limit, offset)
	return comments, meta, nil
}

func (r *commentRepository) CountPostComments(ctx context.Context, postID uuidv7.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM posts_comments WHERE post_id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query, postID)
	return count, err
}

func (r *commentRepository) IncrementRepliesCount(ctx context.Context, parentID uuidv7.UUID) error {
	query := `UPDATE posts_comments SET replies_count = replies_count + 1 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, parentID)
	return err
}

func (r *commentRepository) DecrementRepliesCount(ctx context.Context, parentID uuidv7.UUID) error {
	query := `UPDATE posts_comments SET replies_count = GREATEST(replies_count - 1, 0) WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, parentID)
	return err
}
