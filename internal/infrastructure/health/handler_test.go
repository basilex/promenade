package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockChecker implements health checking for tests
type MockChecker struct {
	ReportFunc func(ctx context.Context) *Report
	DBFunc     func(ctx context.Context) Check
	RedisFunc  func(ctx context.Context) Check
	BusFunc    func(ctx context.Context) Check
}

func (m *MockChecker) CheckAll(ctx context.Context) *Report {
	if m.ReportFunc != nil {
		return m.ReportFunc(ctx)
	}
	return &Report{Status: StatusHealthy, Version: "test"}
}

func (m *MockChecker) CheckDatabase(ctx context.Context) Check {
	if m.DBFunc != nil {
		return m.DBFunc(ctx)
	}
	return Check{Name: "PostgreSQL", Status: StatusHealthy}
}

func (m *MockChecker) CheckRedis(ctx context.Context) Check {
	if m.RedisFunc != nil {
		return m.RedisFunc(ctx)
	}
	return Check{Name: "Redis", Status: StatusHealthy}
}

func (m *MockChecker) CheckEventBus(ctx context.Context) Check {
	if m.BusFunc != nil {
		return m.BusFunc(ctx)
	}
	return Check{Name: "Event Bus", Status: StatusHealthy}
}

func TestHandler_CheckAll_Healthy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	checker := &MockChecker{
		ReportFunc: func(ctx context.Context) *Report {
			return &Report{
				Status:  StatusHealthy,
				Version: "1.0.0",
				Checks: map[string]Check{
					"database": {Name: "PostgreSQL", Status: StatusHealthy, Message: "ok"},
				},
			}
		},
	}

	handler := NewHandler(checker)
	router := gin.New()
	handler.RegisterRoutes(router)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var report Report
	err := json.Unmarshal(w.Body.Bytes(), &report)
	require.NoError(t, err)

	assert.Equal(t, StatusHealthy, report.Status)
	assert.Equal(t, "1.0.0", report.Version)
}

func TestHandler_CheckAll_Unhealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	checker := &MockChecker{
		ReportFunc: func(ctx context.Context) *Report {
			return &Report{
				Status:  StatusUnhealthy,
				Version: "1.0.0",
				Checks: map[string]Check{
					"database": {Name: "PostgreSQL", Status: StatusUnhealthy, Message: "connection failed"},
				},
			}
		},
	}

	handler := NewHandler(checker)
	router := gin.New()
	handler.RegisterRoutes(router)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var report Report
	err := json.Unmarshal(w.Body.Bytes(), &report)
	require.NoError(t, err)

	assert.Equal(t, StatusUnhealthy, report.Status)
}

func TestHandler_CheckAll_Degraded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	checker := &MockChecker{
		ReportFunc: func(ctx context.Context) *Report {
			return &Report{
				Status:  StatusDegraded,
				Version: "1.0.0",
				Checks: map[string]Check{
					"database": {Name: "PostgreSQL", Status: StatusHealthy},
					"redis":    {Name: "Redis", Status: StatusDegraded, Message: "slow response"},
				},
			}
		},
	}

	handler := NewHandler(checker)
	router := gin.New()
	handler.RegisterRoutes(router)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code) // Degraded still returns 200

	var report Report
	err := json.Unmarshal(w.Body.Bytes(), &report)
	require.NoError(t, err)

	assert.Equal(t, StatusDegraded, report.Status)
}

func TestHandler_CheckDatabase_Healthy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	checker := &MockChecker{
		DBFunc: func(ctx context.Context) Check {
			return Check{
				Name:    "PostgreSQL",
				Status:  StatusHealthy,
				Message: "database connection ok",
			}
		},
	}

	handler := NewHandler(checker)
	router := gin.New()
	handler.RegisterRoutes(router)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health/db", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var check Check
	err := json.Unmarshal(w.Body.Bytes(), &check)
	require.NoError(t, err)

	assert.Equal(t, StatusHealthy, check.Status)
	assert.Equal(t, "PostgreSQL", check.Name)
}

func TestHandler_CheckDatabase_Unhealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	checker := &MockChecker{
		DBFunc: func(ctx context.Context) Check {
			return Check{
				Name:    "PostgreSQL",
				Status:  StatusUnhealthy,
				Message: "connection failed",
			}
		},
	}

	handler := NewHandler(checker)
	router := gin.New()
	handler.RegisterRoutes(router)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health/db", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestHandler_CheckRedis(t *testing.T) {
	gin.SetMode(gin.TestMode)

	checker := &MockChecker{
		RedisFunc: func(ctx context.Context) Check {
			return Check{
				Name:    "Redis",
				Status:  StatusHealthy,
				Message: "redis connection ok",
			}
		},
	}

	handler := NewHandler(checker)
	router := gin.New()
	handler.RegisterRoutes(router)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health/redis", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var check Check
	err := json.Unmarshal(w.Body.Bytes(), &check)
	require.NoError(t, err)

	assert.Equal(t, StatusHealthy, check.Status)
	assert.Equal(t, "Redis", check.Name)
}

func TestHandler_CheckEventBus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	checker := &MockChecker{
		BusFunc: func(ctx context.Context) Check {
			return Check{
				Name:    "Event Bus",
				Status:  StatusHealthy,
				Message: "event bus operational",
			}
		},
	}

	handler := NewHandler(checker)
	router := gin.New()
	handler.RegisterRoutes(router)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health/bus", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var check Check
	err := json.Unmarshal(w.Body.Bytes(), &check)
	require.NoError(t, err)

	assert.Equal(t, StatusHealthy, check.Status)
	assert.Equal(t, "Event Bus", check.Name)
}

func TestHandler_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	checker := &MockChecker{}
	handler := NewHandler(checker)
	router := gin.New()
	handler.RegisterRoutes(router)

	routes := router.Routes()

	expectedRoutes := []string{"/health", "/health/db", "/health/redis", "/health/bus"}
	actualRoutes := make([]string, 0, len(routes))
	for _, route := range routes {
		actualRoutes = append(actualRoutes, route.Path)
	}

	assert.ElementsMatch(t, expectedRoutes, actualRoutes)
}
