package saga

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// ISagaRepository defines the interface for fulfillment saga persistence
type ISagaRepository interface {
	// Save creates a new fulfillment saga
	Save(ctx context.Context, saga *FulfillmentSaga) error

	// Update updates an existing fulfillment saga
	Update(ctx context.Context, saga *FulfillmentSaga) error

	// FindByID retrieves a saga by its ID
	FindByID(ctx context.Context, id uuidv7.UUID) (*FulfillmentSaga, error)

	// FindByOrderID retrieves a saga by order ID
	FindByOrderID(ctx context.Context, orderID uuidv7.UUID) (*FulfillmentSaga, error)

	// FindInProgressSagas retrieves all sagas currently in progress
	FindInProgressSagas(ctx context.Context) ([]*FulfillmentSaga, error)

	// Delete soft deletes a saga (if implementing soft delete pattern)
	Delete(ctx context.Context, id uuidv7.UUID) error
}
