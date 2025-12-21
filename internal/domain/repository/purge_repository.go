package repository

import (
	"context"
	"time"
)

// PurgeRepository defines interface for permanently deleting soft-deleted records
type PurgeRepository interface {
	// PurgeUserPosts permanently deletes user_posts that were soft-deleted before cutoffDate
	// Returns the number of records deleted
	PurgeUserPosts(ctx context.Context, cutoffDate time.Time, batchSize int, dryRun bool) (int64, error)

	// PurgePostComments permanently deletes post_comments that were soft-deleted before cutoffDate
	// Returns the number of records deleted
	PurgePostComments(ctx context.Context, cutoffDate time.Time, batchSize int, dryRun bool) (int64, error)

	// CountDeletableUserPosts returns the count of user_posts eligible for purging
	CountDeletableUserPosts(ctx context.Context, cutoffDate time.Time) (int64, error)

	// CountDeletablePostComments returns the count of post_comments eligible for purging
	CountDeletablePostComments(ctx context.Context, cutoffDate time.Time) (int64, error)
}
