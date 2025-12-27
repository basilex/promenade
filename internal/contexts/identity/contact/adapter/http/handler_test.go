package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/contexts/identity/contact"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockUseCase is a mock implementation of contact.IUseCase
type MockUseCase struct {
	mock.Mock
}

func (m *MockUseCase) CreateEmailContact(ctx context.Context, userID uuidv7.UUID, email, label string, isPrimary bool) (*contact.Contact, error) {
	args := m.Called(ctx, userID, email, label, isPrimary)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Contact), args.Error(1)
}

func (m *MockUseCase) CreatePhoneContact(ctx context.Context, userID uuidv7.UUID, phone, label string, isPrimary bool) (*contact.Contact, error) {
	args := m.Called(ctx, userID, phone, label, isPrimary)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Contact), args.Error(1)
}

func (m *MockUseCase) CreateAddressContact(ctx context.Context, userID uuidv7.UUID, street, city, country, postalCode, label string, isPrimary bool) (*contact.Contact, error) {
	args := m.Called(ctx, userID, street, city, country, postalCode, label, isPrimary)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Contact), args.Error(1)
}

func (m *MockUseCase) GetContact(ctx context.Context, contactID uuidv7.UUID) (*contact.Contact, error) {
	args := m.Called(ctx, contactID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Contact), args.Error(1)
}

func (m *MockUseCase) GetUserContacts(ctx context.Context, userID uuidv7.UUID) ([]*contact.Contact, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contact.Contact), args.Error(1)
}

func (m *MockUseCase) GetUserContactsByType(ctx context.Context, userID uuidv7.UUID, contactType contact.ContactType) ([]*contact.Contact, error) {
	args := m.Called(ctx, userID, contactType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contact.Contact), args.Error(1)
}

func (m *MockUseCase) UpdateContact(ctx context.Context, c *contact.Contact) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockUseCase) DeleteContact(ctx context.Context, contactID uuidv7.UUID) error {
	args := m.Called(ctx, contactID)
	return args.Error(0)
}

func (m *MockUseCase) SetAsPrimary(ctx context.Context, contactID uuidv7.UUID) error {
	args := m.Called(ctx, contactID)
	return args.Error(0)
}

func (m *MockUseCase) VerifyContact(ctx context.Context, contactID uuidv7.UUID) error {
	args := m.Called(ctx, contactID)
	return args.Error(0)
}

func (m *MockUseCase) UpdateVisibility(ctx context.Context, contactID uuidv7.UUID, isPublic bool) error {
	args := m.Called(ctx, contactID, isPublic)
	return args.Error(0)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestContactHandler_Create_Success(t *testing.T) {
	mockUseCase := new(MockUseCase)
	handler := NewContactHandler(mockUseCase)
	router := setupTestRouter()

	userID := uuidv7.New()
	email := "test@example.com"

	reqBody := CreateContactRequest{
		Type:  "email",
		Label: "Work",
		Email: &email,
	}

	expectedContact, _ := contact.NewEmailContact(userID, email, "Work")
	mockUseCase.On("CreateEmailContact", mock.Anything, userID, email, "Work", false).Return(expectedContact, nil)

	router.POST("/contacts", handler.Create)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/contacts?user_id=%s", userID.String()), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestContactHandler_Create_MissingUserID(t *testing.T) {
	mockUseCase := new(MockUseCase)
	handler := NewContactHandler(mockUseCase)
	router := setupTestRouter()

	email := "test@example.com"
	reqBody := CreateContactRequest{
		Type:  "email",
		Label: "Work",
		Email: &email,
	}

	router.POST("/contacts", handler.Create)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/contacts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestContactHandler_GetByID_Success(t *testing.T) {
	mockUseCase := new(MockUseCase)
	handler := NewContactHandler(mockUseCase)
	router := setupTestRouter()

	contactID := uuidv7.New()
	userID := uuidv7.New()
	expectedContact, _ := contact.NewEmailContact(userID, "test@example.com", "Work")
	expectedContact.ID = contactID

	mockUseCase.On("GetContact", mock.Anything, contactID).Return(expectedContact, nil)

	router.GET("/contacts/:id", handler.GetByID)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/contacts/%s", contactID.String()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestContactHandler_GetByID_NotFound(t *testing.T) {
	mockUseCase := new(MockUseCase)
	handler := NewContactHandler(mockUseCase)
	router := setupTestRouter()

	contactID := uuidv7.New()

	mockUseCase.On("GetContact", mock.Anything, contactID).Return(nil, contact.ErrNotFound)

	router.GET("/contacts/:id", handler.GetByID)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/contacts/%s", contactID.String()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestContactHandler_Delete_Success(t *testing.T) {
	mockUseCase := new(MockUseCase)
	handler := NewContactHandler(mockUseCase)
	router := setupTestRouter()

	contactID := uuidv7.New()

	mockUseCase.On("DeleteContact", mock.Anything, contactID).Return(nil)

	router.DELETE("/contacts/:id", handler.Delete)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/contacts/%s", contactID.String()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestContactHandler_List_Success(t *testing.T) {
	mockUseCase := new(MockUseCase)
	handler := NewContactHandler(mockUseCase)
	router := setupTestRouter()

	userID := uuidv7.New()
	expectedContact, _ := contact.NewEmailContact(userID, "test@example.com", "Work")
	contacts := []*contact.Contact{expectedContact}

	mockUseCase.On("GetUserContacts", mock.Anything, userID).Return(contacts, nil)

	router.GET("/contacts", handler.List)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/contacts?user_id=%s", userID.String()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestContactHandler_Verify_Success(t *testing.T) {
	mockUseCase := new(MockUseCase)
	handler := NewContactHandler(mockUseCase)
	router := setupTestRouter()

	contactID := uuidv7.New()
	userID := uuidv7.New()
	verifiedContact, _ := contact.NewEmailContact(userID, "test@example.com", "Work")
	verifiedContact.ID = contactID
	verifiedContact.IsVerified = true

	mockUseCase.On("VerifyContact", mock.Anything, contactID).Return(nil)
	mockUseCase.On("GetContact", mock.Anything, contactID).Return(verifiedContact, nil)

	router.PUT("/contacts/:id/verify", handler.Verify)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/contacts/%s/verify", contactID.String()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestContactHandler_SetPrimary_Success(t *testing.T) {
	mockUseCase := new(MockUseCase)
	handler := NewContactHandler(mockUseCase)
	router := setupTestRouter()

	contactID := uuidv7.New()
	userID := uuidv7.New()
	primaryContact, _ := contact.NewEmailContact(userID, "test@example.com", "Work")
	primaryContact.ID = contactID
	primaryContact.IsPrimary = true

	mockUseCase.On("SetAsPrimary", mock.Anything, contactID).Return(nil)
	mockUseCase.On("GetContact", mock.Anything, contactID).Return(primaryContact, nil)

	router.PUT("/contacts/:id/primary", handler.SetPrimary)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/contacts/%s/primary", contactID.String()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}
