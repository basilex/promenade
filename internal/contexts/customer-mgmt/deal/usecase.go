package deal

import (
	"context"
	"errors"
	"time"

	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// IUseCase defines business operations for deals
type IUseCase interface {
	// CreateDeal creates a new deal
	CreateDeal(ctx context.Context, name string, customerID, assignedTo uuidv7.UUID, value int64, currency string, expectedCloseDate string) (*Deal, error)

	// GetDeal retrieves a deal by ID
	GetDeal(ctx context.Context, id uuidv7.UUID) (*Deal, error)

	// UpdateDealBasicInfo updates deal name and description
	UpdateDealBasicInfo(ctx context.Context, id uuidv7.UUID, name, description string) (*Deal, error)

	// UpdateDealValue updates deal value
	UpdateDealValue(ctx context.Context, id uuidv7.UUID, value int64, currency string) (*Deal, error)

	// MoveDealToStage moves deal to a new stage
	MoveDealToStage(ctx context.Context, id uuidv7.UUID, stage DealStage) (*Deal, error)

	// MarkDealAsWon marks deal as closed won
	MarkDealAsWon(ctx context.Context, id uuidv7.UUID, reason string) (*Deal, error)

	// MarkDealAsLost marks deal as closed lost
	MarkDealAsLost(ctx context.Context, id uuidv7.UUID, reason string) (*Deal, error)

	// UpdateDealProbability updates win probability
	UpdateDealProbability(ctx context.Context, id uuidv7.UUID, probability int) (*Deal, error)

	// UpdateDealExpectedCloseDate updates expected close date
	UpdateDealExpectedCloseDate(ctx context.Context, id uuidv7.UUID, date string) (*Deal, error)

	// AssignDealToSalesRep assigns deal to a sales rep
	AssignDealToSalesRep(ctx context.Context, id, userID uuidv7.UUID) (*Deal, error)

	// LinkDealToCompany links deal to a company
	LinkDealToCompany(ctx context.Context, id, companyID uuidv7.UUID) (*Deal, error)

	// SetDealSource sets deal source
	SetDealSource(ctx context.Context, id uuidv7.UUID, source DealSource) (*Deal, error)

	// DeleteDeal soft deletes a deal
	DeleteDeal(ctx context.Context, id uuidv7.UUID) error

	// ListDeals returns paginated deals
	ListDeals(ctx context.Context, page, pageSize int) ([]*Deal, int64, error)

	// ListDealsByStage returns deals in a specific stage
	ListDealsByStage(ctx context.Context, stage DealStage, page, pageSize int) ([]*Deal, int64, error)

	// ListDealsByCustomer returns deals for a specific customer
	ListDealsByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Deal, int64, error)

	// ListDealsByCompany returns deals for a specific company
	ListDealsByCompany(ctx context.Context, companyID uuidv7.UUID, page, pageSize int) ([]*Deal, int64, error)

	// ListDealsByAssignedTo returns deals assigned to a sales rep
	ListDealsByAssignedTo(ctx context.Context, userID uuidv7.UUID, page, pageSize int) ([]*Deal, int64, error)

	// ListDealsBySource returns deals from a specific source
	ListDealsBySource(ctx context.Context, source DealSource, page, pageSize int) ([]*Deal, int64, error)

	// GetPipelineStats returns deal counts by stage
	GetPipelineStats(ctx context.Context) (map[DealStage]int64, error)

	// GetTotalValue returns total value of all active deals
	GetTotalValue(ctx context.Context) (int64, error)

	// GetWonDeals returns won deals count and value
	GetWonDeals(ctx context.Context) (int64, int64, error)
}

// useCase implements IUseCase
type useCase struct {
	repo IRepository
}

// NewUseCase creates a new deal use case
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{
		repo: repo,
	}
}

// CreateDeal creates a new deal
func (uc *useCase) CreateDeal(ctx context.Context, name string, customerID, assignedTo uuidv7.UUID, value int64, currency string, expectedCloseDate string) (*Deal, error) {
	log := logger.FromContext(ctx)

	// Parse expected close date
	closeDate, err := parseDate(expectedCloseDate)
	if err != nil {
		return nil, err // parseDate returns descriptive error already
	}

	// Create Money value object
	money, err := valueobject.NewMoney(value, currency)
	if err != nil {
		return nil, err // NewMoney returns descriptive error already
	}

	// Create deal entity
	deal, err := NewDeal(customerID, name, money, assignedTo, closeDate)
	if err != nil {
		return nil, err // NewDeal returns domain errors (ErrDealNameEmpty, etc.)
	}

	// Persist deal
	if err := uc.repo.Create(ctx, deal); err != nil {
		log.Error("Failed to create deal", "error", err)
		return nil, ErrDealCreateFailed
	}

	log.Info("Deal created", "deal_id", deal.ID, "name", deal.Name)
	return deal, nil
}

// GetDeal retrieves a deal by ID
func (uc *useCase) GetDeal(ctx context.Context, id uuidv7.UUID) (*Deal, error) {
	deal, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrDealNotFound) {
			return nil, ErrDealNotFound
		}
		return nil, err
	}

	return deal, nil
}

// UpdateDealBasicInfo updates deal name and description
func (uc *useCase) UpdateDealBasicInfo(ctx context.Context, id uuidv7.UUID, name, description string) (*Deal, error) {
	log := logger.FromContext(ctx)

	deal, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrDealNotFound) {
			return nil, ErrDealNotFound
		}
		return nil, err
	}

	if err := deal.UpdateBasicInfo(name, description); err != nil {
		return nil, err // Entity returns ErrDealNameEmpty
	}

	if err := uc.repo.Update(ctx, deal); err != nil {
		log.Error("Failed to update deal", "error", err)
		return nil, ErrDealUpdateFailed
	}

	log.Info("Deal basic info updated", "deal_id", deal.ID)
	return deal, nil
}

// UpdateDealValue updates deal value
func (uc *useCase) UpdateDealValue(ctx context.Context, id uuidv7.UUID, value int64, currency string) (*Deal, error) {
	log := logger.FromContext(ctx)

	deal, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrDealNotFound) {
			return nil, ErrDealNotFound
		}
		return nil, err
	}

	// Create Money value object
	money, err := valueobject.NewMoney(value, currency)
	if err != nil {
		return nil, err // NewMoney returns descriptive error
	}

	if err := deal.UpdateValue(money); err != nil {
		return nil, err // Entity returns ErrDealValueNegative
	}

	if err := uc.repo.Update(ctx, deal); err != nil {
		log.Error("Failed to update deal", "error", err)
		return nil, ErrDealUpdateFailed
	}

	log.Info("Deal value updated", "deal_id", deal.ID, "value", value, "currency", currency)
	return deal, nil
}

// MoveDealToStage moves deal to a new stage
func (uc *useCase) MoveDealToStage(ctx context.Context, id uuidv7.UUID, stage DealStage) (*Deal, error) {
	log := logger.FromContext(ctx)

	deal, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrDealNotFound) {
			return nil, ErrDealNotFound
		}
		return nil, err
	}

	if err := deal.MoveTo(stage); err != nil {
		return nil, err // Entity returns ErrDealTerminalStage or ErrDealInvalidStageTransition
	}

	if err := uc.repo.Update(ctx, deal); err != nil {
		log.Error("Failed to update deal", "error", err)
		return nil, ErrDealUpdateFailed
	}

	log.Info("Deal stage moved", "deal_id", deal.ID, "stage", stage)
	return deal, nil
}

// MarkDealAsWon marks deal as closed won
func (uc *useCase) MarkDealAsWon(ctx context.Context, id uuidv7.UUID, reason string) (*Deal, error) {
	log := logger.FromContext(ctx)

	deal, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrDealNotFound) {
			return nil, ErrDealNotFound
		}
		return nil, err
	}

	if err := deal.MarkAsWon(reason); err != nil {
		return nil, err // Entity returns ErrDealAlreadyWon or ErrDealCannotMarkLostAsWon
	}

	if err := uc.repo.Update(ctx, deal); err != nil {
		log.Error("Failed to update deal", "error", err)
		return nil, ErrDealUpdateFailed
	}

	log.Info("Deal marked as won", "deal_id", deal.ID, "value", deal.Value.Amount)
	return deal, nil
}

// MarkDealAsLost marks deal as closed lost
func (uc *useCase) MarkDealAsLost(ctx context.Context, id uuidv7.UUID, reason string) (*Deal, error) {
	log := logger.FromContext(ctx)

	deal, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrDealNotFound) {
			return nil, ErrDealNotFound
		}
		return nil, err
	}

	if err := deal.MarkAsLost(reason); err != nil {
		return nil, err // Entity returns ErrDealAlreadyLost, ErrDealCannotMarkWonAsLost, or ErrDealLossReasonRequired
	}

	if err := uc.repo.Update(ctx, deal); err != nil {
		log.Error("Failed to update deal", "error", err)
		return nil, ErrDealUpdateFailed
	}

	log.Info("Deal marked as lost", "deal_id", deal.ID, "reason", reason)
	return deal, nil
}

// UpdateDealProbability updates win probability
func (uc *useCase) UpdateDealProbability(ctx context.Context, id uuidv7.UUID, probability int) (*Deal, error) {
	log := logger.FromContext(ctx)

	deal, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrDealNotFound) {
			return nil, ErrDealNotFound
		}
		return nil, err
	}

	// Entity returns ErrDealProbabilityRange or ErrDealClosedMutation
	if err := deal.UpdateProbability(probability); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, deal); err != nil {
		log.Error("Failed to update deal", "error", err)
		return nil, ErrDealUpdateFailed
	}

	log.Info("Deal probability updated", "deal_id", deal.ID, "probability", probability)
	return deal, nil
}

// UpdateDealExpectedCloseDate updates expected close date
func (uc *useCase) UpdateDealExpectedCloseDate(ctx context.Context, id uuidv7.UUID, date string) (*Deal, error) {
	log := logger.FromContext(ctx)

	deal, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrDealNotFound) {
			return nil, ErrDealNotFound
		}
		return nil, err
	}

	// parseDate returns descriptive error
	closeDate, err := parseDate(date)
	if err != nil {
		return nil, err
	}

	// Entity returns ErrDealDateInPast or ErrDealClosedMutation
	if err := deal.UpdateExpectedCloseDate(closeDate); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, deal); err != nil {
		log.Error("Failed to update deal", "error", err)
		return nil, ErrDealUpdateFailed
	}

	log.Info("Deal expected close date updated", "deal_id", deal.ID, "date", closeDate)
	return deal, nil
}

// AssignDealToSalesRep assigns deal to a sales rep
func (uc *useCase) AssignDealToSalesRep(ctx context.Context, id, userID uuidv7.UUID) (*Deal, error) {
	log := logger.FromContext(ctx)

	deal, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrDealNotFound) {
			return nil, ErrDealNotFound
		}
		return nil, err
	}

	// Entity returns ErrDealSalesRepRequired
	if err := deal.AssignToSalesRep(userID); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, deal); err != nil {
		log.Error("Failed to update deal", "error", err)
		return nil, ErrDealUpdateFailed
	}

	log.Info("Deal assigned to sales rep", "deal_id", deal.ID, "user_id", userID)
	return deal, nil
}

// LinkDealToCompany links deal to a company
func (uc *useCase) LinkDealToCompany(ctx context.Context, id, companyID uuidv7.UUID) (*Deal, error) {
	log := logger.FromContext(ctx)

	deal, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrDealNotFound) {
			return nil, ErrDealNotFound
		}
		return nil, err
	}

	deal.SetCompany(&companyID)

	if err := uc.repo.Update(ctx, deal); err != nil {
		log.Error("Failed to update deal", "error", err)
		return nil, ErrDealUpdateFailed
	}

	log.Info("Deal linked to company", "deal_id", deal.ID, "company_id", companyID)
	return deal, nil
}

// SetDealSource sets deal source
func (uc *useCase) SetDealSource(ctx context.Context, id uuidv7.UUID, source DealSource) (*Deal, error) {
	log := logger.FromContext(ctx)

	deal, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrDealNotFound) {
			return nil, ErrDealNotFound
		}
		return nil, err
	}

	deal.SetSource(source)

	if err := uc.repo.Update(ctx, deal); err != nil {
		log.Error("Failed to update deal", "error", err)
		return nil, ErrDealUpdateFailed
	}

	log.Info("Deal source set", "deal_id", deal.ID, "source", source)
	return deal, nil
}

// DeleteDeal soft deletes a deal
func (uc *useCase) DeleteDeal(ctx context.Context, id uuidv7.UUID) error {
	log := logger.FromContext(ctx)

	// Check if deal exists
	exists, err := uc.repo.Exists(ctx, id)
	if err != nil {
		return err
	}

	if !exists {
		return ErrDealNotFound
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		log.Error("Failed to delete deal", "error", err)
		return ErrDealDeleteFailed
	}

	log.Info("Deal deleted", "deal_id", id)
	return nil
}

// ListDeals returns paginated deals
func (uc *useCase) ListDeals(ctx context.Context, page, pageSize int) ([]*Deal, int64, error) {
	deals, total, err := uc.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, ErrDealListFailed
	}

	return deals, total, nil
}

// ListDealsByStage returns deals in a specific stage
func (uc *useCase) ListDealsByStage(ctx context.Context, stage DealStage, page, pageSize int) ([]*Deal, int64, error) {
	deals, total, err := uc.repo.ListByStage(ctx, stage, page, pageSize)
	if err != nil {
		return nil, 0, ErrDealListFailed
	}

	return deals, total, nil
}

// ListDealsByCustomer returns deals for a specific customer
func (uc *useCase) ListDealsByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Deal, int64, error) {
	deals, total, err := uc.repo.ListByCustomer(ctx, customerID, page, pageSize)
	if err != nil {
		return nil, 0, ErrDealListFailed
	}

	return deals, total, nil
}

// ListDealsByCompany returns deals for a specific company
func (uc *useCase) ListDealsByCompany(ctx context.Context, companyID uuidv7.UUID, page, pageSize int) ([]*Deal, int64, error) {
	deals, total, err := uc.repo.ListByCompany(ctx, companyID, page, pageSize)
	if err != nil {
		return nil, 0, ErrDealListFailed
	}

	return deals, total, nil
}

// ListDealsByAssignedTo returns deals assigned to a sales rep
func (uc *useCase) ListDealsByAssignedTo(ctx context.Context, userID uuidv7.UUID, page, pageSize int) ([]*Deal, int64, error) {
	deals, total, err := uc.repo.ListByAssignedTo(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, ErrDealListFailed
	}

	return deals, total, nil
}

// ListDealsBySource returns deals from a specific source
func (uc *useCase) ListDealsBySource(ctx context.Context, source DealSource, page, pageSize int) ([]*Deal, int64, error) {
	deals, total, err := uc.repo.ListBySource(ctx, source, page, pageSize)
	if err != nil {
		return nil, 0, ErrDealListFailed
	}

	return deals, total, nil
}

// GetPipelineStats returns deal counts by stage
func (uc *useCase) GetPipelineStats(ctx context.Context) (map[DealStage]int64, error) {
	stats, err := uc.repo.GetPipelineStats(ctx)
	if err != nil {
		return nil, ErrDealStatsFailed
	}

	return stats, nil
}

// GetTotalValue returns total value of all active deals
func (uc *useCase) GetTotalValue(ctx context.Context) (int64, error) {
	total, err := uc.repo.GetTotalValue(ctx)
	if err != nil {
		return 0, ErrDealStatsFailed
	}

	return total, nil
}

// GetWonDeals returns won deals count and value
func (uc *useCase) GetWonDeals(ctx context.Context) (int64, int64, error) {
	count, value, err := uc.repo.GetWonDeals(ctx)
	if err != nil {
		return 0, 0, ErrDealStatsFailed
	}

	return count, value, nil
}

// Helper functions

// parseDate parses date string to time.Time
func parseDate(dateStr string) (time.Time, error) {
	// Try multiple date formats
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05Z07:00",
		time.RFC3339,
	}

	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, ErrDateParseFailed
}
