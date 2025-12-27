package bus_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/pkg/bus"
)

func TestTopicConstants_AllExist(t *testing.T) {
	// User topics
	assert.NotEmpty(t, bus.TopicUserRegistered)
	assert.NotEmpty(t, bus.TopicUserActivated)
	assert.NotEmpty(t, bus.TopicUserSuspended)
	assert.NotEmpty(t, bus.TopicUserPasswordChanged)
	assert.NotEmpty(t, bus.TopicUserEmailVerified)

	// Contact topics
	assert.NotEmpty(t, bus.TopicContactCreated)
	assert.NotEmpty(t, bus.TopicContactUpdated)
	assert.NotEmpty(t, bus.TopicContactVerified)
	assert.NotEmpty(t, bus.TopicContactDeleted)

	// Country topics
	assert.NotEmpty(t, bus.TopicCountryCreated)
	assert.NotEmpty(t, bus.TopicCountryUpdated)
	assert.NotEmpty(t, bus.TopicCountryDeactivated)

	// Currency topics
	assert.NotEmpty(t, bus.TopicCurrencyCreated)
	assert.NotEmpty(t, bus.TopicCurrencyUpdated)
	assert.NotEmpty(t, bus.TopicCurrencyDeactivated)

	// Customer topics
	assert.NotEmpty(t, bus.TopicCustomerCreated)
	assert.NotEmpty(t, bus.TopicCustomerUpdated)
	assert.NotEmpty(t, bus.TopicDealWon)

	// Order topics
	assert.NotEmpty(t, bus.TopicOrderCreated)
	assert.NotEmpty(t, bus.TopicOrderConfirmed)
	assert.NotEmpty(t, bus.TopicOrderShipped)

	// Billing topics
	assert.NotEmpty(t, bus.TopicInvoiceGenerated)
	assert.NotEmpty(t, bus.TopicPaymentReceived)
	assert.NotEmpty(t, bus.TopicPaymentFailed)

	// Notification topics
	assert.NotEmpty(t, bus.TopicNotificationEmail)
	assert.NotEmpty(t, bus.TopicNotificationSMS)
}

func TestTopicConstants_Uniqueness(t *testing.T) {
	topics := []string{
		bus.TopicUserRegistered,
		bus.TopicUserActivated,
		bus.TopicUserSuspended,
		bus.TopicUserPasswordChanged,
		bus.TopicUserEmailVerified,
		bus.TopicContactCreated,
		bus.TopicContactUpdated,
		bus.TopicContactVerified,
		bus.TopicContactDeleted,
		bus.TopicCountryCreated,
		bus.TopicCountryUpdated,
		bus.TopicCountryDeactivated,
		bus.TopicCurrencyCreated,
		bus.TopicCurrencyUpdated,
		bus.TopicCurrencyDeactivated,
		bus.TopicLanguageCreated,
		bus.TopicLanguageUpdated,
		bus.TopicLanguageDeactivated,
		bus.TopicTimezoneCreated,
		bus.TopicTimezoneUpdated,
		bus.TopicTimezoneDeactivated,
		bus.TopicCustomerCreated,
		bus.TopicCustomerUpdated,
		bus.TopicDealWon,
		bus.TopicOrderCreated,
		bus.TopicOrderConfirmed,
		bus.TopicOrderShipped,
		bus.TopicInvoiceGenerated,
		bus.TopicPaymentReceived,
		bus.TopicPaymentFailed,
		bus.TopicNotificationEmail,
		bus.TopicNotificationSMS,
	}

	seen := make(map[string]bool)
	for _, topic := range topics {
		assert.False(t, seen[topic], "Duplicate topic: %s", topic)
		seen[topic] = true
	}
}
