package contact_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/contexts/identity/contact"
	contactHTTP "github.com/basilex/promenade/internal/contexts/identity/contact/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockContactUseCase - minimal mock for smoke tests
type MockContactUseCase struct {
	mock.Mock
}

func (m *MockContactUseCase) CreateEmailContact(ctx context.Context, userID uuidv7.UUID, email, label string, isPrimary bool) (*contact.Contact, error) {
	args := m.Called(ctx, userID, email, label, isPrimary)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Contact), args.Error(1)
}

func (m *MockContactUseCase) CreatePhoneContact(ctx context.Context, userID uuidv7.UUID, phone, label string, isPrimary bool) (*contact.Contact, error) {
	args := m.Called(ctx, userID, phone, label, isPrimary)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Contact), args.Error(1)
}

func (m *MockContactUseCase) CreateAddressContact(ctx context.Context, userID uuidv7.UUID, street, city, postalCode, country, label string, isPrimary bool) (*contact.Contact, error) {
	args := m.Called(ctx, userID, street, city, postalCode, country, label, isPrimary)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Contact), args.Error(1)
}

func (m *MockContactUseCase) GetContact(ctx context.Context, contactID uuidv7.UUID) (*contact.Contact, error) {
	args := m.Called(ctx, contactID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Contact), args.Error(1)
}

func (m *MockContactUseCase) GetUserContacts(ctx context.Context, userID uuidv7.UUID) ([]*contact.Contact, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contact.Contact), args.Error(1)
}

func (m *MockContactUseCase) GetUserContactsByType(ctx context.Context, userID uuidv7.UUID, contactType contact.ContactType) ([]*contact.Contact, error) {
	args := m.Called(ctx, userID, contactType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*contact.Contact), args.Error(1)
}

func (m *MockContactUseCase) UpdateContact(ctx context.Context, c *contact.Contact) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockContactUseCase) SetAsPrimary(ctx context.Context, contactID uuidv7.UUID) error {
	args := m.Called(ctx, contactID)
	return args.Error(0)
}

func (m *MockContactUseCase) UpdateVisibility(ctx context.Context, contactID uuidv7.UUID, isPublic bool) error {
	args := m.Called(ctx, contactID, isPublic)
	return args.Error(0)
}

func (m *MockContactUseCase) DeleteContact(ctx context.Context, contactID uuidv7.UUID) error {
	args := m.Called(ctx, contactID)
	return args.Error(0)
}

func (m *MockContactUseCase) VerifyContact(ctx context.Context, contactID uuidv7.UUID) error {
	args := m.Called(ctx, contactID)
	return args.Error(0)
}

func setupContactRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// TestContactHandler_Smoke - smoke tests for Contact handler
func TestContactHandler_Smoke(t *testing.T) {
	mockUC := new(MockContactUseCase)
	handler := contactHTTP.NewContactHandler(mockUC)
	router := setupContactRouter()

	// Register routes
	router.POST("/contacts", handler.Create)
	router.GET("/contacts/:id", handler.GetByID)
	router.GET("/contacts", handler.List)
	router.PUT("/contacts/:id", handler.Update)
	router.DELETE("/contacts/:id", handler.Delete)
	router.POST("/contacts/:id/verify", handler.Verify)
	router.POST("/contacts/:id/set-primary", handler.SetPrimary)

	t.Run("Create email contact returns 201", func(t *testing.T) {
		userID := uuidv7.New()
		contactID := uuidv7.New()
		email := "test@example.com"

		c := &contact.Contact{
			ID:     contactID,
			UserID: userID,
			Type:   contact.ContactTypeEmail,
			Label:  "Work",
		}
		mockUC.On("CreateEmailContact", mock.Anything, mock.AnythingOfType("uuid.UUID"), email, "Work", false).Return(c, nil).Once()

		reqBody := contactHTTP.CreateContactRequest{
			Type:  "email",
			Label: "Work",
			Email: &email,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/contacts?user_id="+userID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Create should return 201")
	})

	t.Run("GetByID returns 200", func(t *testing.T) {
		contactID := uuidv7.New()
		c := &contact.Contact{
			ID:     contactID,
			UserID: uuidv7.New(),
			Type:   contact.ContactTypeEmail,
			Label:  "Work",
		}
		mockUC.On("GetContact", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(c, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/contacts/"+contactID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "GetByID should return 200")
	})

	t.Run("List returns 200", func(t *testing.T) {
		userID := uuidv7.New()
		contacts := []*contact.Contact{
			{ID: uuidv7.New(), UserID: userID, Type: contact.ContactTypeEmail, Label: "Work"},
			{ID: uuidv7.New(), UserID: userID, Type: contact.ContactTypePhone, Label: "Mobile"},
		}
		mockUC.On("GetUserContacts", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(contacts, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/contacts?user_id="+userID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "List should return 200")
	})

	t.Run("Update returns 200", func(t *testing.T) {
		contactID := uuidv7.New()
		c := &contact.Contact{
			ID:     contactID,
			UserID: uuidv7.New(),
			Type:   contact.ContactTypeEmail,
			Label:  "Work",
		}
		// Handler calls GetContact first, then UpdateContact
		mockUC.On("GetContact", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(c, nil).Once()
		mockUC.On("UpdateContact", mock.Anything, mock.AnythingOfType("*contact.Contact")).Return(nil).Once()

		label := "Updated Label"
		reqBody := contactHTTP.UpdateContactRequest{
			Label: &label,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/contacts/"+contactID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Update should return 200")
	})

	t.Run("Delete returns 204", func(t *testing.T) {
		contactID := uuidv7.New()
		mockUC.On("DeleteContact", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/contacts/"+contactID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Delete should return 200")
	})

	t.Run("Verify returns 200", func(t *testing.T) {
		contactID := uuidv7.New()
		c := &contact.Contact{
			ID:         contactID,
			UserID:     uuidv7.New(),
			Type:       contact.ContactTypeEmail,
			Label:      "Work",
			IsVerified: true,
		}
		mockUC.On("VerifyContact", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil).Once()
		mockUC.On("GetContact", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(c, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/contacts/"+contactID.String()+"/verify", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Verify should return 200")
	})

	t.Run("SetPrimary returns 200", func(t *testing.T) {
		contactID := uuidv7.New()
		c := &contact.Contact{
			ID:        contactID,
			UserID:    uuidv7.New(),
			Type:      contact.ContactTypeEmail,
			Label:     "Work",
			IsPrimary: true,
		}
		mockUC.On("SetAsPrimary", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil).Once()
		mockUC.On("GetContact", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(c, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/contacts/"+contactID.String()+"/set-primary", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "SetPrimary should return 200")
	})
}
