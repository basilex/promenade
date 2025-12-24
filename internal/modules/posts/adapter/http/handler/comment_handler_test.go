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
	usecaseMocks "github.com/basilex/promenade/internal/modules/posts/usecase/mocks"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// TestCommentHandler_CreateComment tests comment creation handler
//  IModule-independent: uses only module types and mocks
func TestCommentHandler_CreateComment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful top-level comment", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockCommentUseCase)
		handler := NewCommentHandler(mockUC)

		userID := uuidv7.New()
		postID := uuidv7.New()

		createReq := dto.CreateCommentRequest{
			Content:  "Great post!",
			ParentID: nil,
		}

		expectedComment := &entity.Comment{
			ID:      uuidv7.New(),
			PostID:  postID,
			UserID:  userID,
			Content: createReq.Content,
			Depth:   0,
		}

		mockUC.On("CreateComment",
			mock.Anything,
			postID,
			userID,
			createReq.Content,
			createReq.ParentID,
		).Return(expectedComment, nil)

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/posts/"+postID.String()+"/comments", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)
		c.Params = gin.Params{{Key: "postId", Value: postID.String()}}

		// Act
		handler.CreateComment(c)

		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, true, response["success"])
		data, ok := response["data"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, createReq.Content, data["content"])

		mockUC.AssertExpectations(t)
	})

	t.Run("successful reply comment", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockCommentUseCase)
		handler := NewCommentHandler(mockUC)

		userID := uuidv7.New()
		postID := uuidv7.New()
		parentID := uuidv7.New()

		createReq := dto.CreateCommentRequest{
			Content:  "I agree!",
			ParentID: &parentID,
		}

		expectedComment := &entity.Comment{
			ID:       uuidv7.New(),
			PostID:   postID,
			UserID:   userID,
			ParentID: &parentID,
			Content:  createReq.Content,
			Depth:    1,
		}

		mockUC.On("CreateComment",
			mock.Anything,
			postID,
			userID,
			createReq.Content,
			&parentID,
		).Return(expectedComment, nil)

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/posts/"+postID.String()+"/comments", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)
		c.Params = gin.Params{{Key: "postId", Value: postID.String()}}

		// Act
		handler.CreateComment(c)

		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("unauthorized - no userID", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockCommentUseCase)
		handler := NewCommentHandler(mockUC)

		postID := uuidv7.New()
		createReq := dto.CreateCommentRequest{
			Content: "Comment",
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/posts/"+postID.String()+"/comments", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "postId", Value: postID.String()}}
		// Don't set userID

		// Act
		handler.CreateComment(c)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUC.AssertNotCalled(t, "CreateComment")
	})

	t.Run("invalid post ID", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockCommentUseCase)
		handler := NewCommentHandler(mockUC)

		userID := uuidv7.New()
		createReq := dto.CreateCommentRequest{
			Content: "Comment",
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/posts/invalid-id/comments", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)
		c.Params = gin.Params{{Key: "postId", Value: "invalid-id"}}

		// Act
		handler.CreateComment(c)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockUC.AssertNotCalled(t, "CreateComment")
	})

	t.Run("invalid request body", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockCommentUseCase)
		handler := NewCommentHandler(mockUC)

		userID := uuidv7.New()
		postID := uuidv7.New()

		// Malformed JSON
		body := []byte(`{"content": `)
		req := httptest.NewRequest(http.MethodPost, "/posts/"+postID.String()+"/comments", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)
		c.Params = gin.Params{{Key: "postId", Value: postID.String()}}

		// Act
		handler.CreateComment(c)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockUC.AssertNotCalled(t, "CreateComment")
	})
}

// TestCommentHandler_GetComment tests comment retrieval handler
func TestCommentHandler_GetComment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful retrieval", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockCommentUseCase)
		handler := NewCommentHandler(mockUC)

		commentID := uuidv7.New()
		expectedComment := &entity.Comment{
			ID:      commentID,
			PostID:  uuidv7.New(),
			UserID:  uuidv7.New(),
			Content: "Test comment",
		}

		mockUC.On("GetComment", mock.Anything, commentID).Return(expectedComment, nil)

		req := httptest.NewRequest(http.MethodGet, "/comments/"+commentID.String(), nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: commentID.String()}}

		// Act
		handler.GetComment(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, true, response["success"])
		data, ok := response["data"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, expectedComment.Content, data["content"])

		mockUC.AssertExpectations(t)
	})

	t.Run("invalid comment ID", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockCommentUseCase)
		handler := NewCommentHandler(mockUC)

		req := httptest.NewRequest(http.MethodGet, "/comments/invalid-id", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "invalid-id"}}

		// Act
		handler.GetComment(c)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockUC.AssertNotCalled(t, "GetComment")
	})

	t.Run("comment not found", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockCommentUseCase)
		handler := NewCommentHandler(mockUC)

		commentID := uuidv7.New()
		mockUC.On("GetComment", mock.Anything, commentID).Return(nil, errors.New("not found"))

		req := httptest.NewRequest(http.MethodGet, "/comments/"+commentID.String(), nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: commentID.String()}}

		// Act
		handler.GetComment(c)

		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})
}

// TestCommentHandler_UpdateComment tests comment update handler
func TestCommentHandler_UpdateComment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful update", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockCommentUseCase)
		handler := NewCommentHandler(mockUC)

		userID := uuidv7.New()
		commentID := uuidv7.New()

		updateReq := dto.UpdateCommentRequest{
			Content: "Updated content",
		}

		updatedComment := &entity.Comment{
			ID:      commentID,
			PostID:  uuidv7.New(),
			UserID:  userID,
			Content: updateReq.Content,
		}

		mockUC.On("UpdateComment",
			mock.Anything,
			userID,
			commentID,
			updateReq.Content,
		).Return(updatedComment, nil)

		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(http.MethodPut, "/comments/"+commentID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)
		c.Params = gin.Params{{Key: "id", Value: commentID.String()}}

		// Act
		handler.UpdateComment(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, true, response["success"])
		mockUC.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockCommentUseCase)
		handler := NewCommentHandler(mockUC)

		commentID := uuidv7.New()
		updateReq := dto.UpdateCommentRequest{
			Content: "Updated content",
		}

		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(http.MethodPut, "/comments/"+commentID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		// Don't set userID
		c.Params = gin.Params{{Key: "id", Value: commentID.String()}}

		// Act
		handler.UpdateComment(c)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUC.AssertNotCalled(t, "UpdateComment")
	})
}

// TestCommentHandler_DeleteComment tests comment deletion handler
func TestCommentHandler_DeleteComment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful deletion", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockCommentUseCase)
		handler := NewCommentHandler(mockUC)

		userID := uuidv7.New()
		commentID := uuidv7.New()

		mockUC.On("DeleteComment", mock.Anything, userID, commentID).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/comments/"+commentID.String(), nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)
		c.Params = gin.Params{{Key: "id", Value: commentID.String()}}

		// Act
		handler.DeleteComment(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		// Arrange
		mockUC := new(usecaseMocks.MockCommentUseCase)
		handler := NewCommentHandler(mockUC)

		commentID := uuidv7.New()

		req := httptest.NewRequest(http.MethodDelete, "/comments/"+commentID.String(), nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		// Don't set userID
		c.Params = gin.Params{{Key: "id", Value: commentID.String()}}

		// Act
		handler.DeleteComment(c)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUC.AssertNotCalled(t, "DeleteComment")
	})
}
