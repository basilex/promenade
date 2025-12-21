package entity

import "time"

// RetentionPolicy defines how long soft-deleted records should be kept before permanent deletion
type RetentionPolicy struct {
	EntityName    string        `json:"entity_name"`
	RetentionDays int           `json:"retention_days"`
	Enabled       bool          `json:"enabled"`
	LastRun       *time.Time    `json:"last_run,omitempty"`
	RecordsPurged int64         `json:"records_purged,omitempty"`
}

// GetCutoffDate returns the date before which records should be purged
func (p *RetentionPolicy) GetCutoffDate() time.Time {
	return time.Now().AddDate(0, 0, -p.RetentionDays)
}

// PurgeResult represents the result of a purge operation
type PurgeResult struct {
	EntityName    string        `json:"entity_name"`
	RecordsPurged int64         `json:"records_purged"`
	Duration      time.Duration `json:"duration"`
	DryRun        bool          `json:"dry_run"`
	Error         string        `json:"error,omitempty"`
	Timestamp     time.Time     `json:"timestamp"`
}

// PurgeSummary aggregates results from multiple purge operations
type PurgeSummary struct {
	TotalRecordsPurged int64          `json:"total_records_purged"`
	Results            []PurgeResult  `json:"results"`
	Duration           time.Duration  `json:"duration"`
	DryRun             bool           `json:"dry_run"`
	Timestamp          time.Time      `json:"timestamp"`
}
