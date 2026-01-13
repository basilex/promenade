package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// TestEngine_Integration_RealJobExecution tests complete job execution flow
func TestEngine_Integration_RealJobExecution(t *testing.T) {
	cfg := DefaultConfig()
	cfg.WorkerPoolSize = 2
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Register test executor
	executed := atomic.Int32{}
	executor := &MockExecutor{
		id: "test-executor",
		executeFunc: func(ctx context.Context, job *Job) error {
			executed.Add(1)
			return nil
		},
	}

	if err := engine.RegisterExecutor(JobTypeLUA, executor); err != nil {
		t.Fatalf("Failed to register executor: %v", err)
	}

	// Start engine
	if err := engine.Start(); err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}
	defer engine.Stop()

	// Add job with immediate execution (every second)
	job := &Job{
		ID:             uuidv7.New(),
		Name:           "Integration Test Job",
		Type:           JobTypeLUA,
		CronExpression: "@every 1s",
		Enabled:        true,
		Payload:        map[string]interface{}{"test": "value"},
	}

	if err := engine.AddJob(job); err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	// Wait for at least 2 executions
	time.Sleep(2500 * time.Millisecond)

	// Verify executions happened
	count := executed.Load()
	if count < 2 {
		t.Errorf("Expected at least 2 executions, got %d", count)
	}

	// Verify job exists and is enabled
	retrieved, err := engine.GetJob(job.ID)
	if err != nil {
		t.Errorf("Failed to get job: %v", err)
	}
	if !retrieved.Enabled {
		t.Error("Job should still be enabled")
	}
}

// TestEngine_Integration_JobFailureRetry tests retry mechanism
func TestEngine_Integration_JobFailureRetry(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DefaultRetryAttempts = 2
	cfg.DefaultRetryDelay = 100 * time.Millisecond
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Track execution attempts
	attempts := atomic.Int32{}
	executor := &MockExecutor{
		id: "retry-executor",
		executeFunc: func(ctx context.Context, job *Job) error {
			count := attempts.Add(1)
			if count < 3 {
				return ErrJobExecutionFailed // Fail first 2 attempts
			}
			return nil // Success on 3rd attempt
		},
	}

	if err := engine.RegisterExecutor(JobTypeLUA, executor); err != nil {
		t.Fatalf("Failed to register executor: %v", err)
	}

	if err := engine.Start(); err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}
	defer engine.Stop()

	// Add job
	job := &Job{
		ID:             uuidv7.New(),
		Name:           "Retry Test Job",
		Type:           JobTypeLUA,
		CronExpression: "@every 1s",
		Enabled:        true,
		Payload:        map[string]interface{}{},
	}

	if err := engine.AddJob(job); err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	// Wait for retries to complete
	time.Sleep(1500 * time.Millisecond)

	// Verify multiple attempts were made
	count := attempts.Load()
	if count < 3 {
		t.Errorf("Expected at least 3 attempts (initial + 2 retries), got %d", count)
	}
}

// TestEngine_Integration_MultipleJobs tests concurrent job execution
func TestEngine_Integration_MultipleJobs(t *testing.T) {
	cfg := DefaultConfig()
	cfg.WorkerPoolSize = 3
	cfg.MaxConcurrentJobs = 3
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Track which jobs executed
	job1Executed := atomic.Bool{}
	job2Executed := atomic.Bool{}
	job3Executed := atomic.Bool{}

	executor := &MockExecutor{
		id: "multi-executor",
		executeFunc: func(ctx context.Context, job *Job) error {
			jobName := job.Payload["name"].(string)
			switch jobName {
			case "job1":
				job1Executed.Store(true)
			case "job2":
				job2Executed.Store(true)
			case "job3":
				job3Executed.Store(true)
			}
			time.Sleep(100 * time.Millisecond) // Simulate work
			return nil
		},
	}

	if err := engine.RegisterExecutor(JobTypeLUA, executor); err != nil {
		t.Fatalf("Failed to register executor: %v", err)
	}

	if err := engine.Start(); err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}
	defer engine.Stop()

	// Add multiple jobs
	jobs := []*Job{
		{
			ID:             uuidv7.New(),
			Name:           "Job 1",
			Type:           JobTypeLUA,
			CronExpression: "@every 1s",
			Enabled:        true,
			Payload:        map[string]interface{}{"name": "job1"},
		},
		{
			ID:             uuidv7.New(),
			Name:           "Job 2",
			Type:           JobTypeLUA,
			CronExpression: "@every 1s",
			Enabled:        true,
			Payload:        map[string]interface{}{"name": "job2"},
		},
		{
			ID:             uuidv7.New(),
			Name:           "Job 3",
			Type:           JobTypeLUA,
			CronExpression: "@every 1s",
			Enabled:        true,
			Payload:        map[string]interface{}{"name": "job3"},
		},
	}

	for _, job := range jobs {
		if err := engine.AddJob(job); err != nil {
			t.Fatalf("Failed to add job: %v", err)
		}
	}

	// Wait for executions
	time.Sleep(1500 * time.Millisecond)

	// Verify all jobs executed
	if !job1Executed.Load() {
		t.Error("Job 1 did not execute")
	}
	if !job2Executed.Load() {
		t.Error("Job 2 did not execute")
	}
	if !job3Executed.Load() {
		t.Error("Job 3 did not execute")
	}

	// Check stats
	stats := engine.GetStats()
	if stats["total_jobs"].(int) != 3 {
		t.Errorf("Expected 3 total jobs, got %d", stats["total_jobs"])
	}
	if stats["enabled_jobs"].(int) != 3 {
		t.Errorf("Expected 3 enabled jobs, got %d", stats["enabled_jobs"])
	}
}

// TestEngine_Integration_DisableJobDuringExecution tests disabling job during run
func TestEngine_Integration_DisableJobDuringExecution(t *testing.T) {
	cfg := DefaultConfig()
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	executed := atomic.Int32{}
	executor := &MockExecutor{
		id: "disable-executor",
		executeFunc: func(ctx context.Context, job *Job) error {
			executed.Add(1)
			return nil
		},
	}

	if err := engine.RegisterExecutor(JobTypeLUA, executor); err != nil {
		t.Fatalf("Failed to register executor: %v", err)
	}

	if err := engine.Start(); err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}
	defer engine.Stop()

	// Add job
	job := &Job{
		ID:             uuidv7.New(),
		Name:           "Disable Test Job",
		Type:           JobTypeLUA,
		CronExpression: "@every 1s",
		Enabled:        true,
	}

	if err := engine.AddJob(job); err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	// Wait for first execution
	time.Sleep(1200 * time.Millisecond)
	firstCount := executed.Load()

	// Disable job
	if err := engine.DisableJob(job.ID); err != nil {
		t.Errorf("Failed to disable job: %v", err)
	}

	// Wait and verify no more executions
	time.Sleep(1500 * time.Millisecond)
	finalCount := executed.Load()

	// Should have had at least 1 execution before disable
	if firstCount < 1 {
		t.Error("Expected at least 1 execution before disable")
	}

	// Should not have significantly more executions after disable
	// (might have 1 more if execution was already queued)
	if finalCount > firstCount+1 {
		t.Errorf("Job continued executing after disable: before=%d, after=%d", firstCount, finalCount)
	}
}

// TestEngine_Integration_HealthMonitor tests health monitoring
func TestEngine_Integration_HealthMonitor(t *testing.T) {
	cfg := DefaultConfig()
	cfg.HealthCheckInterval = 200 * time.Millisecond
	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	if err := engine.Start(); err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}
	defer engine.Stop()

	// Check health multiple times
	for i := 0; i < 3; i++ {
		health := engine.GetHealth()
		if health["status"] != "healthy" {
			t.Errorf("Iteration %d: Expected healthy status, got %v", i, health["status"])
		}
		if !health["running"].(bool) {
			t.Errorf("Iteration %d: Expected running=true", i)
		}
		time.Sleep(250 * time.Millisecond)
	}

	// Stop and check health
	engine.Stop()
	health := engine.GetHealth()
	if health["status"] != "stopped" {
		t.Errorf("Expected stopped status after Stop(), got %v", health["status"])
	}
	if health["running"].(bool) {
		t.Error("Expected running=false after Stop()")
	}
}
