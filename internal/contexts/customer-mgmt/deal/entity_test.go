package deal

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

func TestNewDeal_Success(t *testing.T) {
	customerID := uuidv7.New()
	assignedTo := uuidv7.New()
	value, _ := valueobject.NewMoney(100000, "USD")
	expectedCloseDate := time.Now().AddDate(0, 1, 0)

	deal, err := NewDeal(customerID, "Enterprise License", value, assignedTo, expectedCloseDate)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, deal.ID)
	assert.Equal(t, "Enterprise License", deal.Name)
	assert.Equal(t, DealStageLead, deal.Stage)
	assert.Equal(t, 10, deal.Probability)
}

func TestNewDeal_EmptyName(t *testing.T) {
	customerID := uuidv7.New()
	assignedTo := uuidv7.New()
	value, _ := valueobject.NewMoney(100000, "USD")
	expectedCloseDate := time.Now().AddDate(0, 1, 0)

	deal, err := NewDeal(customerID, "", value, assignedTo, expectedCloseDate)

	assert.Error(t, err)
	assert.Nil(t, deal)
}

func TestDeal_MoveTo_Valid(t *testing.T) {
	deal := createTestDeal(t)
	deal.Stage = DealStageLead

	err := deal.MoveTo(DealStageQualified)

	assert.NoError(t, err)
	assert.Equal(t, DealStageQualified, deal.Stage)
}

func TestDeal_MarkAsWon(t *testing.T) {
	deal := createTestDeal(t)
	deal.Stage = DealStageNegotiation

	err := deal.MarkAsWon("Customer accepted")

	assert.NoError(t, err)
	assert.Equal(t, DealStageClosedWon, deal.Stage)
	assert.Equal(t, 100, deal.Probability)
	assert.NotNil(t, deal.ActualCloseDate)
}

func createTestDeal(t *testing.T) *Deal {
	t.Helper()
	customerID := uuidv7.New()
	assignedTo := uuidv7.New()
	value, _ := valueobject.NewMoney(100000, "USD")
	expectedCloseDate := time.Now().AddDate(0, 1, 0)
	deal, _ := NewDeal(customerID, "Test Deal", value, assignedTo, expectedCloseDate)
	return deal
}
