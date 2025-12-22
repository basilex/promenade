package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetentionPolicy_GetCutoffDate(t *testing.T) {
	tests := []struct {
		name          string
		retentionDays int
		wantDaysBefore int
	}{
		{
			name:          "30 days retention",
			retentionDays: 30,
			wantDaysBefore: 30,
		},
		{
			name:          "90 days retention",
			retentionDays: 90,
			wantDaysBefore: 90,
		},
		{
			name:          "7 days retention",
			retentionDays: 7,
			wantDaysBefore: 7,
		},
		{
			name:          "1 day retention",
			retentionDays: 1,
			wantDaysBefore: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &RetentionPolicy{
				EntityName:    "test_entity",
				RetentionDays: tt.retentionDays,
				Enabled:       true,
			}

			cutoffDate := policy.GetCutoffDate()
			expectedDate := time.Now().AddDate(0, 0, -tt.wantDaysBefore)

			// Allow 1 second tolerance for test execution time
			diff := expectedDate.Sub(cutoffDate)
			assert.LessOrEqual(t, diff.Abs().Seconds(), 1.0)
		})
	}
}

func TestRetentionPolicy_Fields(t *testing.T) {
	now := time.Now()
	policy := &RetentionPolicy{
		EntityName:    "user_posts",
		RetentionDays: 30,
		Enabled:       true,
		LastRun:       &now,
		RecordsPurged: 42,
	}

	assert.Equal(t, "user_posts", policy.EntityName)
	assert.Equal(t, 30, policy.RetentionDays)
	assert.True(t, policy.Enabled)
	assert.NotNil(t, policy.LastRun)
	assert.Equal(t, int64(42), policy.RecordsPurged)
}

func TestPurgeResult_Structure(t *testing.T) {
	duration := 5 * time.Second
	now := time.Now()

	result := &PurgeResult{
		EntityName:    "user_posts",
		RecordsPurged: 100,
		Duration:      duration,
		DryRun:        false,
		Error:         "",
		Timestamp:     now,
	}

	assert.Equal(t, "user_posts", result.EntityName)
	assert.Equal(t, int64(100), result.RecordsPurged)
	assert.Equal(t, duration, result.Duration)
	assert.False(t, result.DryRun)
	assert.Empty(t, result.Error)
	assert.Equal(t, now, result.Timestamp)
}

func TestPurgeResult_WithError(t *testing.T) {
	result := &PurgeResult{
		EntityName:    "user_posts",
		RecordsPurged: 0,
		Duration:      1 * time.Second,
		DryRun:        false,
		Error:         "database connection failed",
		Timestamp:     time.Now(),
	}

	assert.Equal(t, int64(0), result.RecordsPurged)
	assert.NotEmpty(t, result.Error)
	assert.Contains(t, result.Error, "database connection failed")
}

func TestPurgeResult_DryRun(t *testing.T) {
	result := &PurgeResult{
		EntityName:    "user_posts",
		RecordsPurged: 50,
		Duration:      2 * time.Second,
		DryRun:        true,
		Timestamp:     time.Now(),
	}

	assert.True(t, result.DryRun)
	assert.Equal(t, int64(50), result.RecordsPurged)
}

func TestPurgeSummary_SingleResult(t *testing.T) {
	now := time.Now()
	result := PurgeResult{
		EntityName:    "user_posts",
		RecordsPurged: 100,
		Duration:      5 * time.Second,
		DryRun:        false,
		Timestamp:     now,
	}

	summary := &PurgeSummary{
		TotalRecordsPurged: 100,
		Results:            []PurgeResult{result},
		Duration:           5 * time.Second,
		DryRun:             false,
		Timestamp:          now,
	}

	assert.Equal(t, int64(100), summary.TotalRecordsPurged)
	assert.Len(t, summary.Results, 1)
	assert.Equal(t, 5*time.Second, summary.Duration)
	assert.False(t, summary.DryRun)
}

func TestPurgeSummary_MultipleResults(t *testing.T) {
	now := time.Now()
	results := []PurgeResult{
		{
			EntityName:    "user_posts",
			RecordsPurged: 100,
			Duration:      3 * time.Second,
			DryRun:        false,
			Timestamp:     now,
		},
		{
			EntityName:    "post_comments",
			RecordsPurged: 250,
			Duration:      5 * time.Second,
			DryRun:        false,
			Timestamp:     now,
		},
		{
			EntityName:    "comment_likes",
			RecordsPurged: 500,
			Duration:      2 * time.Second,
			DryRun:        false,
			Timestamp:     now,
		},
	}

	totalPurged := int64(850)
	totalDuration := 10 * time.Second

	summary := &PurgeSummary{
		TotalRecordsPurged: totalPurged,
		Results:            results,
		Duration:           totalDuration,
		DryRun:             false,
		Timestamp:          now,
	}

	assert.Equal(t, int64(850), summary.TotalRecordsPurged)
	assert.Len(t, summary.Results, 3)
	assert.Equal(t, 10*time.Second, summary.Duration)
	assert.False(t, summary.DryRun)

	// Verify individual results
	assert.Equal(t, "user_posts", summary.Results[0].EntityName)
	assert.Equal(t, int64(100), summary.Results[0].RecordsPurged)
	assert.Equal(t, "post_comments", summary.Results[1].EntityName)
	assert.Equal(t, int64(250), summary.Results[1].RecordsPurged)
	assert.Equal(t, "comment_likes", summary.Results[2].EntityName)
	assert.Equal(t, int64(500), summary.Results[2].RecordsPurged)
}

func TestPurgeSummary_DryRun(t *testing.T) {
	now := time.Now()
	results := []PurgeResult{
		{
			EntityName:    "user_posts",
			RecordsPurged: 100,
			Duration:      1 * time.Second,
			DryRun:        true,
			Timestamp:     now,
		},
	}

	summary := &PurgeSummary{
		TotalRecordsPurged: 100,
		Results:            results,
		Duration:           1 * time.Second,
		DryRun:             true,
		Timestamp:          now,
	}

	assert.True(t, summary.DryRun)
	assert.Equal(t, int64(100), summary.TotalRecordsPurged)
}

func TestPurgeSummary_WithErrors(t *testing.T) {
	now := time.Now()
	results := []PurgeResult{
		{
			EntityName:    "user_posts",
			RecordsPurged: 100,
			Duration:      2 * time.Second,
			DryRun:        false,
			Timestamp:     now,
		},
		{
			EntityName:    "post_comments",
			RecordsPurged: 0,
			Duration:      1 * time.Second,
			DryRun:        false,
			Error:         "database timeout",
			Timestamp:     now,
		},
	}

	summary := &PurgeSummary{
		TotalRecordsPurged: 100,
		Results:            results,
		Duration:           3 * time.Second,
		DryRun:             false,
		Timestamp:          now,
	}

	assert.Equal(t, int64(100), summary.TotalRecordsPurged)
	assert.Len(t, summary.Results, 2)

	// First result succeeded
	assert.Empty(t, summary.Results[0].Error)
	assert.Equal(t, int64(100), summary.Results[0].RecordsPurged)

	// Second result failed
	assert.NotEmpty(t, summary.Results[1].Error)
	assert.Equal(t, int64(0), summary.Results[1].RecordsPurged)
	assert.Contains(t, summary.Results[1].Error, "database timeout")
}

func TestRetentionPolicy_DisabledPolicy(t *testing.T) {
	policy := &RetentionPolicy{
		EntityName:    "user_posts",
		RetentionDays: 30,
		Enabled:       false,
	}

	assert.False(t, policy.Enabled)
	assert.Equal(t, 30, policy.RetentionDays)

	// GetCutoffDate should still work even if disabled
	cutoffDate := policy.GetCutoffDate()
	assert.NotZero(t, cutoffDate)
}
