package usecase

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/banking/bankaccount"
	"github.com/basilex/promenade/internal/contexts/banking/bankaccount/aggregate"
	"github.com/basilex/promenade/internal/contexts/banking/bankaccount/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IBankAccountUseCase defines business operations for BankAccount aggregate
type IBankAccountUseCase interface {
	CreateManualAccount(ctx context.Context, organizationID uuidv7.UUID, name, bankName, currencyCode string, lastUpdatedBy uuidv7.UUID) (*aggregate.BankAccount, error)
	ConnectProviderAccount(ctx context.Context, organizationID uuidv7.UUID, name, bankName string, provider aggregate.BankProvider, providerAccountID string, lastUpdatedBy uuidv7.UUID) (*aggregate.BankAccount, error)
	GetAccount(ctx context.Context, id uuidv7.UUID) (*aggregate.BankAccount, error)
	GetAccountByProvider(ctx context.Context, provider aggregate.BankProvider, providerAccountID string) (*aggregate.BankAccount, error)
	ListAccounts(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.BankAccount, error)
	CountAccounts(ctx context.Context, organizationID uuidv7.UUID) (int, error)
	UpdateAccountDetails(ctx context.Context, id uuidv7.UUID, name, bankName, iban, accountNumber string, lastUpdatedBy uuidv7.UUID) error
	UpdateBalance(ctx context.Context, id uuidv7.UUID, balanceCents int64, lastUpdatedBy uuidv7.UUID) error
	RecordSync(ctx context.Context, id uuidv7.UUID, balanceCents int64, lastUpdatedBy uuidv7.UUID) error
	ActivateAccount(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	DeactivateAccount(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	ArchiveAccount(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	DeleteAccount(ctx context.Context, id uuidv7.UUID) error
}

// BankAccountUseCase implements IBankAccountUseCase
type BankAccountUseCase struct {
	repo repository.IBankAccountRepository
}

// NewBankAccountUseCase creates a new BankAccountUseCase
func NewBankAccountUseCase(repo repository.IBankAccountRepository) *BankAccountUseCase {
	return &BankAccountUseCase{repo: repo}
}

func (uc *BankAccountUseCase) CreateManualAccount(ctx context.Context, organizationID uuidv7.UUID, name, bankName, currencyCode string, lastUpdatedBy uuidv7.UUID) (*aggregate.BankAccount, error) {
	account, err := aggregate.NewBankAccount(organizationID, name, bankName, currencyCode, lastUpdatedBy)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (uc *BankAccountUseCase) ConnectProviderAccount(ctx context.Context, organizationID uuidv7.UUID, name, bankName string, provider aggregate.BankProvider, providerAccountID string, lastUpdatedBy uuidv7.UUID) (*aggregate.BankAccount, error) {
	existing, _ := uc.repo.GetByProviderAccountID(ctx, provider, providerAccountID)
	if existing != nil {
		return nil, bankaccount.ErrBankAccountNotFound
	}
	account, err := aggregate.NewBankAccount(organizationID, name, bankName, "UAH", lastUpdatedBy)
	if err != nil {
		return nil, err
	}
	if err := account.ConnectProvider(provider, providerAccountID, "", ""); err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (uc *BankAccountUseCase) GetAccount(ctx context.Context, id uuidv7.UUID) (*aggregate.BankAccount, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *BankAccountUseCase) GetAccountByProvider(ctx context.Context, provider aggregate.BankProvider, providerAccountID string) (*aggregate.BankAccount, error) {
	return uc.repo.GetByProviderAccountID(ctx, provider, providerAccountID)
}

func (uc *BankAccountUseCase) ListAccounts(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.BankAccount, error) {
	return uc.repo.List(ctx, organizationID, limit, offset)
}

func (uc *BankAccountUseCase) CountAccounts(ctx context.Context, organizationID uuidv7.UUID) (int, error) {
	return uc.repo.CountByOrganization(ctx, organizationID)
}

func (uc *BankAccountUseCase) UpdateAccountDetails(ctx context.Context, id uuidv7.UUID, name, bankName, iban, accountNumber string, lastUpdatedBy uuidv7.UUID) error {
	account, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if name != "" {
		account.Name = name
		account.Touch()
	}
	if bankName != "" {
		account.BankName = bankName
		account.Touch()
	}
	if iban != "" {
		account.IBAN = iban
		account.Touch()
	}
	if accountNumber != "" {
		account.AccountNumber = accountNumber
		account.Touch()
	}
	account.LastUpdatedBy = lastUpdatedBy
	return uc.repo.Update(ctx, account)
}

func (uc *BankAccountUseCase) UpdateBalance(ctx context.Context, id uuidv7.UUID, balanceCents int64, lastUpdatedBy uuidv7.UUID) error {
	account, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := account.UpdateBalance(balanceCents); err != nil {
		return err
	}
	account.LastUpdatedBy = lastUpdatedBy
	return uc.repo.Update(ctx, account)
}

func (uc *BankAccountUseCase) RecordSync(ctx context.Context, id uuidv7.UUID, balanceCents int64, lastUpdatedBy uuidv7.UUID) error {
	account, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := account.RecordSync(); err != nil {
		return err
	}
	account.LastUpdatedBy = lastUpdatedBy
	return uc.repo.Update(ctx, account)
}

func (uc *BankAccountUseCase) ActivateAccount(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	account, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := account.Activate(); err != nil {
		return err
	}
	account.LastUpdatedBy = lastUpdatedBy
	return uc.repo.Update(ctx, account)
}

func (uc *BankAccountUseCase) DeactivateAccount(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	account, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := account.Deactivate(); err != nil {
		return err
	}
	account.LastUpdatedBy = lastUpdatedBy
	return uc.repo.Update(ctx, account)
}

func (uc *BankAccountUseCase) ArchiveAccount(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	account, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := account.Archive(); err != nil {
		return err
	}
	account.LastUpdatedBy = lastUpdatedBy
	return uc.repo.Update(ctx, account)
}

func (uc *BankAccountUseCase) DeleteAccount(ctx context.Context, id uuidv7.UUID) error {
	return uc.repo.Delete(ctx, id)
}
