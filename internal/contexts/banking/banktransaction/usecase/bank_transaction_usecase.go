package usecase

import (
	"context"
	"time"

	"github.com/basilex/promenade/internal/contexts/banking/banktransaction"
	"github.com/basilex/promenade/internal/contexts/banking/banktransaction/aggregate"
	"github.com/basilex/promenade/internal/contexts/banking/banktransaction/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IBankTransactionUseCase defines business operations for BankTransaction aggregate
type IBankTransactionUseCase interface {
	RecordTransaction(ctx context.Context, accountID uuidv7.UUID, direction aggregate.TransactionDirection, amountCents int64, currencyCode string, transactionAt time.Time, description string, lastUpdatedBy uuidv7.UUID) (*aggregate.BankTransaction, error)
	RecordTransactionWithExternalID(ctx context.Context, accountID uuidv7.UUID, externalID string, direction aggregate.TransactionDirection, amountCents int64, currencyCode string, transactionAt time.Time, description string, lastUpdatedBy uuidv7.UUID) (*aggregate.BankTransaction, error)
	GetTransaction(ctx context.Context, id uuidv7.UUID) (*aggregate.BankTransaction, error)
	GetTransactionByExternalID(ctx context.Context, accountID uuidv7.UUID, externalID string) (*aggregate.BankTransaction, error)
	ListTransactionsByAccount(ctx context.Context, accountID uuidv7.UUID, limit, offset int) ([]*aggregate.BankTransaction, error)
	ListUnmatchedTransactions(ctx context.Context, accountID uuidv7.UUID, limit, offset int) ([]*aggregate.BankTransaction, error)
	CountTransactionsByAccount(ctx context.Context, accountID uuidv7.UUID) (int, error)
	BookTransaction(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	CancelTransaction(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	MatchToInvoice(ctx context.Context, transactionID, invoiceID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	MatchToOrder(ctx context.Context, transactionID, orderID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	MatchToPayment(ctx context.Context, transactionID, paymentID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	UnmatchTransaction(ctx context.Context, transactionID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	SetCounterparty(ctx context.Context, transactionID uuidv7.UUID, name, iban string, lastUpdatedBy uuidv7.UUID) error
	DeleteTransaction(ctx context.Context, id uuidv7.UUID) error
}

// BankTransactionUseCase implements IBankTransactionUseCase
type BankTransactionUseCase struct {
	repo repository.IBankTransactionRepository
}

// NewBankTransactionUseCase creates a new BankTransactionUseCase
func NewBankTransactionUseCase(repo repository.IBankTransactionRepository) *BankTransactionUseCase {
	return &BankTransactionUseCase{repo: repo}
}

func (uc *BankTransactionUseCase) RecordTransaction(ctx context.Context, accountID uuidv7.UUID, direction aggregate.TransactionDirection, amountCents int64, currencyCode string, transactionAt time.Time, description string, lastUpdatedBy uuidv7.UUID) (*aggregate.BankTransaction, error) {
	tx, err := aggregate.NewBankTransaction(accountID, direction, amountCents, currencyCode, transactionAt, description, lastUpdatedBy)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (uc *BankTransactionUseCase) RecordTransactionWithExternalID(ctx context.Context, accountID uuidv7.UUID, externalID string, direction aggregate.TransactionDirection, amountCents int64, currencyCode string, transactionAt time.Time, description string, lastUpdatedBy uuidv7.UUID) (*aggregate.BankTransaction, error) {
	existing, _ := uc.repo.GetByExternalID(ctx, accountID, externalID)
	if existing != nil {
		return nil, banktransaction.ErrExternalIDExists
	}
	tx, err := aggregate.NewBankTransaction(accountID, direction, amountCents, currencyCode, transactionAt, description, lastUpdatedBy)
	if err != nil {
		return nil, err
	}
	tx.SetExternalID(externalID)
	if err := uc.repo.Create(ctx, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (uc *BankTransactionUseCase) GetTransaction(ctx context.Context, id uuidv7.UUID) (*aggregate.BankTransaction, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *BankTransactionUseCase) GetTransactionByExternalID(ctx context.Context, accountID uuidv7.UUID, externalID string) (*aggregate.BankTransaction, error) {
	return uc.repo.GetByExternalID(ctx, accountID, externalID)
}

func (uc *BankTransactionUseCase) ListTransactionsByAccount(ctx context.Context, accountID uuidv7.UUID, limit, offset int) ([]*aggregate.BankTransaction, error) {
	return uc.repo.ListByAccount(ctx, accountID, limit, offset)
}

func (uc *BankTransactionUseCase) ListUnmatchedTransactions(ctx context.Context, accountID uuidv7.UUID, limit, offset int) ([]*aggregate.BankTransaction, error) {
	return uc.repo.ListUnmatched(ctx, accountID, limit, offset)
}

func (uc *BankTransactionUseCase) CountTransactionsByAccount(ctx context.Context, accountID uuidv7.UUID) (int, error) {
	return uc.repo.CountByAccount(ctx, accountID)
}

func (uc *BankTransactionUseCase) BookTransaction(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	tx, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := tx.Book(); err != nil {
		return err
	}
	tx.LastUpdatedBy = lastUpdatedBy
	return uc.repo.Update(ctx, tx)
}

func (uc *BankTransactionUseCase) CancelTransaction(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	tx, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := tx.Cancel(); err != nil {
		return err
	}
	tx.LastUpdatedBy = lastUpdatedBy
	return uc.repo.Update(ctx, tx)
}

func (uc *BankTransactionUseCase) MatchToInvoice(ctx context.Context, transactionID, invoiceID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	tx, err := uc.repo.GetByID(ctx, transactionID)
	if err != nil {
		return err
	}
	if err := tx.Match(aggregate.MatchedEntityInvoice, invoiceID); err != nil {
		return err
	}
	tx.LastUpdatedBy = lastUpdatedBy
	return uc.repo.Update(ctx, tx)
}

func (uc *BankTransactionUseCase) MatchToOrder(ctx context.Context, transactionID, orderID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	tx, err := uc.repo.GetByID(ctx, transactionID)
	if err != nil {
		return err
	}
	if err := tx.Match(aggregate.MatchedEntityOrder, orderID); err != nil {
		return err
	}
	tx.LastUpdatedBy = lastUpdatedBy
	return uc.repo.Update(ctx, tx)
}

func (uc *BankTransactionUseCase) MatchToPayment(ctx context.Context, transactionID, paymentID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	tx, err := uc.repo.GetByID(ctx, transactionID)
	if err != nil {
		return err
	}
	if err := tx.Match(aggregate.MatchedEntityPayment, paymentID); err != nil {
		return err
	}
	tx.LastUpdatedBy = lastUpdatedBy
	return uc.repo.Update(ctx, tx)
}

func (uc *BankTransactionUseCase) UnmatchTransaction(ctx context.Context, transactionID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	tx, err := uc.repo.GetByID(ctx, transactionID)
	if err != nil {
		return err
	}
	if err := tx.Unmatch(); err != nil {
		return err
	}
	tx.LastUpdatedBy = lastUpdatedBy
	return uc.repo.Update(ctx, tx)
}

func (uc *BankTransactionUseCase) SetCounterparty(ctx context.Context, transactionID uuidv7.UUID, name, iban string, lastUpdatedBy uuidv7.UUID) error {
	tx, err := uc.repo.GetByID(ctx, transactionID)
	if err != nil {
		return err
	}
	tx.SetCounterparty(name, iban)
	tx.LastUpdatedBy = lastUpdatedBy
	return uc.repo.Update(ctx, tx)
}

func (uc *BankTransactionUseCase) DeleteTransaction(ctx context.Context, id uuidv7.UUID) error {
	return uc.repo.Delete(ctx, id)
}
