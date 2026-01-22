package aggregate

import (
	"github.com/basilex/promenade/internal/contexts/accounting/taxcode"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type TaxType string

const (
	TaxTypeVAT         TaxType = "vat"
	TaxTypeIncomeTax   TaxType = "income_tax"
	TaxTypePayrollTax  TaxType = "payroll_tax"
	TaxTypeWithholding TaxType = "withholding"
	TaxTypeExcise      TaxType = "excise"
	TaxTypeCustoms     TaxType = "customs"
	TaxTypeProperty    TaxType = "property"
	TaxTypeOther       TaxType = "other"
)

type TaxCode struct {
	aggregate.BaseAggregate

	OrganizationID         uuidv7.UUID
	Code                   string
	Name                   string
	TaxType                TaxType
	Rate                   int
	TaxPayableAccountID    *uuidv7.UUID
	TaxReceivableAccountID *uuidv7.UUID
	IsActive               bool
	LastUpdatedBy          uuidv7.UUID
}

func NewTaxCode(organizationID uuidv7.UUID, code, name string, taxType TaxType, rate int, createdBy uuidv7.UUID) (*TaxCode, error) {
	if code == "" {
		return nil, taxcode.ErrTaxCodeEmpty
	}
	if name == "" {
		return nil, taxcode.ErrTaxNameEmpty
	}
	if !isValidTaxType(taxType) {
		return nil, taxcode.ErrInvalidTaxType
	}
	if rate < 0 || rate > 10000 {
		return nil, taxcode.ErrInvalidTaxRate
	}

	return &TaxCode{
		BaseAggregate:  aggregate.NewBaseAggregate(),
		OrganizationID: organizationID,
		Code:           code,
		Name:           name,
		TaxType:        taxType,
		Rate:           rate,
		IsActive:       true,
		LastUpdatedBy:  createdBy,
	}, nil
}

func (t *TaxCode) CalculateTax(taxableAmount int64) int64 {
	return (taxableAmount * int64(t.Rate)) / 10000
}

func (t *TaxCode) CalculateTaxableBase(grossAmount int64) int64 {
	return (grossAmount * 10000) / (10000 + int64(t.Rate))
}

func (t *TaxCode) CalculateTaxFromGross(grossAmount int64) int64 {
	taxableBase := t.CalculateTaxableBase(grossAmount)
	return t.CalculateTax(taxableBase)
}

func (t *TaxCode) UpdateRate(newRate int) error {
	if newRate < 0 || newRate > 10000 {
		return taxcode.ErrInvalidTaxRate
	}

	t.Rate = newRate
	t.Touch()
	return nil
}

func (t *TaxCode) SetTaxPayableAccount(accountID uuidv7.UUID) error {
	if accountID == uuidv7.Nil {
		return taxcode.ErrGLAccountRequired
	}

	t.TaxPayableAccountID = &accountID
	t.Touch()
	return nil
}

func (t *TaxCode) SetTaxReceivableAccount(accountID uuidv7.UUID) error {
	if accountID == uuidv7.Nil {
		return taxcode.ErrGLAccountRequired
	}

	t.TaxReceivableAccountID = &accountID
	t.Touch()
	return nil
}

func (t *TaxCode) Activate() {
	t.IsActive = true
	t.Touch()
}

func (t *TaxCode) Deactivate() {
	t.IsActive = false
	t.Touch()
}

func isValidTaxType(tt TaxType) bool {
	validTypes := []TaxType{
		TaxTypeVAT, TaxTypeIncomeTax, TaxTypePayrollTax,
		TaxTypeWithholding, TaxTypeExcise, TaxTypeCustoms,
		TaxTypeProperty, TaxTypeOther,
	}
	for _, vt := range validTypes {
		if tt == vt {
			return true
		}
	}
	return false
}
