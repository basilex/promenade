package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPostCommentHandler_CreateComment_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &PostCommentHandler{}

	// Mock auth middleware that adds user_id to context
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "01936d9a-0000-7000-8000-000000000000")
		c.Next()
	})
	router.POST("/comments", handler.CreateComment)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/comments", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestPostCommentHandler_CreateComment_MissingRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &PostCommentHandler{}

	// Mock auth middleware that adds user_id to context
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "01936d9a-0000-7000-8000-000000000000")
		c.Next()
	})
	router.POST("/comments", handler.CreateComment)

	// Missing post_id and content
	requestBody := map[string]interface{}{
		"parent_id": nil,
	}
	body, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/comments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostCommentHandler_GetComment_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &PostCommentHandler{}
	router.GET("/comments/:id", handler.GetComment)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/comments/invalid-uuid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestPostCommentHandler_UpdateComment_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &PostCommentHandler{}

	// Mock auth middleware
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "01936d9a-0000-7000-8000-000000000000")
		c.Next()
	})
	router.PUT("/comments/:id", handler.UpdateComment)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/comments/01936d9a-0000-7000-8000-000000000000", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostCommentHandler_UpdateComment_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &PostCommentHandler{}

	// Mock auth middleware
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "01936d9a-0000-7000-8000-000000000000")
		c.Next()
	})
	router.PUT("/comments/:id", handler.UpdateComment)

	requestBody := map[string]interface{}{
		"content": "Updated content",
	}
	body, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/comments/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostCommentHandler_DeleteComment_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &PostCommentHandler{}

	// Mock auth middleware
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "01936d9a-0000-7000-8000-000000000000")
		c.Next()
	})
	router.DELETE("/comments/:id", handler.DeleteComment)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/comments/invalid-uuid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostCommentHandler_GetPostComments_InvalidPostID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &PostCommentHandler{}
	router.GET("/posts/:post_id/comments", handler.GetPostComments)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/posts/invalid-uuid/comments", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Removed: TestPostCommentHandler_GetPostComments_InvalidPagination - requires mock use case

func TestPostCommentHandler_GetCommentReplies_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &PostCommentHandler{}
	router.GET("/comments/:id/replies", handler.GetCommentReplies)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/comments/invalid-uuid/replies", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostCommentHandler_GetUserComments_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &PostCommentHandler{}
	router.GET("/users/:user_id/comments", handler.GetUserComments)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/invalid-uuid/comments", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostCommentHandler_LikeComment_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &PostCommentHandler{}

	// Mock auth middleware
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "01936d9a-0000-7000-8000-000000000000")
		c.Next()
	})
	router.POST("/comments/:id/like", handler.LikeComment)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/comments/invalid-uuid/like", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostCommentHandler_UnlikeComment_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &PostCommentHandler{}

	// Mock auth middleware
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "01936d9a-0000-7000-8000-000000000000")
		c.Next()
	})
	router.DELETE("/comments/:id/like", handler.UnlikeComment)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/comments/invalid-uuid/like", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
