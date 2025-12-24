package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success response with data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		data := map[string]string{"name": "John Doe"}
		Success(c, http.StatusOK, data)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.NotNil(t, resp.Data)
		assert.Greater(t, resp.Timestamp, int64(0))
	})

	t.Run("success response with 201 Created", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		data := map[string]string{"id": "123"}
		Success(c, http.StatusCreated, data)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("success response with nil data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		Success(c, http.StatusOK, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Nil(t, resp.Data)
	})
}

func TestError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("error response with error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		err := errors.New("something went wrong")
		Error(c, http.StatusBadRequest, "invalid request", err)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp Response
		jsonErr := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, jsonErr)
		assert.False(t, resp.Success)
		assert.Equal(t, "invalid request", resp.Message)
		assert.Equal(t, "something went wrong", resp.Error)
		assert.Greater(t, resp.Timestamp, int64(0))
	})

	t.Run("error response without error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		Error(c, http.StatusUnauthorized, "unauthorized", nil)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var resp Response
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Equal(t, "unauthorized", resp.Message)
		assert.Empty(t, resp.Error)
	})

	t.Run("error response with 500 status", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		err := errors.New("database connection failed")
		Error(c, http.StatusInternalServerError, "internal server error", err)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var resp Response
		jsonErr := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, jsonErr)
		assert.False(t, resp.Success)
		assert.Equal(t, "internal server error", resp.Message)
		assert.Equal(t, "database connection failed", resp.Error)
	})
}

func TestGetPageFromQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid page number", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			URL: &url.URL{RawQuery: "page=5"},
		}

		page := GetPageFromQuery(c)
		assert.Equal(t, 5, page)
	})

	t.Run("default page when not provided", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			URL: &url.URL{},
		}

		page := GetPageFromQuery(c)
		assert.Equal(t, 1, page)
	})

	t.Run("page less than 1 defaults to 1", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			URL: &url.URL{RawQuery: "page=0"},
		}

		page := GetPageFromQuery(c)
		assert.Equal(t, 1, page)
	})

	t.Run("negative page defaults to 1", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			URL: &url.URL{RawQuery: "page=-5"},
		}

		page := GetPageFromQuery(c)
		assert.Equal(t, 1, page)
	})

	t.Run("invalid page string defaults to 1", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			URL: &url.URL{RawQuery: "page=abc"},
		}

		page := GetPageFromQuery(c)
		assert.Equal(t, 1, page)
	})
}

func TestGetPageSizeFromQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid page size", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			URL: &url.URL{RawQuery: "page_size=50"},
		}

		pageSize := GetPageSizeFromQuery(c)
		assert.Equal(t, 50, pageSize)
	})

	t.Run("default page size when not provided", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			URL: &url.URL{},
		}

		pageSize := GetPageSizeFromQuery(c)
		assert.Equal(t, 20, pageSize)
	})

	t.Run("page size less than 1 defaults to 20", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			URL: &url.URL{RawQuery: "page_size=0"},
		}

		pageSize := GetPageSizeFromQuery(c)
		assert.Equal(t, 20, pageSize)
	})

	t.Run("page size greater than 100 capped at 100", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			URL: &url.URL{RawQuery: "page_size=150"},
		}

		pageSize := GetPageSizeFromQuery(c)
		assert.Equal(t, 100, pageSize)
	})

	t.Run("invalid page size defaults to 20", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			URL: &url.URL{RawQuery: "page_size=xyz"},
		}

		pageSize := GetPageSizeFromQuery(c)
		assert.Equal(t, 20, pageSize)
	})
}

func TestNewPaginatedResponse(t *testing.T) {
	t.Run("exact page division", func(t *testing.T) {
		items := []string{"item1", "item2", "item3"}
		page := 1
		pageSize := 20
		total := 60

		resp := NewPaginatedResponse(items, page, pageSize, total)

		assert.Equal(t, items, resp.Items)
		assert.Equal(t, 1, resp.Page)
		assert.Equal(t, 20, resp.PageSize)
		assert.Equal(t, 60, resp.Total)
		assert.Equal(t, 3, resp.TotalPages) // 60/20 = 3
		assert.Greater(t, resp.Timestamp, int64(0))
	})

	t.Run("with remainder", func(t *testing.T) {
		items := []string{"item1", "item2"}
		page := 2
		pageSize := 10
		total := 25

		resp := NewPaginatedResponse(items, page, pageSize, total)

		assert.Equal(t, 2, resp.Page)
		assert.Equal(t, 10, resp.PageSize)
		assert.Equal(t, 25, resp.Total)
		assert.Equal(t, 3, resp.TotalPages) // 25/10 = 2 remainder 5, so 3 pages
	})

	t.Run("single page", func(t *testing.T) {
		items := []string{"item1", "item2", "item3"}
		page := 1
		pageSize := 20
		total := 3

		resp := NewPaginatedResponse(items, page, pageSize, total)

		assert.Equal(t, 1, resp.TotalPages)
		assert.Equal(t, 3, resp.Total)
	})

	t.Run("empty items", func(t *testing.T) {
		items := []string{}
		page := 1
		pageSize := 20
		total := 0

		resp := NewPaginatedResponse(items, page, pageSize, total)

		assert.Equal(t, 0, resp.TotalPages)
		assert.Equal(t, 0, resp.Total)
		assert.NotNil(t, resp.Items)
	})

	t.Run("large dataset", func(t *testing.T) {
		items := make([]int, 50)
		page := 5
		pageSize := 50
		total := 1000

		resp := NewPaginatedResponse(items, page, pageSize, total)

		assert.Equal(t, 5, resp.Page)
		assert.Equal(t, 50, resp.PageSize)
		assert.Equal(t, 1000, resp.Total)
		assert.Equal(t, 20, resp.TotalPages) // 1000/50 = 20
	})
}
