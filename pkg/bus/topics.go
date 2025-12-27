package bus

// Topic constants define standard event topics/channels.
// Using constants ensures consistency and prevents typos.
const (
	// Identity Context Topics
	TopicUserRegistered      = "user.registered"
	TopicUserActivated       = "user.activated"
	TopicUserSuspended       = "user.suspended"
	TopicUserPasswordChanged = "user.password.changed"
	TopicUserEmailVerified   = "user.email.verified"

	TopicContactCreated  = "contact.created"
	TopicContactUpdated  = "contact.updated"
	TopicContactVerified = "contact.verified"
	TopicContactDeleted  = "contact.deleted"

	// Shared Context Topics (Reference Data Management)
	TopicCountryCreated     = "country.created"
	TopicCountryUpdated     = "country.updated"
	TopicCountryDeactivated = "country.deactivated"

	TopicCurrencyCreated     = "currency.created"
	TopicCurrencyUpdated     = "currency.updated"
	TopicCurrencyDeactivated = "currency.deactivated"

	TopicLanguageCreated     = "language.created"
	TopicLanguageUpdated     = "language.updated"
	TopicLanguageDeactivated = "language.deactivated"

	TopicTimezoneCreated     = "timezone.created"
	TopicTimezoneUpdated     = "timezone.updated"
	TopicTimezoneDeactivated = "timezone.deactivated"

	// Customer Management Context Topics (future)
	TopicCustomerCreated = "customer.created"
	TopicCustomerUpdated = "customer.updated"
	TopicDealWon         = "deal.won"

	// Order Management Context Topics (future)
	TopicOrderCreated   = "order.created"
	TopicOrderConfirmed = "order.confirmed"
	TopicOrderShipped   = "order.shipped"

	// Billing Context Topics (future)
	TopicInvoiceGenerated = "invoice.generated"
	TopicPaymentReceived  = "payment.received"
	TopicPaymentFailed    = "payment.failed"

	// Notification Topics (cross-cutting)
	TopicNotificationEmail = "notification.email"
	TopicNotificationSMS   = "notification.sms"
)
