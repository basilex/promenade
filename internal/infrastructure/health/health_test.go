package health

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/bus/memory"
)

func TestChecker_CheckDatabase_Healthy(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectPing()
	mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	checker := NewChecker(sqlxDB, nil, nil, "test")
	check := checker.CheckDatabase(context.Background())

	assert.Equal(t, StatusHealthy, check.Status)
	assert.Contains(t, check.Message, "database connection ok")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChecker_CheckDatabase_Unhealthy_PingFailed(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectPing().WillReturnError(errors.New("connection refused"))

	checker := NewChecker(sqlxDB, nil, nil, "test")
	check := checker.CheckDatabase(context.Background())

	assert.Equal(t, StatusUnhealthy, check.Status)
	assert.Contains(t, check.Message, "database ping failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChecker_CheckDatabase_Degraded_QueryFailed(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectPing()
	mock.ExpectQuery("SELECT 1").WillReturnError(errors.New("query failed"))

	checker := NewChecker(sqlxDB, nil, nil, "test")
	check := checker.CheckDatabase(context.Background())

	assert.Equal(t, StatusDegraded, check.Status)
	assert.Contains(t, check.Message, "database ping ok but query failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChecker_CheckRedis_Healthy(t *testing.T) {
	redisClient, mock := redismock.NewClientMock()

	mock.ExpectPing().SetVal("PONG")

	checker := NewChecker(nil, redisClient, nil, "test")
	check := checker.CheckRedis(context.Background())

	assert.Equal(t, StatusHealthy, check.Status)
	assert.Contains(t, check.Message, "redis connection ok")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChecker_CheckRedis_Unhealthy(t *testing.T) {
	redisClient, mock := redismock.NewClientMock()

	mock.ExpectPing().SetErr(errors.New("connection refused"))

	checker := NewChecker(nil, redisClient, nil, "test")
	check := checker.CheckRedis(context.Background())

	assert.Equal(t, StatusUnhealthy, check.Status)
	assert.Contains(t, check.Message, "redis ping failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChecker_CheckRedis_NotConfigured(t *testing.T) {
	checker := NewChecker(nil, nil, nil, "test")
	check := checker.CheckRedis(context.Background())

	assert.Equal(t, StatusHealthy, check.Status)
	assert.Contains(t, check.Message, "redis not configured")
}

func TestChecker_CheckEventBus_Healthy(t *testing.T) {
	eventBus := memory.NewMemoryBus(bus.Config{
		WorkerPoolSize: 1,
		BufferSize:     10,
	})

	checker := NewChecker(nil, nil, eventBus, "test")
	check := checker.CheckEventBus(context.Background())

	assert.Equal(t, StatusHealthy, check.Status)
	assert.Contains(t, check.Message, "event bus operational")
}

func TestChecker_CheckAll_AllHealthy(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	redisClient, redisMock := redismock.NewClientMock()
	eventBus := memory.NewMemoryBus(bus.Config{
		WorkerPoolSize: 1,
		BufferSize:     10,
	})

	mock.ExpectPing()
	mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	redisMock.ExpectPing().SetVal("PONG")

	checker := NewChecker(sqlxDB, redisClient, eventBus, "1.0.0")
	report := checker.CheckAll(context.Background())

	assert.Equal(t, StatusHealthy, report.Status)
	assert.Equal(t, "1.0.0", report.Version)
	assert.Len(t, report.Checks, 3) // db, redis, bus

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestChecker_CheckAll_DatabaseUnhealthy(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	redisClient, redisMock := redismock.NewClientMock()
	eventBus := memory.NewMemoryBus(bus.Config{
		WorkerPoolSize: 1,
		BufferSize:     10,
	})

	mock.ExpectPing().WillReturnError(errors.New("db down"))
	redisMock.ExpectPing().SetVal("PONG")

	checker := NewChecker(sqlxDB, redisClient, eventBus, "1.0.0")
	report := checker.CheckAll(context.Background())

	assert.Equal(t, StatusUnhealthy, report.Status)
	assert.Equal(t, StatusUnhealthy, report.Checks["database"].Status)

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.NoError(t, redisMock.ExpectationsWereMet())
}

func TestChecker_CheckAll_WithTimeout(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	eventBus := memory.NewMemoryBus(bus.Config{
		WorkerPoolSize: 1,
		BufferSize:     10,
	})

	// Simulate slow ping
	mock.ExpectPing().WillDelayFor(10 * time.Second)

	checker := NewChecker(sqlxDB, nil, eventBus, "test")

	start := time.Now()
	report := checker.CheckAll(context.Background())
	duration := time.Since(start)

	// Should timeout within 5 seconds (not wait 10 seconds)
	assert.Less(t, duration, 6*time.Second)
	assert.Equal(t, StatusUnhealthy, report.Status)
}
