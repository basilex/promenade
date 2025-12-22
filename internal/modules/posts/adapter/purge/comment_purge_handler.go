package purge

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/pkg/logger"
)

// CommentPurgeHandler implements purge.Handler for post_comments table
type CommentPurgeHandler struct {
	db *sqlx.DB
}

// NewCommentPurgeHandler creates a new purge handler for comments
func NewCommentPurgeHandler(db *sqlx.DB) *CommentPurgeHandler {
	return &CommentPurgeHandler{
		db: db,
	}
}

// EntityName returns the entity name this handler manages
func (h *CommentPurgeHandler) EntityName() string {
	return "post_comments"
}

// Purge permanently deletes soft-deleted comments older than cutoffDate
func (h *CommentPurgeHandler) Purge(ctx context.Context, cutoffDate time.Time, batchSize int, dryRun bool) (int64, error) {
	log := logger.FromContext(ctx)

	if dryRun {
		// Count records that would be purged
		var count int64
		query := `
			SELECT COUNT(*) 
			FROM post_comments 
			WHERE deleted_at IS NOT NULL 
			  AND deleted_at < $1
		`
		if err := h.db.GetContext(ctx, &count, query, cutoffDate); err != nil {
			log.Error("Failed to count purgeable comments", "error", err)
			return 0, err
		}
		log.Info("Dry run: would purge comments", "count", count, "cutoff_date", cutoffDate)
		return count, nil
	}

	// Purge in batches to avoid locking the table for too long
	var totalPurged int64
	for {
		query := `
			DELETE FROM post_comments 
			WHERE id IN (
				SELECT id 
				FROM post_comments 
				WHERE deleted_at IS NOT NULL 
				  AND deleted_at < $1 
				LIMIT $2
			)
		`
		result, err := h.db.ExecContext(ctx, query, cutoffDate, batchSize)
		if err != nil {
			log.Error("Failed to purge comments batch", "error", err, "total_purged", totalPurged)
			return totalPurged, err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Error("Failed to get rows affected", "error", err)
			return totalPurged, err
		}

		totalPurged += rowsAffected

		if rowsAffected < int64(batchSize) {
			// No more records to purge
			break
		}

		log.Debug("Purged comments batch", "batch_size", rowsAffected, "total_purged", totalPurged)
	}

	log.Info("Successfully purged comments", "total_purged", totalPurged, "cutoff_date", cutoffDate)
	return totalPurged, nil
}
