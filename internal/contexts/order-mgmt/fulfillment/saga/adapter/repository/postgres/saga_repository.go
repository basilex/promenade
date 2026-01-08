package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/order-mgmt/fulfillment/saga"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type sagaRepository struct {
	*BaseRepository
}

func NewSagaRepository(db *sqlx.DB) saga.ISagaRepository {
	return &sagaRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

type sagaRow struct {
	ID              string         `db:"id"`
	OrderID         string         `db:"order_id"`
	CustomerID      string         `db:"customer_id"`
	State           string         `db:"state"`
	CurrentStep     int            `db:"current_step"`
	CompletedSteps  string         `db:"completed_steps"`
	FailedStep      sql.NullString `db:"failed_step"`
	PaymentID       sql.NullString `db:"payment_id"`
	ReservedItems   string         `db:"reserved_items"`
	ShipmentID      sql.NullString `db:"shipment_id"`
	TrackingNumber  sql.NullString `db:"tracking_number"`
	StartedAt       string         `db:"started_at"`
	CompletedAt     sql.NullString `db:"completed_at"`
	CancelledAt     sql.NullString `db:"cancelled_at"`
	FailureReason   sql.NullString `db:"failure_reason"`
	CreatedAt       string         `db:"created_at"`
	UpdatedAt       string         `db:"updated_at"`
}

func (r *sagaRepository) Save(ctx context.Context, s *saga.FulfillmentSaga) error {
	completedStepsJSON, err := json.Marshal(s.CompletedSteps.Get())
	if err != nil {
		return fmt.Errorf("failed to marshal completed_steps: %w", err)
	}

	reservedItemsJSON, err := json.Marshal(s.ReservedItems.Get())
	if err != nil {
		return fmt.Errorf("failed to marshal reserved_items: %w", err)
	}

	query := `
		INSERT INTO order_fulfillment_sagas (
			id, order_id, customer_id, state, current_step, completed_steps,
			failed_step, payment_id, reserved_items, shipment_id, tracking_number,
			started_at, completed_at, cancelled_at, failure_reason,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	args := []interface{}{
		s.ID.String(),
		s.OrderID.String(),
		s.CustomerID.String(),
		s.State,
		s.CurrentStep,
		completedStepsJSON,
		stringPtr(s.FailedStep),
		uuidToString(s.PaymentID),
		reservedItemsJSON,
		uuidToString(s.ShipmentID),
		stringPtr(s.TrackingNumber),
		s.StartedAt,
		timeToString(s.CompletedAt),
		timeToString(s.CancelledAt),
		stringPtr(s.FailureReason),
		s.CreatedAt,
		s.UpdatedAt,
	}

	_, err = r.Exec(ctx, r.Rebind(query), args...)
	return err
}

func (r *sagaRepository) Update(ctx context.Context, s *saga.FulfillmentSaga) error {
	completedStepsJSON, err := json.Marshal(s.CompletedSteps.Get())
	if err != nil {
		return fmt.Errorf("failed to marshal completed_steps: %w", err)
	}

	reservedItemsJSON, err := json.Marshal(s.ReservedItems.Get())
	if err != nil {
		return fmt.Errorf("failed to marshal reserved_items: %w", err)
	}

	query := `
		UPDATE order_fulfillment_sagas
		SET state = ?, current_step = ?, completed_steps = ?,
		    failed_step = ?, payment_id = ?, reserved_items = ?,
		    shipment_id = ?, tracking_number = ?,
		    completed_at = ?, cancelled_at = ?, failure_reason = ?,
		    updated_at = ?
		WHERE id = ?
	`

	args := []interface{}{
		s.State,
		s.CurrentStep,
		completedStepsJSON,
		stringPtr(s.FailedStep),
		uuidToString(s.PaymentID),
		reservedItemsJSON,
		uuidToString(s.ShipmentID),
		stringPtr(s.TrackingNumber),
		timeToString(s.CompletedAt),
		timeToString(s.CancelledAt),
		stringPtr(s.FailureReason),
		s.UpdatedAt,
		s.ID.String(),
	}

	_, err = r.Exec(ctx, r.Rebind(query), args...)
	return err
}

func (r *sagaRepository) FindByID(ctx context.Context, id uuidv7.UUID) (*saga.FulfillmentSaga, error) {
	var row sagaRow
	query := `
		SELECT id, order_id, customer_id, state, current_step, completed_steps,
		       failed_step, payment_id, reserved_items, shipment_id, tracking_number,
		       started_at, completed_at, cancelled_at, failure_reason,
		       created_at, updated_at
		FROM order_fulfillment_sagas
		WHERE id = ?
	`

	if err := r.Get(ctx, &row, r.Rebind(query), id.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return rowToSaga(&row)
}

func (r *sagaRepository) FindByOrderID(ctx context.Context, orderID uuidv7.UUID) (*saga.FulfillmentSaga, error) {
	var row sagaRow
	query := `
		SELECT id, order_id, customer_id, state, current_step, completed_steps,
		       failed_step, payment_id, reserved_items, shipment_id, tracking_number,
		       started_at, completed_at, cancelled_at, failure_reason,
		       created_at, updated_at
		FROM order_fulfillment_sagas
		WHERE order_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	if err := r.Get(ctx, &row, r.Rebind(query), orderID.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return rowToSaga(&row)
}

func (r *sagaRepository) FindInProgressSagas(ctx context.Context) ([]*saga.FulfillmentSaga, error) {
	var rows []sagaRow
	query := `
		SELECT id, order_id, customer_id, state, current_step, completed_steps,
		       failed_step, payment_id, reserved_items, shipment_id, tracking_number,
		       started_at, completed_at, cancelled_at, failure_reason,
		       created_at, updated_at
		FROM order_fulfillment_sagas
		WHERE state IN ('payment_processing', 'inventory_processing', 'shipping_processing', 'compensating')
		ORDER BY created_at ASC
	`

	if err := r.Select(ctx, &rows, r.Rebind(query)); err != nil {
		return nil, err
	}

	sagas := make([]*saga.FulfillmentSaga, 0, len(rows))
	for i := range rows {
		s, err := rowToSaga(&rows[i])
		if err != nil {
			return nil, err
		}
		sagas = append(sagas, s)
	}

	return sagas, nil
}

func (r *sagaRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM order_fulfillment_sagas WHERE id = ?`
	_, err := r.Exec(ctx, r.Rebind(query), id.String())
	return err
}

func rowToSaga(row *sagaRow) (*saga.FulfillmentSaga, error) {
	id, err := uuidv7.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse id: %w", err)
	}

	orderID, err := uuidv7.Parse(row.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse order_id: %w", err)
	}

	customerID, err := uuidv7.Parse(row.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse customer_id: %w", err)
	}

	var completedSteps []string
	if err := json.Unmarshal([]byte(row.CompletedSteps), &completedSteps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal completed_steps: %w", err)
	}

	var reservedItems []saga.ReservedItem
	if err := json.Unmarshal([]byte(row.ReservedItems), &reservedItems); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reserved_items: %w", err)
	}

	s := &saga.FulfillmentSaga{
		ID:            id,
		OrderID:       orderID,
		CustomerID:    customerID,
		State:         saga.FulfillmentSagaState(row.State),
		CurrentStep:   row.CurrentStep,
		StartedAt:     parseTime(row.StartedAt),
		CreatedAt:     parseTime(row.CreatedAt),
		UpdatedAt:     parseTime(row.UpdatedAt),
	}

	s.CompletedSteps.Set(completedSteps)
	s.ReservedItems.Set(reservedItems)

	if row.FailedStep.Valid {
		s.FailedStep = &row.FailedStep.String
	}

	if row.PaymentID.Valid {
		paymentID, err := uuidv7.Parse(row.PaymentID.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse payment_id: %w", err)
		}
		s.PaymentID = &paymentID
	}

	if row.ShipmentID.Valid {
		shipmentID, err := uuidv7.Parse(row.ShipmentID.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse shipment_id: %w", err)
		}
		s.ShipmentID = &shipmentID
	}

	if row.TrackingNumber.Valid {
		s.TrackingNumber = &row.TrackingNumber.String
	}

	if row.CompletedAt.Valid {
		completedAt := parseTime(row.CompletedAt.String)
		s.CompletedAt = &completedAt
	}

	if row.CancelledAt.Valid {
		cancelledAt := parseTime(row.CancelledAt.String)
		s.CancelledAt = &cancelledAt
	}

	if row.FailureReason.Valid {
		s.FailureReason = &row.FailureReason.String
	}

	return s, nil
}

// Helper functions

func stringPtr(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func uuidToString(u *uuidv7.UUID) interface{} {
	if u == nil {
		return nil
	}
	return u.String()
}

func timeToString(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
