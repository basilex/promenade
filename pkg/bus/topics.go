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
	TopicOrderCancelled = "order.cancelled"
	TopicOrderFulfilled = "order.fulfilled"
	TopicOrderShipped   = "order.shipped"

	// Billing Context Topics (future)
	TopicInvoiceGenerated = "invoice.generated"
	TopicPaymentReceived  = "payment.received"
	TopicPaymentFailed    = "payment.failed"

	// Banking Context Topics
	BankAccountCreated             = "bank.account.created"
	BankAccountConnected           = "bank.account.connected"
	BankAccountBalanceUpdated      = "bank.account.balance.updated"
	BankAccountSynced              = "bank.account.synced"
	BankTransactionRecorded        = "bank.transaction.recorded"
	BankTransactionMatched         = "bank.transaction.matched"
	BankTransactionReconciled      = "bank.transaction.reconciled"
	TopicBankingStatementImported  = "banking.statement.imported"
	TopicBankingTransactionCreated = "banking.transaction.created"

	// Accounting Context Topics
	TopicAccountCreated         = "accounting.account.created"
	TopicAccountUpdated         = "accounting.account.updated"
	TopicAccountDeactivated     = "accounting.account.deactivated"
	TopicJournalEntryCreated    = "accounting.journal_entry.created"
	TopicJournalEntryPosted     = "accounting.journal_entry.posted"
	TopicJournalEntryReversed   = "accounting.journal_entry.reversed"
	TopicLedgerBalanceUpdated   = "accounting.ledger.balance.updated"
	TopicAccountingPeriodClosed = "accounting.period.closed"

	// Notification Topics (cross-cutting)
	TopicNotificationEmail = "notification.email"
	TopicNotificationSMS   = "notification.sms"
)
