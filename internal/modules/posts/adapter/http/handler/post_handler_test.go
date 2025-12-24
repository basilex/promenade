package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/posts/adapter/http/dto"
	"github.com/basilex/promenade/internal/modules/posts/domain/entity"
	"github.com/basilex/promenade/internal/modules/posts/usecase"
	usecaseMocks "github.com/basilex/promenade/internal/modules/posts/usecase/mocks"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// TestUserPostHandler_CreatePost tests the CreatePost handler
//
//	IModule-independent: uses only module types and mocks
func TestUserPostHandler_CreatePost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful post creation", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockUserPostUseCase)
		handler := NewUserPostHandler(mockUC)

		userID := uuidv7.New()
		createReq := dto.CreatePostRequest{
			Title:   "Test Post",
			Content: "This is test content for the post",
			Excerpt: "Test excerpt",
			Tags:    []string{"golang", "testing"},
		}

		expectedPost := &entity.UserPost{
			ID:      uuidv7.New(),
			UserID:  userID,
			Title:   createReq.Title,
			Content: createReq.Content,
			Status:  entity.PostStatusDraft,
		}

		mockUC.On("CreatePost",
			mock.Anything,
			userID,
			createReq.Title,
			createReq.Content,
			createReq.Excerpt,
			createReq.Tags,
			mock.AnythingOfType("[]string"), // categories
		).Return(expectedPost, nil)

		// Setup request
		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Setup Gin context with user_id
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)

		// Act
		handler.CreatePost(c)

		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, true, response["success"])
		data, ok := response["data"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, createReq.Title, data["title"])

		mockUC.AssertExpectations(t)
	})

	t.Run("unauthorized - no user_id", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockUserPostUseCase)
		handler := NewUserPostHandler(mockUC)

		createReq := dto.CreatePostRequest{
			Title:   "Test Post",
			Content: "This is test content",
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		// Don't set user_id

		// Act
		handler.CreatePost(c)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, false, response["success"])

		mockUC.AssertNotCalled(t, "CreatePost")
	})

	t.Run("invalid request body", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockUserPostUseCase)
		handler := NewUserPostHandler(mockUC)

		userID := uuidv7.New()

		// Invalid JSON
		body := []byte(`{"title": "ab"}`) // Title too short
		req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)

		// Act
		handler.CreatePost(c)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockUC.AssertNotCalled(t, "CreatePost")
	})

	t.Run("slug already exists", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockUserPostUseCase)
		handler := NewUserPostHandler(mockUC)

		userID := uuidv7.New()
		createReq := dto.CreatePostRequest{
			Title:   "Test Post",
			Content: "This is test content",
		}

		mockUC.On("CreatePost",
			mock.Anything,
			userID,
			createReq.Title,
			createReq.Content,
			createReq.Excerpt,
			mock.Anything,
			mock.Anything,
		).Return(nil, usecase.ErrSlugAlreadyExists)

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)

		// Act
		handler.CreatePost(c)

		// Assert
		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, false, response["success"])

		mockUC.AssertExpectations(t)
	})
}

// TestUserPostHandler_GetPost tests the GetPost handler
func TestUserPostHandler_GetPost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful retrieval", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockUserPostUseCase)
		handler := NewUserPostHandler(mockUC)

		postID := uuidv7.New()
		expectedPost := &entity.UserPost{
			ID:      postID,
			UserID:  uuidv7.New(),
			Title:   "Test Post",
			Content: "Test content",
			Status:  entity.PostStatusPublished,
		}

		mockUC.On("GetPost", mock.Anything, postID).Return(expectedPost, nil)

		req := httptest.NewRequest(http.MethodGet, "/posts/"+postID.String(), nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: postID.String()}}

		// Act
		handler.GetPost(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, true, response["success"])
		data, ok := response["data"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, expectedPost.Title, data["title"])

		mockUC.AssertExpectations(t)
	})

	t.Run("invalid post ID", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockUserPostUseCase)
		handler := NewUserPostHandler(mockUC)

		req := httptest.NewRequest(http.MethodGet, "/posts/invalid-uuid", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

		// Act
		handler.GetPost(c)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockUC.AssertNotCalled(t, "GetPost")
	})

	t.Run("post not found", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockUserPostUseCase)
		handler := NewUserPostHandler(mockUC)

		postID := uuidv7.New()
		mockUC.On("GetPost", mock.Anything, postID).Return(nil, entity.ErrNotFound)

		req := httptest.NewRequest(http.MethodGet, "/posts/"+postID.String(), nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: postID.String()}}

		// Act
		handler.GetPost(c)

		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, false, response["success"])

		mockUC.AssertExpectations(t)
	})
}

// TestUserPostHandler_DeletePost tests the DeletePost handler (uses SoftDelete)
func TestUserPostHandler_DeletePost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful deletion", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockUserPostUseCase)
		handler := NewUserPostHandler(mockUC)

		userID := uuidv7.New()
		postID := uuidv7.New()

		mockUC.On("SoftDeletePost", mock.Anything, userID, postID).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/posts/"+postID.String(), nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)
		c.Params = gin.Params{{Key: "id", Value: postID.String()}}

		// Act
		handler.DeletePost(c)

		// Assert
		// Gin's c.Status() writes with 200, not the status code passed
		// This is a known Gin behavior - use assert for actual code
		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockUserPostUseCase)
		handler := NewUserPostHandler(mockUC)

		userID := uuidv7.New()
		postID := uuidv7.New()

		mockUC.On("SoftDeletePost", mock.Anything, userID, postID).Return(usecase.ErrUnauthorized)

		req := httptest.NewRequest(http.MethodDelete, "/posts/"+postID.String(), nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)
		c.Params = gin.Params{{Key: "id", Value: postID.String()}}

		// Act
		handler.DeletePost(c)

		// Assert
		assert.Equal(t, http.StatusForbidden, w.Code)
		mockUC.AssertExpectations(t)
	})
}

// TestUserPostHandler_PublishPost tests the PublishPost handler
func TestUserPostHandler_PublishPost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful publish", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockUserPostUseCase)
		handler := NewUserPostHandler(mockUC)

		userID := uuidv7.New()
		postID := uuidv7.New()

		publishedPost := &entity.UserPost{
			ID:      postID,
			UserID:  userID,
			Title:   "Published Post",
			Content: "Content",
			Status:  entity.PostStatusPublished,
		}

		mockUC.On("PublishPost", mock.Anything, userID, postID).Return(nil)
		// Handler calls GetPost after publishing to return updated post
		mockUC.On("GetPost", mock.Anything, postID).Return(publishedPost, nil)

		req := httptest.NewRequest(http.MethodPost, "/posts/"+postID.String()+"/publish", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)
		c.Params = gin.Params{{Key: "id", Value: postID.String()}}

		// Act
		handler.PublishPost(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, true, response["success"])

		mockUC.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockUserPostUseCase)
		handler := NewUserPostHandler(mockUC)

		userID := uuidv7.New()
		postID := uuidv7.New()

		mockUC.On("PublishPost", mock.Anything, userID, postID).Return(errors.New("database error"))

		req := httptest.NewRequest(http.MethodPost, "/posts/"+postID.String()+"/publish", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", userID)
		c.Params = gin.Params{{Key: "id", Value: postID.String()}}

		// Act
		handler.PublishPost(c)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestUserPostHandler_UpdatePost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	userID := uuidv7.New()
	postID := uuidv7.New()

	title := "Updated"
	content := "Updated content here"
	updateReq := dto.UpdatePostRequest{
		Title:   &title,
		Content: &content,
	}

	expectedPost := &entity.UserPost{
		ID:      postID,
		UserID:  userID,
		Title:   "Updated",
		Content: content,
		Status:  entity.PostStatusDraft,
	}

	mockUC.On("UpdatePost", mock.Anything, userID, postID, mock.Anything).Return(expectedPost, nil)

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/posts/"+postID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Params = gin.Params{{Key: "id", Value: postID.String()}}

	handler.UpdatePost(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUserPostHandler_GetPost_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	postID := uuidv7.New()

	mockUC.On("GetPost", mock.Anything, postID).Return(nil, entity.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/posts/"+postID.String(), nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: postID.String()}}

	handler.GetPost(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUserPostHandler_DeletePost_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	userID := uuidv7.New()
	postID := uuidv7.New()

	mockUC.On("SoftDeletePost", mock.Anything, userID, postID).Return(entity.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/posts/"+postID.String(), nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Params = gin.Params{{Key: "id", Value: postID.String()}}

	handler.DeletePost(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUserPostHandler_PublishPost_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	userID := uuidv7.New()
	postID := uuidv7.New()

	mockUC.On("PublishPost", mock.Anything, userID, postID).Return(usecase.ErrUnauthorized)

	req := httptest.NewRequest(http.MethodPost, "/posts/"+postID.String()+"/publish", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Params = gin.Params{{Key: "id", Value: postID.String()}}

	handler.PublishPost(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUserPostHandler_CreatePost_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	userID := uuidv7.New()

	// Invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)

	handler.CreatePost(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserPostHandler_GetPost_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/posts/invalid-uuid", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

	handler.GetPost(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserPostHandler_GetUserPosts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	userID := uuidv7.New()

	posts := []*entity.UserPost{
		{ID: uuidv7.New(), UserID: userID, Title: "My Post 1"},
	}

	mockUC.On("GetUserPosts", mock.Anything, userID, mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(posts, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID.String()+"/posts", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "user_id", Value: userID.String()}}

	handler.GetUserPosts(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUserPostHandler_UnpublishPost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	userID := uuidv7.New()
	postID := uuidv7.New()

	post := &entity.UserPost{
		ID:     postID,
		UserID: userID,
		Title:  "Test",
		Status: entity.PostStatusDraft,
	}

	mockUC.On("UnpublishPost", mock.Anything, userID, postID).Return(nil)
	mockUC.On("GetPost", mock.Anything, postID).Return(post, nil)

	req := httptest.NewRequest(http.MethodPost, "/posts/"+postID.String()+"/unpublish", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Params = gin.Params{{Key: "id", Value: postID.String()}}

	handler.UnpublishPost(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUserPostHandler_GetPublishedPosts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	posts := []*entity.UserPost{
		{ID: uuidv7.New(), Title: "Post 1", Status: entity.PostStatusPublished},
	}

	mockUC.On("GetPublishedPosts", mock.Anything, mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(posts, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts/published", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.GetPublishedPosts(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUserPostHandler_GetFeaturedPosts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	posts := []*entity.UserPost{
		{ID: uuidv7.New(), Title: "Featured", IsFeatured: true, Status: entity.PostStatusPublished},
	}

	mockUC.On("GetFeaturedPosts", mock.Anything, mock.AnythingOfType("int")).Return(posts, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts/featured", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.GetFeaturedPosts(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUserPostHandler_SearchPosts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	posts := []*entity.UserPost{
		{ID: uuidv7.New(), Title: "Search Result", Status: entity.PostStatusPublished},
	}

	mockUC.On("SearchPosts", mock.Anything, "test", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(posts, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts/search?q=test", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.SearchPosts(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUserPostHandler_GetPostsByTag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	posts := []*entity.UserPost{
		{ID: uuidv7.New(), Title: "Tagged Post", Tags: []string{"golang"}, Status: entity.PostStatusPublished},
	}

	mockUC.On("GetPostsByTag", mock.Anything, "golang", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(posts, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts/tags/golang", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "tag", Value: "golang"}}

	handler.GetPostsByTag(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUserPostHandler_ToggleFeatured(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	userID := uuidv7.New()
	postID := uuidv7.New()

	post := &entity.UserPost{
		ID:         postID,
		UserID:     userID,
		Title:      "Test",
		IsFeatured: true,
	}

	mockUC.On("ToggleFeatured", mock.Anything, userID, postID).Return(nil)
	mockUC.On("GetPost", mock.Anything, postID).Return(post, nil)

	req := httptest.NewRequest(http.MethodPost, "/posts/"+postID.String()+"/toggle-featured", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Params = gin.Params{{Key: "id", Value: postID.String()}}

	handler.ToggleFeatured(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUserPostHandler_ToggleComments(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	userID := uuidv7.New()
	postID := uuidv7.New()

	post := &entity.UserPost{
		ID:                postID,
		UserID:            userID,
		Title:             "Test",
		IsCommentsEnabled: false,
	}

	mockUC.On("ToggleComments", mock.Anything, userID, postID).Return(nil)
	mockUC.On("GetPost", mock.Anything, postID).Return(post, nil)

	req := httptest.NewRequest(http.MethodPost, "/posts/"+postID.String()+"/toggle-comments", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Params = gin.Params{{Key: "id", Value: postID.String()}}

	handler.ToggleComments(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

// Additional error test cases

func TestUserPostHandler_UpdatePost_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	userID := uuidv7.New()
	postID := uuidv7.New()

	req := httptest.NewRequest(http.MethodPut, "/posts/"+postID.String(), bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Params = gin.Params{{Key: "id", Value: postID.String()}}

	handler.UpdatePost(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserPostHandler_UpdatePost_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	userID := uuidv7.New()
	postID := uuidv7.New()

	title := "Updated Title"
	reqData := dto.UpdatePostRequest{
		Title: &title,
	}
	reqBody, _ := json.Marshal(reqData)

	mockUC.On("UpdatePost", mock.Anything, userID, postID, mock.Anything).Return(nil, entity.ErrNotFound)

	req := httptest.NewRequest(http.MethodPut, "/posts/"+postID.String(), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Params = gin.Params{{Key: "id", Value: postID.String()}}

	handler.UpdatePost(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUserPostHandler_UpdatePost_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	userID := uuidv7.New()
	postID := uuidv7.New()

	titleUpdate := "Updated Title"
	reqData := dto.UpdatePostRequest{
		Title: &titleUpdate,
	}
	reqBody, _ := json.Marshal(reqData)

	mockUC.On("UpdatePost", mock.Anything, userID, postID, mock.Anything).Return(nil, usecase.ErrUnauthorized)

	req := httptest.NewRequest(http.MethodPut, "/posts/"+postID.String(), bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Params = gin.Params{{Key: "id", Value: postID.String()}}

	handler.UpdatePost(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUserPostHandler_UpdatePost_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	userID := uuidv7.New()

	titleInvalid := "Updated Title"
	reqData := dto.UpdatePostRequest{
		Title: &titleInvalid,
	}
	reqBody, _ := json.Marshal(reqData)

	req := httptest.NewRequest(http.MethodPut, "/posts/invalid-uuid", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

	handler.UpdatePost(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserPostHandler_DeletePost_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	userID := uuidv7.New()

	req := httptest.NewRequest(http.MethodDelete, "/posts/invalid-uuid", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

	handler.DeletePost(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserPostHandler_PublishPost_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	userID := uuidv7.New()
	postID := uuidv7.New()

	mockUC.On("PublishPost", mock.Anything, userID, postID).Return(entity.ErrNotFound)

	req := httptest.NewRequest(http.MethodPost, "/posts/"+postID.String()+"/publish", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Params = gin.Params{{Key: "id", Value: postID.String()}}

	handler.PublishPost(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUserPostHandler_PublishPost_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	userID := uuidv7.New()

	req := httptest.NewRequest(http.MethodPost, "/posts/invalid-uuid/publish", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

	handler.PublishPost(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserPostHandler_UnpublishPost_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	userID := uuidv7.New()

	req := httptest.NewRequest(http.MethodPost, "/posts/invalid-uuid/unpublish", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

	handler.UnpublishPost(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserPostHandler_UnpublishPost_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	userID := uuidv7.New()
	postID := uuidv7.New()

	mockUC.On("UnpublishPost", mock.Anything, userID, postID).Return(entity.ErrNotFound)

	req := httptest.NewRequest(http.MethodPost, "/posts/"+postID.String()+"/unpublish", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Params = gin.Params{{Key: "id", Value: postID.String()}}

	handler.UnpublishPost(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUserPostHandler_GetUserPosts_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/users/invalid-uuid/posts", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "user_id", Value: "invalid-uuid"}}

	handler.GetUserPosts(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserPostHandler_GetPostsByTag_InvalidTag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	// Empty tag - handler should still call usecase with empty string
	mockUC.On("GetPostsByTag", mock.Anything, "", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return([]*entity.UserPost{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts/tags/", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "tag", Value: ""}}

	handler.GetPostsByTag(c)

	// Handler returns OK even with empty tag
	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestUserPostHandler_SearchPosts_EmptyQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := new(usecaseMocks.MockUserPostUseCase)
	handler := NewUserPostHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/posts/search", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.SearchPosts(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Comment handler test (already exists in file)

