package contact_test

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/contexts/identity/contact"
	contactHTTP "github.com/basilex/promenade/internal/contexts/identity/contact/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
	"github.com/stretchr/testify/assert"
)

// MockContactUseCase implements minimal contact.IUseCase interface for testing
type MockContactUseCase struct {
	CreateEmailContactFunc   func(ctx context.Context, userID uuidv7.UUID, email, label string, isPrimary bool) (*contact.Contact, error)
	GetContactFunc           func(ctx context.Context, contactID uuidv7.UUID) (*contact.Contact, error)
	GetUserContactsFunc      func(ctx context.Context, userID uuidv7.UUID) ([]*contact.Contact, error)
	DeleteContactFunc        func(ctx context.Context, contactID uuidv7.UUID) error
	VerifyContactFunc        func(ctx context.Context, contactID uuidv7.UUID) error
	SetAsPrimaryFunc         func(ctx context.Context, contactID uuidv7.UUID) error
}

func (m *MockContactUseCase) CreateEmailContact(ctx context.Context, userID uuidv7.UUID, email, label string, isPrimary bool) (*contact.Contact, error) {
	if m.CreateEmailContactFunc != nil {
		return m.CreateEmailContactFunc(ctx, userID, email, label, isPrimary)
	}
	return fakeContact(), nil
}

func (m *MockContactUseCase) GetContact(ctx context.Context, contactID uuidv7.UUID) (*contact.Contact, error) {
	if m.GetContactFunc != nil {
		return m.GetContactFunc(ctx, contactID)
	}
	return fakeContact(), nil
}

func (m *MockContactUseCase) GetUserContacts(ctx context.Context, userID uuidv7.UUID) ([]*contact.Contact, error) {
	if m.GetUserContactsFunc != nil {
		return m.GetUserContactsFunc(ctx, userID)
	}
	return []*contact.Contact{fakeContact()}, nil
}

func (m *MockContactUseCase) DeleteContact(ctx context.Context, contactID uuidv7.UUID) error {
	if m.DeleteContactFunc != nil {
		return m.DeleteContactFunc(ctx, contactID)
	}
	return nil
}

func (m *MockContactUseCase) VerifyContact(ctx context.Context, contactID uuidv7.UUID) error {
	if m.VerifyContactFunc != nil {
		return m.VerifyContactFunc(ctx, contactID)
	}
	return nil
}

func (m *MockContactUseCase) SetAsPrimary(ctx context.Context, contactID uuidv7.UUID) error {
	if m.SetAsPrimaryFunc != nil {
		return m.SetAsPrimaryFunc(ctx, contactID)
	}
	return nil
}

// Stub implementations for other IUseCase methods
func (m *MockContactUseCase) CreatePhoneContact(ctx context.Context, userID uuidv7.UUID, phone, label string, isPrimary bool) (*contact.Contact, error) {
	return fakeContact(), nil
}
func (m *MockContactUseCase) CreateAddressContact(ctx context.Context, userID uuidv7.UUID, street, city, country, postalCode, label string, isPrimary bool) (*contact.Contact, error) {
	return fakeContact(), nil
}
func (m *MockContactUseCase) GetUserContactsByType(ctx context.Context, userID uuidv7.UUID, contactType contact.ContactType) ([]*contact.Contact, error) {
	return []*contact.Contact{fakeContact()}, nil
}
func (m *MockContactUseCase) UpdateContact(ctx context.Context, c *contact.Contact) error {
	return nil
}
func (m *MockContactUseCase) UpdateVisibility(ctx context.Context, contactID uuidv7.UUID, isPublic bool) error {
	return nil
}

func fakeContact() *contact.Contact {
	userID := uuidv7.New()
	c, _ := contact.NewEmailContact(userID, "test@example.com", "Work")
	return c
}

func TestContactHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContactUseCase{
		CreateEmailContactFunc: func(ctx context.Context, userID uuidv7.UUID, email, label string, isPrimary bool) (*contact.Contact, error) {
			return fakeContact(), nil
		},
	}

	handler := contactHTTP.NewContactHandler(mockUC)
	router.POST("/contacts", handler.Create)

	resp := smoke.MakeRequest(t, router, "POST", "/contacts?user_id="+smoke.FakeUUID(), map[string]any{
		"type":  "email",
		"email": "test@example.com",
		"label": "Work",
	})

	smoke.AssertSuccessResponse(t, resp, 201)
}

func TestContactHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()
	mockUC := &MockContactUseCase{}
	handler := contactHTTP.NewContactHandler(mockUC)
	router.POST("/contacts", handler.Create)

	resp := smoke.MakeRequest(t, router, "POST", "/contacts?user_id="+smoke.FakeUUID(), map[string]any{
		"type": "email",
		// Missing email field
	})

	smoke.AssertErrorResponse(t, resp, 400, "BAD_REQUEST")
}

func TestContactHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContactUseCase{
		GetContactFunc: func(ctx context.Context, contactID uuidv7.UUID) (*contact.Contact, error) {
			return fakeContact(), nil
		},
	}

	handler := contactHTTP.NewContactHandler(mockUC)
	router.GET("/contacts/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/contacts/"+smoke.FakeUUID(), nil)

	smoke.AssertSuccessResponse(t, resp, 200)
}

func TestContactHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContactUseCase{
		GetContactFunc: func(ctx context.Context, contactID uuidv7.UUID) (*contact.Contact, error) {
			return nil, contact.ErrNotFound
		},
	}

	handler := contactHTTP.NewContactHandler(mockUC)
	router.GET("/contacts/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/contacts/"+smoke.FakeUUID(), nil)

	smoke.AssertErrorResponse(t, resp, 404, "NOT_FOUND")
}

func TestContactHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContactUseCase{
		GetUserContactsFunc: func(ctx context.Context, userID uuidv7.UUID) ([]*contact.Contact, error) {
			return []*contact.Contact{fakeContact()}, nil
		},
	}

	handler := contactHTTP.NewContactHandler(mockUC)
	router.GET("/contacts", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/contacts?user_id="+smoke.FakeUUID(), nil)

	assert.Equal(t, 200, resp.Code)
}

func TestContactHandler_List_EmptyResult(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContactUseCase{
		GetUserContactsFunc: func(ctx context.Context, userID uuidv7.UUID) ([]*contact.Contact, error) {
			return []*contact.Contact{}, nil
		},
	}

	handler := contactHTTP.NewContactHandler(mockUC)
	router.GET("/contacts", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/contacts?user_id="+smoke.FakeUUID(), nil)

	assert.Equal(t, 200, resp.Code)
}

func TestContactHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContactUseCase{
		DeleteContactFunc: func(ctx context.Context, contactID uuidv7.UUID) error {
			return nil
		},
	}

	handler := contactHTTP.NewContactHandler(mockUC)
	router.DELETE("/contacts/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/contacts/"+smoke.FakeUUID(), nil)

	assert.Equal(t, 200, resp.Code)
}

func TestContactHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContactUseCase{
		DeleteContactFunc: func(ctx context.Context, contactID uuidv7.UUID) error {
			return contact.ErrNotFound
		},
	}

	handler := contactHTTP.NewContactHandler(mockUC)
	router.DELETE("/contacts/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/contacts/"+smoke.FakeUUID(), nil)

	smoke.AssertErrorResponse(t, resp, 404, "NOT_FOUND")
}
