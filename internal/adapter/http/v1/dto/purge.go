package dto

import (
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
)

// PurgeRequest represents a manual purge request
type PurgeRequest struct {
	EntityName string `json:"entity_name" binding:"omitempty,oneof=user_posts post_comments"` // Empty means all
	DryRun     bool   `json:"dry_run"`
}

// PurgeResponse represents the result of a purge operation
type PurgeResponse struct {
	Success            bool                  `json:"success"`
	Message            string                `json:"message"`
	TotalRecordsPurged int64                 `json:"total_records_purged"`
	Results            []PurgeResultResponse `json:"results"`
	Duration           string                `json:"duration"`
	DryRun             bool                  `json:"dry_run"`
	Timestamp          time.Time             `json:"timestamp"`
}

// PurgeResultResponse represents individual entity purge result
type PurgeResultResponse struct {
	EntityName    string `json:"entity_name"`
	RecordsPurged int64  `json:"records_purged"`
	Duration      string `json:"duration"`
	Error         string `json:"error,omitempty"`
}

// RetentionPolicyResponse represents a retention policy
type RetentionPolicyResponse struct {
	EntityName    string     `json:"entity_name"`
	RetentionDays int        `json:"retention_days"`
	Enabled       bool       `json:"enabled"`
	LastRun       *time.Time `json:"last_run,omitempty"`
	RecordsPurged int64      `json:"records_purged,omitempty"`
}

// PurgePreviewResponse represents a preview of what would be purged
type PurgePreviewResponse struct {
	EntityName string `json:"entity_name"`
	Count      int64  `json:"count"`
	CutoffDate string `json:"cutoff_date"`
}

// SchedulerStatusResponse represents scheduler status
type SchedulerStatusResponse struct {
	Enabled       bool       `json:"enabled"`
	Schedule      string     `json:"schedule"`
	DryRun        bool       `json:"dry_run"`
	IsRunning     bool       `json:"is_running"`
	LastRunTime   *time.Time `json:"last_run_time,omitempty"`
	LastRunStatus string     `json:"last_run_status,omitempty"`
}

// ToPurgeResponse converts PurgeSummary to PurgeResponse
func ToPurgeResponse(summary *entity.PurgeSummary) *PurgeResponse {
	results := make([]PurgeResultResponse, len(summary.Results))
	for i, result := range summary.Results {
		results[i] = PurgeResultResponse{
			EntityName:    result.EntityName,
			RecordsPurged: result.RecordsPurged,
			Duration:      result.Duration.String(),
			Error:         result.Error,
		}
	}

	return &PurgeResponse{
		Success:            true,
		Message:            "Purge operation completed successfully",
		TotalRecordsPurged: summary.TotalRecordsPurged,
		Results:            results,
		Duration:           summary.Duration.String(),
		DryRun:             summary.DryRun,
		Timestamp:          summary.Timestamp,
	}
}

// ToRetentionPolicyResponse converts RetentionPolicy to RetentionPolicyResponse
func ToRetentionPolicyResponse(policy *entity.RetentionPolicy) *RetentionPolicyResponse {
	return &RetentionPolicyResponse{
		EntityName:    policy.EntityName,
		RetentionDays: policy.RetentionDays,
		Enabled:       policy.Enabled,
		LastRun:       policy.LastRun,
		RecordsPurged: policy.RecordsPurged,
	}
}

// ToRetentionPolicyListResponse converts slice of RetentionPolicy to response slice
func ToRetentionPolicyListResponse(policies []entity.RetentionPolicy) []RetentionPolicyResponse {
	responses := make([]RetentionPolicyResponse, len(policies))
	for i, policy := range policies {
		responses[i] = *ToRetentionPolicyResponse(&policy)
	}
	return responses
}
