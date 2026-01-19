package interaction_test

import (
	"context"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction"
	interactionHTTP "github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction/adapter/http"
	interactionAggregate "github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

type MockInteractionUseCase struct {
	CreateInteractionFunc    func(ctx context.Context, customerID uuidv7.UUID, companyID *uuidv7.UUID, interactionType string, direction string, subject string, description string, createdBy uuidv7.UUID, startedAt time.Time) (*interactionAggregate.Interaction, error)
	GetInteractionFunc       func(ctx context.Context, id uuidv7.UUID) (*interactionAggregate.Interaction, error)
	DeleteInteractionFunc    func(ctx context.Context, id uuidv7.UUID) error
	ListByCustomerFunc       func(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*interactionAggregate.Interaction, int64, error)
	UpdateContentFunc        func(ctx context.Context, id uuidv7.UUID, subject, description string) (*interactionAggregate.Interaction, error)
	SetOutcomeFunc           func(ctx context.Context, id uuidv7.UUID, outcome string) (*interactionAggregate.Interaction, error)
	EndInteractionFunc       func(ctx context.Context, id uuidv7.UUID, endedAt time.Time) (*interactionAggregate.Interaction, error)
	SetFollowUpFunc          func(ctx context.Context, id uuidv7.UUID, required bool, followUpDate *time.Time, notes string) (*interactionAggregate.Interaction, error)
	AddAttendeeFunc          func(ctx context.Context, id uuidv7.UUID, attendeeID uuidv7.UUID) (*interactionAggregate.Interaction, error)
	RemoveAttendeeFunc       func(ctx context.Context, id uuidv7.UUID, attendeeID uuidv7.UUID) (*interactionAggregate.Interaction, error)
	ListByCompanyFunc        func(ctx context.Context, companyID uuidv7.UUID, page, pageSize int) ([]*interactionAggregate.Interaction, int64, error)
	ListByTypeFunc           func(ctx context.Context, interactionType string, page, pageSize int) ([]*interactionAggregate.Interaction, int64, error)
	ListByCreatedByFunc      func(ctx context.Context, createdBy uuidv7.UUID, page, pageSize int) ([]*interactionAggregate.Interaction, int64, error)
	ListPendingFollowUpsFunc func(ctx context.Context, page, pageSize int) ([]*interactionAggregate.Interaction, int64, error)
}

func (m *MockInteractionUseCase) CreateInteraction(ctx context.Context, customerID uuidv7.UUID, companyID *uuidv7.UUID, interactionType string, direction string, subject string, description string, createdBy uuidv7.UUID, startedAt time.Time) (*interactionAggregate.Interaction, error) {
	if m.CreateInteractionFunc != nil {
		return m.CreateInteractionFunc(ctx, customerID, companyID, interactionType, direction, subject, description, createdBy, startedAt)
	}
	return nil, nil
}

func (m *MockInteractionUseCase) GetInteraction(ctx context.Context, id uuidv7.UUID) (*interactionAggregate.Interaction, error) {
	if m.GetInteractionFunc != nil {
		return m.GetInteractionFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockInteractionUseCase) UpdateContent(ctx context.Context, id uuidv7.UUID, subject, description string) (*interactionAggregate.Interaction, error) {
	if m.UpdateContentFunc != nil {
		return m.UpdateContentFunc(ctx, id, subject, description)
	}
	return nil, nil
}

func (m *MockInteractionUseCase) SetOutcome(ctx context.Context, id uuidv7.UUID, outcome string) (*interactionAggregate.Interaction, error) {
	if m.SetOutcomeFunc != nil {
		return m.SetOutcomeFunc(ctx, id, outcome)
	}
	return nil, nil
}

func (m *MockInteractionUseCase) EndInteraction(ctx context.Context, id uuidv7.UUID, endedAt time.Time) (*interactionAggregate.Interaction, error) {
	if m.EndInteractionFunc != nil {
		return m.EndInteractionFunc(ctx, id, endedAt)
	}
	return nil, nil
}

func (m *MockInteractionUseCase) SetFollowUp(ctx context.Context, id uuidv7.UUID, required bool, followUpDate *time.Time, notes string) (*interactionAggregate.Interaction, error) {
	if m.SetFollowUpFunc != nil {
		return m.SetFollowUpFunc(ctx, id, required, followUpDate, notes)
	}
	return nil, nil
}

func (m *MockInteractionUseCase) AddAttendee(ctx context.Context, id uuidv7.UUID, attendeeID uuidv7.UUID) (*interactionAggregate.Interaction, error) {
	if m.AddAttendeeFunc != nil {
		return m.AddAttendeeFunc(ctx, id, attendeeID)
	}
	return nil, nil
}

func (m *MockInteractionUseCase) RemoveAttendee(ctx context.Context, id uuidv7.UUID, attendeeID uuidv7.UUID) (*interactionAggregate.Interaction, error) {
	if m.RemoveAttendeeFunc != nil {
		return m.RemoveAttendeeFunc(ctx, id, attendeeID)
	}
	return nil, nil
}

func (m *MockInteractionUseCase) ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*interactionAggregate.Interaction, int64, error) {
	if m.ListByCustomerFunc != nil {
		return m.ListByCustomerFunc(ctx, customerID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockInteractionUseCase) ListByCompany(ctx context.Context, companyID uuidv7.UUID, page, pageSize int) ([]*interactionAggregate.Interaction, int64, error) {
	if m.ListByCompanyFunc != nil {
		return m.ListByCompanyFunc(ctx, companyID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockInteractionUseCase) ListByType(ctx context.Context, interactionType string, page, pageSize int) ([]*interactionAggregate.Interaction, int64, error) {
	if m.ListByTypeFunc != nil {
		return m.ListByTypeFunc(ctx, interactionType, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockInteractionUseCase) ListByCreatedBy(ctx context.Context, createdBy uuidv7.UUID, page, pageSize int) ([]*interactionAggregate.Interaction, int64, error) {
	if m.ListByCreatedByFunc != nil {
		return m.ListByCreatedByFunc(ctx, createdBy, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockInteractionUseCase) ListPendingFollowUps(ctx context.Context, page, pageSize int) ([]*interactionAggregate.Interaction, int64, error) {
	if m.ListPendingFollowUpsFunc != nil {
		return m.ListPendingFollowUpsFunc(ctx, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockInteractionUseCase) DeleteInteraction(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteInteractionFunc != nil {
		return m.DeleteInteractionFunc(ctx, id)
	}
	return nil
}

func fakeInteraction() *interactionAggregate.Interaction {
	customerID := uuidv7.New()
	createdBy := uuidv7.New()
	i, _ := interactionAggregate.NewInteraction(
		customerID,
		nil,
		interactionAggregate.InteractionTypeCall,
		interactionAggregate.InteractionDirectionInbound,
		"Test Call",
		"Test interaction description",
		createdBy,
		time.Now(),
	)
	return i
}

func TestInteractionHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInteractionUseCase{
		CreateInteractionFunc: func(ctx context.Context, customerID uuidv7.UUID, companyID *uuidv7.UUID, interactionType string, direction string, subject string, description string, createdBy uuidv7.UUID, startedAt time.Time) (*interactionAggregate.Interaction, error) {
			return fakeInteraction(), nil
		},
	}

	handler := interactionHTTP.NewInteractionHandler(mockUC)
	router.POST("/interactions", handler.Create)

	resp := smoke.MakeRequest(t, router, "POST", "/interactions", map[string]any{
		"customer_id": smoke.FakeUUID(),
		"type":        "call",
		"direction":   "inbound",
		"subject":     "Test Call",
		"description": "Test description",
		"created_by":  smoke.FakeUUID(),
		"started_at":  time.Now().Format(time.RFC3339),
	})

	smoke.AssertSuccessResponse(t, resp, 201)
}

func TestInteractionHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInteractionUseCase{}
	handler := interactionHTTP.NewInteractionHandler(mockUC)
	router.POST("/interactions", handler.Create)

	resp := smoke.MakeRequest(t, router, "POST", "/interactions", map[string]any{
		"type":        "call",
		"direction":   "inbound",
		"subject":     "Test",
		"description": "Test",
	})

	smoke.AssertErrorResponse(t, resp, 400, "BAD_REQUEST")
}

func TestInteractionHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInteractionUseCase{
		GetInteractionFunc: func(ctx context.Context, id uuidv7.UUID) (*interactionAggregate.Interaction, error) {
			return fakeInteraction(), nil
		},
	}

	handler := interactionHTTP.NewInteractionHandler(mockUC)
	router.GET("/interactions/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/interactions/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, resp, 200)
}

func TestInteractionHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInteractionUseCase{
		GetInteractionFunc: func(ctx context.Context, id uuidv7.UUID) (*interactionAggregate.Interaction, error) {
			return nil, interaction.ErrInteractionNotFound
		},
	}

	handler := interactionHTTP.NewInteractionHandler(mockUC)
	router.GET("/interactions/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/interactions/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, resp, 404, "NOT_FOUND")
}

func TestInteractionHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInteractionUseCase{
		DeleteInteractionFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	handler := interactionHTTP.NewInteractionHandler(mockUC)
	router.DELETE("/interactions/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/interactions/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, resp, 200)
}

func TestInteractionHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInteractionUseCase{
		DeleteInteractionFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return interaction.ErrInteractionNotFound
		},
	}

	handler := interactionHTTP.NewInteractionHandler(mockUC)
	router.DELETE("/interactions/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/interactions/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, resp, 404, "NOT_FOUND")
}
