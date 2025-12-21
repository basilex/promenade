package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/jmoiron/sqlx"
)

type purgeRepository struct {
	*BaseRepository
}

// NewPurgeRepository creates a new purge repository
func NewPurgeRepository(db *sqlx.DB) repository.PurgeRepository {
	return &purgeRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// PurgeUserPosts permanently deletes user_posts that were soft-deleted before cutoffDate
func (r *purgeRepository) PurgeUserPosts(ctx context.Context, cutoffDate time.Time, batchSize int, dryRun bool) (int64, error) {
	log := logger.FromContext(ctx)

	if dryRun {
		count, err := r.CountDeletableUserPosts(ctx, cutoffDate)
		if err != nil {
			return 0, err
		}
		log.Info("[DRY RUN] Would purge user_posts", "count", count)
		return count, nil
	}

	var totalDeleted int64
	executor := r.getExecutor(ctx)

	for {
		query := `
			DELETE FROM user_posts 
			WHERE id IN (
				SELECT id FROM user_posts 
				WHERE deleted_at IS NOT NULL 
				  AND deleted_at < $1
				ORDER BY deleted_at ASC
				LIMIT $2
			)
		`

		result, err := executor.ExecContext(ctx, query, cutoffDate, batchSize)
		if err != nil {
			log.Error("Failed to purge user_posts batch", "error", err)
			return totalDeleted, fmt.Errorf("failed to purge user_posts: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to get rows affected: %w", err)
		}

		totalDeleted += rowsAffected

		log.Info("Purged user_posts batch",
			"batch_size", rowsAffected,
			"total_deleted", totalDeleted,
		)

		// If we deleted fewer rows than the batch size, we're done
		if rowsAffected < int64(batchSize) {
			break
		}

		// Small delay between batches to avoid overwhelming the database
		time.Sleep(100 * time.Millisecond)
	}

	return totalDeleted, nil
}

// PurgePostComments permanently deletes post_comments that were soft-deleted before cutoffDate
func (r *purgeRepository) PurgePostComments(ctx context.Context, cutoffDate time.Time, batchSize int, dryRun bool) (int64, error) {
	log := logger.FromContext(ctx)

	if dryRun {
		count, err := r.CountDeletablePostComments(ctx, cutoffDate)
		if err != nil {
			return 0, err
		}
		log.Info("[DRY RUN] Would purge post_comments", "count", count)
		return count, nil
	}

	var totalDeleted int64
	executor := r.getExecutor(ctx)

	for {
		query := `
			DELETE FROM post_comments 
			WHERE id IN (
				SELECT id FROM post_comments 
				WHERE deleted_at IS NOT NULL 
				  AND deleted_at < $1
				ORDER BY deleted_at ASC
				LIMIT $2
			)
		`

		result, err := executor.ExecContext(ctx, query, cutoffDate, batchSize)
		if err != nil {
			log.Error("Failed to purge post_comments batch", "error", err)
			return totalDeleted, fmt.Errorf("failed to purge post_comments: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to get rows affected: %w", err)
		}

		totalDeleted += rowsAffected

		log.Info("Purged post_comments batch",
			"batch_size", rowsAffected,
			"total_deleted", totalDeleted,
		)

		// If we deleted fewer rows than the batch size, we're done
		if rowsAffected < int64(batchSize) {
			break
		}

		// Small delay between batches to avoid overwhelming the database
		time.Sleep(100 * time.Millisecond)
	}

	return totalDeleted, nil
}

// CountDeletableUserPosts returns the count of user_posts eligible for purging
func (r *purgeRepository) CountDeletableUserPosts(ctx context.Context, cutoffDate time.Time) (int64, error) {
	query := `
		SELECT COUNT(*) 
		FROM user_posts 
		WHERE deleted_at IS NOT NULL 
		  AND deleted_at < $1
	`

	executor := r.getExecutor(ctx)
	var count int64
	err := sqlx.GetContext(ctx, executor, &count, query, cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("failed to count deletable user_posts: %w", err)
	}

	return count, nil
}

// CountDeletablePostComments returns the count of post_comments eligible for purging
func (r *purgeRepository) CountDeletablePostComments(ctx context.Context, cutoffDate time.Time) (int64, error) {
	query := `
		SELECT COUNT(*) 
		FROM post_comments 
		WHERE deleted_at IS NOT NULL 
		  AND deleted_at < $1
	`

	executor := r.getExecutor(ctx)
	var count int64
	err := sqlx.GetContext(ctx, executor, &count, query, cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("failed to count deletable post_comments: %w", err)
	}

	return count, nil
}
