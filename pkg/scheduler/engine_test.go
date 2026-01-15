package scheduler

import (
	"fmt"
	"testing"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// TestNewEngine tests engine initialization
func TestNewEngine(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, err := NewEngine(cfg)
		
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}
		if engine == nil {
			t.Fatal("Expected non-nil engine")
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		cfg := Config{WorkerPoolSize: -1}
		engine, err := NewEngine(cfg)
		
		if err == nil {
			t.Error("Expected error for invalid config")
		}
		if engine != nil {
			t.Error("Expected nil engine for invalid config")
		}
	})
}

// TestEngine_StartStop tests engine lifecycle
func TestEngine_StartStop(t *testing.T) {
	t.Run("start success", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		
		err := engine.Start()
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		
		if !engine.IsRunning() {
			t.Error("Engine should be running after Start()")
		}
		
		_ = engine.Stop()
	})

	t.Run("start already running", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		
		_ = engine.Start()
		defer func() { _ = engine.Stop() }()
		
		err := engine.Start()
		if err != ErrEngineAlreadyRunning {
			t.Errorf("Expected ErrEngineAlreadyRunning, got %v", err)
		}
	})

	t.Run("stop success", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		
		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		if err := engine.Stop(); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
		
		if engine.IsRunning() {
			t.Error("Engine should not be running after Stop()")
		}
	})

	t.Run("stop not running", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		
		// Stop without Start should return ErrEngineNotRunning
		err := engine.Stop()
		if err != ErrEngineNotRunning {
			t.Errorf("Expected ErrEngineNotRunning, got %v", err)
		}
		
		if engine.IsRunning() {
			t.Error("Engine should not be running")
		}
	})
}

// TestEngine_RegisterExecutor tests executor registration
func TestEngine_RegisterExecutor(t *testing.T) {
	t.Run("register success", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		
		executor := NewMockExecutor()
		err := engine.RegisterExecutor("test", executor)
		
		if err != nil {
			t.Errorf("RegisterExecutor() error = %v", err)
		}
	})

	t.Run("register duplicate", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		
		executor := NewMockExecutor()
		if err := engine.RegisterExecutor("test", executor); err != nil {
			t.Fatalf("RegisterExecutor() error = %v", err)
		}
		
		err := engine.RegisterExecutor("test", executor)
		if err != ErrExecutorAlreadyRegistered {
			t.Errorf("Expected ErrExecutorAlreadyRegistered, got %v", err)
		}
	})
}

// TestEngine_AddJob tests job addition
func TestEngine_AddJob(t *testing.T) {
	t.Run("add valid job enabled", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer func() {
			if err := engine.Stop(); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
		}()
		
		executor := NewMockExecutor()
		if err := engine.RegisterExecutor("test", executor); err != nil {
			t.Fatalf("RegisterExecutor() error = %v", err)
		}
		
		job := &Job{
			ID:           uuidv7.New(),
			Name:         "Test Job",
			Type:         "test",
			CronExpression: "* * * * *",
			Enabled:      true,
		}
		
		err := engine.AddJob(job)
		if err != nil {
			t.Errorf("AddJob() error = %v", err)
		}
	})

	t.Run("add valid job disabled", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer func() {
			if err := engine.Stop(); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
		}()
		
		executor := NewMockExecutor()
		if err := engine.RegisterExecutor("test", executor); err != nil {
			t.Fatalf("RegisterExecutor() error = %v", err)
		}
		
		job := &Job{
			ID:             uuidv7.New(),
			Name:           "Disabled Job",
			Type:           "test",
			CronExpression: "* * * * *",
			Enabled:        false,
		}
		
		err := engine.AddJob(job)
		if err != nil {
			t.Errorf("AddJob() error = %v", err)
		}
	})

	t.Run("add duplicate job", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer func() {
			if err := engine.Stop(); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
		}()
		
		executor := NewMockExecutor()
		if err := engine.RegisterExecutor("test", executor); err != nil {
			t.Fatalf("RegisterExecutor() error = %v", err)
		}
		
		job := &Job{
			ID:           uuidv7.New(),
			Name:         "Test Job",
			Type:         "test",
			CronExpression: "* * * * *",
			Enabled:      true,
		}
		
		if err := engine.AddJob(job); err != nil {
			t.Fatalf("AddJob() error = %v", err)
		}
		err := engine.AddJob(job)
		
		if err != ErrJobAlreadyExists {
			t.Errorf("Expected ErrJobAlreadyExists, got %v", err)
		}
	})

	t.Run("add invalid cron", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer func() {
			if err := engine.Stop(); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
		}()
		
		executor := NewMockExecutor()
		if err := engine.RegisterExecutor("test", executor); err != nil {
			t.Fatalf("RegisterExecutor() error = %v", err)
		}
		
		job := &Job{
			ID:           uuidv7.New(),
			Name:         "Bad Cron",
			Type:         "test",
			CronExpression: "invalid cron",
			Enabled:      true,
		}
		
		err := engine.AddJob(job)
		if err == nil {
			t.Error("Expected error for invalid cron expression")
		}
	})

	t.Run("add when not running", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		
		job := &Job{
			ID:           uuidv7.New(),
			Name:         "Test Job",
			Type:         "test",
			CronExpression: "* * * * *",
			Enabled:      true,
		}
		
		err := engine.AddJob(job)
		if err != ErrEngineNotRunning {
			t.Errorf("Expected ErrEngineNotRunning, got %v", err)
		}
	})
}

// TestEngine_RemoveJob tests job removal
func TestEngine_RemoveJob(t *testing.T) {
	t.Run("remove existing job", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer func() {
			if err := engine.Stop(); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
		}()
		
		executor := NewMockExecutor()
		if err := engine.RegisterExecutor("test", executor); err != nil {
			t.Fatalf("RegisterExecutor() error = %v", err)
		}
		
		job := &Job{
			ID:           uuidv7.New(),
			Name:         "Test Job",
			Type:         "test",
			CronExpression: "* * * * *",
			Enabled:      true,
		}
		
		if err := engine.AddJob(job); err != nil {
			t.Fatalf("AddJob() error = %v", err)
		}
		err := engine.RemoveJob(job.ID)
		
		if err != nil {
			t.Errorf("RemoveJob() error = %v", err)
		}
		
		// Verify job is removed
		_, err = engine.GetJob(job.ID)
		if err != ErrJobNotFound {
			t.Error("Expected job to be removed")
		}
	})

	t.Run("remove non-existent job", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer func() {
			if err := engine.Stop(); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
		}()
		
		err := engine.RemoveJob(uuidv7.New())
		if err != ErrJobNotFound {
			t.Errorf("Expected ErrJobNotFound, got %v", err)
		}
	})

	t.Run("remove when not running", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		
		err := engine.RemoveJob(uuidv7.New())
		if err != ErrEngineNotRunning {
			t.Errorf("Expected ErrEngineNotRunning, got %v", err)
		}
	})
}

// TestEngine_GetJob tests job retrieval
func TestEngine_GetJob(t *testing.T) {
	t.Run("get existing job", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer func() {
			if err := engine.Stop(); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
		}()
		
		executor := NewMockExecutor()
		if err := engine.RegisterExecutor("test", executor); err != nil {
			t.Fatalf("RegisterExecutor() error = %v", err)
		}
		
		job := &Job{
			ID:           uuidv7.New(),
			Name:         "Test Job",
			Type:         "test",
			CronExpression: "* * * * *",
			Enabled:      true,
		}
		
		if err := engine.AddJob(job); err != nil {
			t.Fatalf("AddJob() error = %v", err)
		}
		
		retrieved, err := engine.GetJob(job.ID)
		if err != nil {
			t.Fatalf("GetJob() error = %v", err)
		}
		
		if retrieved.ID != job.ID {
			t.Errorf("Expected ID=%s, got %s", job.ID, retrieved.ID)
		}
	})

	t.Run("get non-existent job", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer func() {
			if err := engine.Stop(); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
		}()
		
		_, err := engine.GetJob(uuidv7.New())
		if err != ErrJobNotFound {
			t.Errorf("Expected ErrJobNotFound, got %v", err)
		}
	})
}

// TestEngine_UpdateJob tests job updates
func TestEngine_UpdateJob(t *testing.T) {
	t.Run("update existing job", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer func() {
			if err := engine.Stop(); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
		}()
		
		executor := NewMockExecutor()
		if err := engine.RegisterExecutor("test", executor); err != nil {
			t.Fatalf("RegisterExecutor() error = %v", err)
		}
		
		job := &Job{
			ID:           uuidv7.New(),
			Name:         "Original Name",
			Type:         "test",
			CronExpression: "* * * * *",
			Enabled:      true,
		}
		
		if err := engine.AddJob(job); err != nil {
			t.Fatalf("AddJob() error = %v", err)
		}
		
		// Update job
		job.Name = "Updated Name"
		err := engine.UpdateJob(job)
		
		if err != nil {
			t.Errorf("UpdateJob() error = %v", err)
		}
		
		// Verify update
		updated, _ := engine.GetJob(job.ID)
		if updated.Name != "Updated Name" {
			t.Errorf("Expected Name='Updated Name', got '%s'", updated.Name)
		}
	})

	t.Run("update non-existent job", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer func() {
			if err := engine.Stop(); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
		}()
		
		job := &Job{
		ID:           uuidv7.New(),
			Type:         "test",
			CronExpression: "* * * * *",
			Enabled:      true,
		}
		
		err := engine.UpdateJob(job)
		if err != ErrJobNotFound {
			t.Errorf("Expected ErrJobNotFound, got %v", err)
		}
	})
}

// TestEngine_EnableDisableJob tests job enable/disable operations
func TestEngine_EnableDisableJob(t *testing.T) {
	t.Run("enable job", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer func() {
			if err := engine.Stop(); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
		}()
		
		executor := NewMockExecutor()
		if err := engine.RegisterExecutor("test", executor); err != nil {
			t.Fatalf("RegisterExecutor() error = %v", err)
		}
		
		job := &Job{
			ID:           uuidv7.New(),
			Name:         "Test Job",
			Type:         "test",
			CronExpression: "* * * * *",
			Enabled:      false,
		}
		
	if err := engine.AddJob(job); err != nil {
		t.Fatalf("AddJob() error = %v", err)
	}
	if err := engine.EnableJob(job.ID); err != nil {
		t.Fatalf("EnableJob() error = %v", err)
	}
	// Verify enabled
		retrieved, _ := engine.GetJob(job.ID)
		if !retrieved.Enabled {
			t.Error("Job should be enabled")
		}
	})

	t.Run("disable job", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer func() {
			if err := engine.Stop(); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
		}()
		
		executor := NewMockExecutor()
		if err := engine.RegisterExecutor("test", executor); err != nil {
			t.Fatalf("RegisterExecutor() error = %v", err)
		}
		
		job := &Job{
			ID:           uuidv7.New(),
			Name:         "Test Job",
			Type:         "test",
			CronExpression: "* * * * *",
			Enabled:      true,
		}
		
	if err := engine.AddJob(job); err != nil {
		t.Fatalf("AddJob() error = %v", err)
	}
	if err := engine.DisableJob(job.ID); err != nil {
		t.Fatalf("DisableJob() error = %v", err)
	}
	// Verify disabled
		retrieved, _ := engine.GetJob(job.ID)
		if retrieved.Enabled {
			t.Error("Job should be disabled")
		}
	})

	t.Run("enable non-existent job", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer func() {
			if err := engine.Stop(); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
		}()
		
		err := engine.EnableJob(uuidv7.New())
		if err != ErrJobNotFound {
			t.Errorf("Expected ErrJobNotFound, got %v", err)
		}
	})
}

// TestEngine_ListJobs tests job listing
func TestEngine_ListJobs(t *testing.T) {
	t.Run("list empty", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer func() {
			if err := engine.Stop(); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
		}()
		
		jobs := engine.ListJobs()
		if len(jobs) != 0 {
			t.Errorf("Expected 0 jobs, got %d", len(jobs))
		}
	})

	t.Run("list multiple jobs", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer func() {
			if err := engine.Stop(); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
		}()
		
		executor := NewMockExecutor()
		if err := engine.RegisterExecutor("test", executor); err != nil {
			t.Fatalf("RegisterExecutor() error = %v", err)
		}
		
		// Add 3 jobs
		for i := 1; i <= 3; i++ {
			job := &Job{
			ID:           uuidv7.New(),
				Name:         fmt.Sprintf("Job %d", i),
				Type:         "test",
				CronExpression: "* * * * *",
				Enabled:      true,
			}
			if err := engine.AddJob(job); err != nil {
				t.Fatalf("AddJob() error = %v", err)
			}
		}
		
		jobs := engine.ListJobs()
		if len(jobs) != 3 {
			t.Errorf("Expected 3 jobs, got %d", len(jobs))
		}
	})
}

// TestEngine_GetStats tests statistics retrieval
func TestEngine_GetStats(t *testing.T) {
	cfg := DefaultConfig()
	engine, _ := NewEngine(cfg)
	if err := engine.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() {
		if err := engine.Stop(); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
	}()
	
	stats := engine.GetStats()
	
	// Verify all stat fields
	if stats["running"] != true {
		t.Error("Expected running=true")
	}
	if stats["total_jobs"] != 0 {
		t.Error("Expected total_jobs=0")
	}
	if stats["enabled_jobs"] != 0 {
		t.Error("Expected enabled_jobs=0")
	}
	if _, exists := stats["worker_pool"]; !exists {
		t.Error("Expected worker_pool in stats")
	}
	if _, exists := stats["max_concurrent"]; !exists {
		t.Error("Expected max_concurrent in stats")
	}
}

// TestEngine_GetHealth tests health check
func TestEngine_GetHealth(t *testing.T) {
	t.Run("healthy when running", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		if err := engine.Start(); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		defer func() {
			if err := engine.Stop(); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
		}()
		
		health := engine.GetHealth()
		
		if health["status"] != "healthy" {
			t.Errorf("Expected status='healthy', got '%v'", health["status"])
		}
	})

	t.Run("stopped when not running", func(t *testing.T) {
		cfg := DefaultConfig()
		engine, _ := NewEngine(cfg)
		
		health := engine.GetHealth()
		
		if health["status"] != "stopped" {
			t.Errorf("Expected status='stopped', got '%v'", health["status"])
		}
	})
}

// TestEngine_IsRunning tests running state check
func TestEngine_IsRunning(t *testing.T) {
	cfg := DefaultConfig()
	engine, _ := NewEngine(cfg)
	
	// Before Start
	if engine.IsRunning() {
		t.Error("Engine should not be running before Start()")
	}
	
	// After Start
	if err := engine.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !engine.IsRunning() {
		t.Error("Engine should be running after Start()")
	}
	
	// After Stop
	if err := engine.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if engine.IsRunning() {
		t.Error("Engine should not be running after Stop()")
	}
}
