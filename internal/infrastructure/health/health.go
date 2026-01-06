package health

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"github.com/basilex/promenade/pkg/bus"
)

// Status represents health check status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// Check represents a single health check result
type Check struct {
	Name      string        `json:"name"`
	Status    Status        `json:"status"`
	Message   string        `json:"message,omitempty"`
	Duration  time.Duration `json:"duration_ms" swaggertype:"integer"`
	Timestamp time.Time     `json:"timestamp"`
}

// Report represents overall health check report
type Report struct {
	Status    Status           `json:"status"`
	Checks    map[string]Check `json:"checks"`
	Timestamp time.Time        `json:"timestamp"`
	Version   string           `json:"version,omitempty"`
}

// Checker performs health checks on dependencies
type Checker struct {
	db       *sqlx.DB
	redis    *redis.Client
	eventBus bus.IBus
	version  string
}

// NewChecker creates a new health checker
func NewChecker(db *sqlx.DB, redis *redis.Client, eventBus bus.IBus, version string) *Checker {
	return &Checker{
		db:       db,
		redis:    redis,
		eventBus: eventBus,
		version:  version,
	}
}

// CheckAll performs all health checks with timeout
func (c *Checker) CheckAll(ctx context.Context) *Report {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	checks := make(map[string]Check)

	// PostgreSQL check
	checks["database"] = c.checkDatabase(ctx)

	// Redis check (optional)
	if c.redis != nil {
		checks["redis"] = c.checkRedis(ctx)
	}

	// Event Bus check
	checks["event_bus"] = c.checkEventBus(ctx)

	// Determine overall status
	overallStatus := StatusHealthy
	for _, check := range checks {
		if check.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
			break
		}
		if check.Status == StatusDegraded && overallStatus == StatusHealthy {
			overallStatus = StatusDegraded
		}
	}

	return &Report{
		Status:    overallStatus,
		Checks:    checks,
		Timestamp: time.Now(),
		Version:   c.version,
	}
}

// CheckDatabase checks PostgreSQL database connectivity
func (c *Checker) CheckDatabase(ctx context.Context) Check {
	return c.checkDatabase(ctx)
}

func (c *Checker) checkDatabase(ctx context.Context) Check {
	start := time.Now()
	check := Check{
		Name:      "PostgreSQL",
		Timestamp: start,
	}

	// Ping database
	err := c.db.PingContext(ctx)
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("database ping failed: %v", err)
		check.Duration = time.Since(start)
		return check
	}

	// Check if database is writable (optional quick test)
	var result int
	err = c.db.GetContext(ctx, &result, "SELECT 1")
	if err != nil {
		check.Status = StatusDegraded
		check.Message = "database ping ok but query failed"
		check.Duration = time.Since(start)
		return check
	}

	check.Status = StatusHealthy
	check.Message = "database connection ok"
	check.Duration = time.Since(start)
	return check
}

// CheckRedis checks Redis connectivity
func (c *Checker) CheckRedis(ctx context.Context) Check {
	if c.redis == nil {
		return Check{
			Name:      "Redis",
			Status:    StatusHealthy,
			Message:   "redis not configured (optional)",
			Timestamp: time.Now(),
		}
	}

	return c.checkRedis(ctx)
}

func (c *Checker) checkRedis(ctx context.Context) Check {
	start := time.Now()
	check := Check{
		Name:      "Redis",
		Timestamp: start,
	}

	// Ping Redis
	err := c.redis.Ping(ctx).Err()
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("redis ping failed: %v", err)
		check.Duration = time.Since(start)
		return check
	}

	check.Status = StatusHealthy
	check.Message = "redis connection ok"
	check.Duration = time.Since(start)
	return check
}

// CheckEventBus checks Event Bus health
func (c *Checker) CheckEventBus(ctx context.Context) Check {
	return c.checkEventBus(ctx)
}

func (c *Checker) checkEventBus(ctx context.Context) Check {
	start := time.Now()
	check := Check{
		Name:      "Event Bus",
		Timestamp: start,
	}

	// Check event bus health
	err := c.eventBus.Health(ctx)
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("event bus unhealthy: %v", err)
		check.Duration = time.Since(start)
		return check
	}

	check.Status = StatusHealthy
	check.Message = "event bus operational"
	check.Duration = time.Since(start)
	return check
}
