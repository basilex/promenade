package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/profiles/adapter/http/dto"
	"github.com/basilex/promenade/internal/modules/profiles/entity"
	"github.com/basilex/promenade/internal/modules/profiles/usecase"
	ucmocks "github.com/basilex/promenade/internal/modules/profiles/usecase/mocks"
	"github.com/basilex/promenade/pkg/uuidv7"
)

//  IModule-independent test: imports only module and pkg, no core dependencies

func setupContactTest() (*gin.Engine, *ucmocks.MockUserContactUseCase) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUC := new(ucmocks.MockUserContactUseCase)
	return router, mockUC
}

func TestCreateContact(t *testing.T) {
	router, mockUC := setupContactTest()
	handler := NewUserContactHandler(mockUC)
	router.POST("/contacts", handler.CreateContact)

	userID := uuidv7.New()
	contactValue := "test@example.com"

	t.Run("success", func(t *testing.T) {
		reqBody := dto.CreateUserContactRequest{
			ContactType:  "email",
			ContactValue: contactValue,
			IsPublic:     true,
		}
		body, _ := json.Marshal(reqBody)

		contact := &entity.UserContact{
			ID:           uuidv7.New(),
			UserID:       userID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: contactValue,
			IsPublic:     true,
		}

		mockUC.On("CreateContact", mock.Anything, userID, entity.ContactTypeEmail, contactValue, 
			mock.Anything, true, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(contact, nil).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/contacts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("user_id", userID)
		
		handler.CreateContact(ctx)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, true, response["success"])
		mockUC.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		reqBody := dto.CreateUserContactRequest{
			ContactType:  "email",
			ContactValue: contactValue,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/contacts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		// No user_id set
		
		handler.CreateContact(ctx)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/contacts", bytes.NewBufferString("{invalid"))
		req.Header.Set("Content-Type", "application/json")
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("user_id", userID)
		
		handler.CreateContact(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid contact type", func(t *testing.T) {
		reqBody := dto.CreateUserContactRequest{
			ContactType:  "invalid",
			ContactValue: contactValue,
		}
		body, _ := json.Marshal(reqBody)

		mockUC.On("CreateContact", mock.Anything, userID, entity.ContactType("invalid"), contactValue,
			mock.Anything, false, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, usecase.ErrInvalidContactType).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/contacts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("user_id", userID)
		
		handler.CreateContact(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("invalid availability", func(t *testing.T) {
		availFrom := "18:00:00"
		availTo := "09:00:00"
		reqBody := dto.CreateUserContactRequest{
			ContactType:   "email",
			ContactValue:  contactValue,
			AvailableFrom: &availFrom,
			AvailableTo:   &availTo, // Invalid: from > to
		}
		body, _ := json.Marshal(reqBody)

		mockUC.On("CreateContact", mock.Anything, userID, entity.ContactTypeEmail, contactValue,
			mock.Anything, false, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, usecase.ErrInvalidAvailability).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/contacts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("user_id", userID)
		
		handler.CreateContact(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestGetContact(t *testing.T) {
	router, mockUC := setupContactTest()
	handler := NewUserContactHandler(mockUC)
	router.GET("/contacts/:id", handler.GetContact)

	contactID := uuidv7.New()
	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:           contactID,
			UserID:       userID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: "test@example.com",
			IsPublic:     true,
		}

		mockUC.On("GetContact", mock.Anything, contactID, userID).Return(contact, nil).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/contacts/"+contactID.String(), nil)
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: contactID.String()}}
		ctx.Set("user_id", userID)
		
		handler.GetContact(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, true, response["success"])
		mockUC.AssertExpectations(t)
	})

	t.Run("invalid contact ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/contacts/invalid-uuid", nil)
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}
		ctx.Set("user_id", userID)
		
		handler.GetContact(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/contacts/"+contactID.String(), nil)
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: contactID.String()}}
		// No user_id set
		
		handler.GetContact(ctx)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("contact not found", func(t *testing.T) {
		mockUC.On("GetContact", mock.Anything, contactID, userID).Return(nil, usecase.ErrContactNotFound).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/contacts/"+contactID.String(), nil)
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: contactID.String()}}
		ctx.Set("user_id", userID)
		
		handler.GetContact(ctx)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		mockUC.On("GetContact", mock.Anything, contactID, userID).Return(nil, usecase.ErrUnauthorized).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/contacts/"+contactID.String(), nil)
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: contactID.String()}}
		ctx.Set("user_id", userID)
		
		handler.GetContact(ctx)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestUpdateContact(t *testing.T) {
	router, mockUC := setupContactTest()
	handler := NewUserContactHandler(mockUC)
	router.PUT("/contacts/:id", handler.UpdateContact)

	contactID := uuidv7.New()
	userID := uuidv7.New()
	newValue := "new@example.com"

	t.Run("success", func(t *testing.T) {
		reqBody := dto.UpdateUserContactRequest{
			ContactType:  "email",
			ContactValue: newValue,
			IsPublic:     true,
		}
		body, _ := json.Marshal(reqBody)

		contact := &entity.UserContact{
			ID:           contactID,
			UserID:       userID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: newValue,
			IsPublic:     true,
		}

		mockUC.On("UpdateContact", mock.Anything, contactID, userID, entity.ContactTypeEmail, newValue,
			mock.Anything, true, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(contact, nil).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/contacts/"+contactID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: contactID.String()}}
		ctx.Set("user_id", userID)
		
		handler.UpdateContact(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, true, response["success"])
		mockUC.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		reqBody := dto.UpdateUserContactRequest{
			ContactType:  "email",
			ContactValue: newValue,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/contacts/"+contactID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: contactID.String()}}
		// No user_id set
		
		handler.UpdateContact(ctx)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("contact not found", func(t *testing.T) {
		reqBody := dto.UpdateUserContactRequest{
			ContactType:  "email",
			ContactValue: newValue,
		}
		body, _ := json.Marshal(reqBody)

		mockUC.On("UpdateContact", mock.Anything, contactID, userID, entity.ContactTypeEmail, newValue,
			mock.Anything, false, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, usecase.ErrContactNotFound).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/contacts/"+contactID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: contactID.String()}}
		ctx.Set("user_id", userID)
		
		handler.UpdateContact(ctx)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		reqBody := dto.UpdateUserContactRequest{
			ContactType:  "email",
			ContactValue: newValue,
		}
		body, _ := json.Marshal(reqBody)

		mockUC.On("UpdateContact", mock.Anything, contactID, userID, entity.ContactTypeEmail, newValue,
			mock.Anything, false, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, usecase.ErrUnauthorized).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/contacts/"+contactID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: contactID.String()}}
		ctx.Set("user_id", userID)
		
		handler.UpdateContact(ctx)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestDeleteContact(t *testing.T) {
	router, mockUC := setupContactTest()
	handler := NewUserContactHandler(mockUC)
	router.DELETE("/contacts/:id", handler.DeleteContact)

	contactID := uuidv7.New()
	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		mockUC.On("DeleteContact", mock.Anything, contactID, userID).Return(nil).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/contacts/"+contactID.String(), nil)
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: contactID.String()}}
		ctx.Set("user_id", userID)
		
		handler.DeleteContact(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/contacts/"+contactID.String(), nil)
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: contactID.String()}}
		// No user_id set
		
		handler.DeleteContact(ctx)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("contact not found", func(t *testing.T) {
		mockUC.On("DeleteContact", mock.Anything, contactID, userID).Return(usecase.ErrContactNotFound).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/contacts/"+contactID.String(), nil)
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: contactID.String()}}
		ctx.Set("user_id", userID)
		
		handler.DeleteContact(ctx)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		mockUC.On("DeleteContact", mock.Anything, contactID, userID).Return(usecase.ErrUnauthorized).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/contacts/"+contactID.String(), nil)
		
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: contactID.String()}}
		ctx.Set("user_id", userID)
		
		handler.DeleteContact(ctx)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockUC.AssertExpectations(t)
	})
}
