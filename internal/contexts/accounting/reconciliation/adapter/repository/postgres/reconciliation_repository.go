package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation"
	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation/repository"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type reconciliationRepository struct {
	db *sqlx.DB
}

// NewReconciliationRepository creates a new PostgreSQL bank reconciliation repository
func NewReconciliationRepository(db *sqlx.DB) repository.IReconciliationRepository {
	return &reconciliationRepository{db: db}
}

func (r *reconciliationRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

func (r *reconciliationRepository) Create(ctx context.Context, br *aggregate.Reconciliation) error {
	// Insert reconciliation
	query := `
        INSERT INTO accounting_bank_reconciliations (
            id, version, organization_id, bank_account_id, account_id,
            reconciliation_date, statement_date, bank_statement_balance_cents,
            book_balance_cents, status, outstanding_deposits_cents,
            outstanding_checks_cents, bank_fees_cents, interest_earned_cents,
            currency_code, reconciled_by, reconciled_at, approved_by, approved_at,
            notes, last_updated_by, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
        )`

	_, err := r.getExecutor(ctx).ExecContext(
		ctx, query,
		br.ID.String(),
		br.Version,
		br.OrganizationID.String(),
		br.BankAccountID.String(),
		br.AccountID.String(),
		br.ReconciliationDate.Format("2006-01-02"),
		br.StatementDate.Format("2006-01-02"),
		br.BankStatementBalanceCents,
		br.BookBalanceCents,
		br.Status,
		br.OutstandingDepositsCents,
		br.OutstandingChecksCents,
		br.BankFeesCents,
		br.InterestEarnedCents,
		br.CurrencyCode,
		uuidPtrToString(br.ReconciledBy),
		timePtrToString(br.ReconciledAt),
		uuidPtrToString(br.ApprovedBy),
		timePtrToString(br.ApprovedAt),
		br.Notes,
		br.LastUpdatedBy.String(),
		br.CreatedAt,
		br.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Insert reconciliation items
	if len(br.Items) > 0 {
		itemQuery := `
            INSERT INTO accounting_bank_reconciliation_items (
                id, reconciliation_id, transaction_type, transaction_id,
                transaction_date, description, amount_cents, is_matched,
                matched_at, notes, created_at
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

		for _, item := range br.Items {
			_, err := r.getExecutor(ctx).ExecContext(
				ctx, itemQuery,
				item.ID.String(),
				br.ID.String(),
				item.TransactionType,
				uuidPtrToString(item.TransactionID),
				item.TransactionDate,
				item.Description,
				item.AmountCents,
				item.IsMatched,
				timePtrToString(item.MatchedAt),
				item.Notes,
				item.CreatedAt,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *reconciliationRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Reconciliation, error) {
	// Get reconciliation
	query := `
        SELECT id, version, organization_id, bank_account_id, account_id,
               reconciliation_date, statement_date, bank_statement_balance_cents,
               book_balance_cents, status, outstanding_deposits_cents,
               outstanding_checks_cents, bank_fees_cents, interest_earned_cents,
               currency_code, reconciled_by, reconciled_at, approved_by, approved_at,
               notes, last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_bank_reconciliations
        WHERE id = $1 AND deleted_at IS NULL`

	var br aggregate.Reconciliation
	var orgID, bankAccountID, accountID, lastUpdatedBy string
	var reconciledBy, approvedBy sql.NullString
	var reconciledAt, approvedAt sql.NullTime

	err := r.getExecutor(ctx).QueryRowxContext(ctx, query, id.String()).Scan(
		&br.ID,
		&br.Version,
		&orgID,
		&bankAccountID,
		&accountID,
		&br.ReconciliationDate,
		&br.StatementDate,
		&br.BankStatementBalanceCents,
		&br.BookBalanceCents,
		&br.Status,
		&br.OutstandingDepositsCents,
		&br.OutstandingChecksCents,
		&br.BankFeesCents,
		&br.InterestEarnedCents,
		&br.CurrencyCode,
		&reconciledBy,
		&reconciledAt,
		&approvedBy,
		&approvedAt,
		&br.Notes,
		&lastUpdatedBy,
		&br.CreatedAt,
		&br.UpdatedAt,
		&br.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, reconciliation.ErrReconciliationNotFound
	}
	if err != nil {
		return nil, err
	}

	br.OrganizationID, _ = uuidv7.Parse(orgID)
	br.BankAccountID, _ = uuidv7.Parse(bankAccountID)
	br.AccountID, _ = uuidv7.Parse(accountID)
	br.LastUpdatedBy, _ = uuidv7.Parse(lastUpdatedBy)

	if reconciledBy.Valid {
		id, _ := uuidv7.Parse(reconciledBy.String)
		br.ReconciledBy = &id
	}
	if reconciledAt.Valid {
		br.ReconciledAt = &reconciledAt.Time
	}
	if approvedBy.Valid {
		id, _ := uuidv7.Parse(approvedBy.String)
		br.ApprovedBy = &id
	}
	if approvedAt.Valid {
		br.ApprovedAt = &approvedAt.Time
	}

	// Get reconciliation items
	items, err := r.getReconciliationItems(ctx, br.ID)
	if err != nil {
		return nil, err
	}
	br.Items = items

	return &br, nil
}

func (r *reconciliationRepository) Update(ctx context.Context, br *aggregate.Reconciliation) error {
	// Update reconciliation
	query := `
        UPDATE accounting_bank_reconciliations
        SET version = $2,
            bank_statement_balance_cents = $3,
            book_balance_cents = $4,
            status = $5,
            outstanding_deposits_cents = $6,
            outstanding_checks_cents = $7,
            bank_fees_cents = $8,
            interest_earned_cents = $9,
            reconciled_by = $10,
            reconciled_at = $11,
            approved_by = $12,
            approved_at = $13,
            notes = $14,
            last_updated_by = $15,
            updated_at = $16
        WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.getExecutor(ctx).ExecContext(
		ctx, query,
		br.ID.String(),
		br.Version,
		br.BankStatementBalanceCents,
		br.BookBalanceCents,
		br.Status,
		br.OutstandingDepositsCents,
		br.OutstandingChecksCents,
		br.BankFeesCents,
		br.InterestEarnedCents,
		uuidPtrToString(br.ReconciledBy),
		timePtrToString(br.ReconciledAt),
		uuidPtrToString(br.ApprovedBy),
		timePtrToString(br.ApprovedAt),
		br.Notes,
		br.LastUpdatedBy.String(),
		br.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return reconciliation.ErrReconciliationNotFound
	}

	// Delete existing items
	deleteQuery := `DELETE FROM accounting_bank_reconciliation_items WHERE reconciliation_id = $1`
	_, err = r.getExecutor(ctx).ExecContext(ctx, deleteQuery, br.ID.String())
	if err != nil {
		return err
	}

	// Insert updated items
	if len(br.Items) > 0 {
		itemQuery := `
            INSERT INTO accounting_bank_reconciliation_items (
                id, reconciliation_id, transaction_type, transaction_id,
                transaction_date, description, amount_cents, is_matched,
                matched_at, notes, created_at
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

		for _, item := range br.Items {
			_, err := r.getExecutor(ctx).ExecContext(
				ctx, itemQuery,
				item.ID.String(),
				br.ID.String(),
				item.TransactionType,
				uuidPtrToString(item.TransactionID),
				item.TransactionDate,
				item.Description,
				item.AmountCents,
				item.IsMatched,
				timePtrToString(item.MatchedAt),
				item.Notes,
				item.CreatedAt,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *reconciliationRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
        UPDATE accounting_bank_reconciliations
        SET deleted_at = $2
        WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.getExecutor(ctx).ExecContext(ctx, query, id.String(), time.Now())
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return reconciliation.ErrReconciliationNotFound
	}

	return nil
}

func (r *reconciliationRepository) ListByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.Reconciliation, error) {
	query := `
		SELECT r.id, r.version, r.organization_id, r.bank_account_id, r.account_id,
		       r.reconciliation_date, r.statement_date, r.bank_statement_balance_cents,
		       r.book_balance_cents, r.status, r.outstanding_deposits_cents,
		       r.outstanding_checks_cents, r.bank_fees_cents, r.interest_earned_cents,
		       r.currency_code, r.reconciled_by, r.reconciled_at, r.approved_by, r.approved_at,
		       r.notes, r.last_updated_by, r.created_at, r.updated_at,
		       ri.id as item_id, ri.transaction_type, ri.transaction_id, ri.transaction_date,
		       ri.description, ri.amount_cents, ri.is_matched, ri.matched_at, ri.notes as item_notes,
		       ri.created_at as item_created_at
		FROM accounting_bank_reconciliations r
		LEFT JOIN accounting_bank_reconciliation_items ri ON r.id = ri.reconciliation_id
		WHERE r.organization_id = $1 AND r.deleted_at IS NULL
		ORDER BY r.reconciliation_date DESC, r.id, ri.id
		LIMIT $2 OFFSET $3`

	rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, organizationID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanReconciliationsWithItems(rows)
}

func (r *reconciliationRepository) ListByBankAccount(ctx context.Context, bankAccountID uuidv7.UUID) ([]*aggregate.Reconciliation, error) {
	query := `
		SELECT r.id, r.version, r.organization_id, r.bank_account_id, r.account_id,
		       r.reconciliation_date, r.statement_date, r.bank_statement_balance_cents,
		       r.book_balance_cents, r.status, r.outstanding_deposits_cents,
		       r.outstanding_checks_cents, r.bank_fees_cents, r.interest_earned_cents,
		       r.currency_code, r.reconciled_by, r.reconciled_at, r.approved_by, r.approved_at,
		       r.notes, r.last_updated_by, r.created_at, r.updated_at,
		       ri.id as item_id, ri.transaction_type, ri.transaction_id, ri.transaction_date,
		       ri.description, ri.amount_cents, ri.is_matched, ri.matched_at, ri.notes as item_notes,
		       ri.created_at as item_created_at
		FROM accounting_bank_reconciliations r
		LEFT JOIN accounting_bank_reconciliation_items ri ON r.id = ri.reconciliation_id
		WHERE r.bank_account_id = $1 AND r.deleted_at IS NULL
		ORDER BY r.reconciliation_date DESC, r.id, ri.id`

	rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, bankAccountID.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanReconciliationsWithItems(rows)
}

func (r *reconciliationRepository) ListByStatus(ctx context.Context, organizationID uuidv7.UUID, status reconciliation.Status) ([]*aggregate.Reconciliation, error) {
	query := `
		SELECT r.id, r.version, r.organization_id, r.bank_account_id, r.account_id,
		       r.reconciliation_date, r.statement_date, r.bank_statement_balance_cents,
		       r.book_balance_cents, r.status, r.outstanding_deposits_cents,
		       r.outstanding_checks_cents, r.bank_fees_cents, r.interest_earned_cents,
		       r.currency_code, r.reconciled_by, r.reconciled_at, r.approved_by, r.approved_at,
		       r.notes, r.last_updated_by, r.created_at, r.updated_at,
		       ri.id as item_id, ri.transaction_type, ri.transaction_id, ri.transaction_date,
		       ri.description, ri.amount_cents, ri.is_matched, ri.matched_at, ri.notes as item_notes,
		       ri.created_at as item_created_at
		FROM accounting_bank_reconciliations r
		LEFT JOIN accounting_bank_reconciliation_items ri ON r.id = ri.reconciliation_id
		WHERE r.organization_id = $1 AND r.status = $2 AND r.deleted_at IS NULL
		ORDER BY r.reconciliation_date DESC, r.id, ri.id`

	rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, organizationID.String(), status)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanReconciliationsWithItems(rows)
}

func (r *reconciliationRepository) ListByDateRange(ctx context.Context, organizationID uuidv7.UUID, startDate, endDate string) ([]*aggregate.Reconciliation, error) {
	query := `
        SELECT id, version, organization_id, bank_account_id, account_id,
               reconciliation_date, statement_date, bank_statement_balance_cents,
               book_balance_cents, status, outstanding_deposits_cents,
               outstanding_checks_cents, bank_fees_cents, interest_earned_cents,
               currency_code, reconciled_by, reconciled_at, approved_by, approved_at,
               notes, last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_bank_reconciliations
        WHERE organization_id = $1 
          AND reconciliation_date >= $2 
          AND reconciliation_date <= $3 
          AND deleted_at IS NULL
        ORDER BY reconciliation_date DESC`

	rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, organizationID.String(), startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanReconciliations(ctx, rows)
}

func (r *reconciliationRepository) scanReconciliations(ctx context.Context, rows *sqlx.Rows) ([]*aggregate.Reconciliation, error) {
	var reconciliations []*aggregate.Reconciliation

	for rows.Next() {
		var br aggregate.Reconciliation
		var orgID, bankAccountID, accountID, lastUpdatedBy string
		var reconciledBy, approvedBy sql.NullString
		var reconciledAt, approvedAt sql.NullTime

		err := rows.Scan(
			&br.ID,
			&br.Version,
			&orgID,
			&bankAccountID,
			&accountID,
			&br.ReconciliationDate,
			&br.StatementDate,
			&br.BankStatementBalanceCents,
			&br.BookBalanceCents,
			&br.Status,
			&br.OutstandingDepositsCents,
			&br.OutstandingChecksCents,
			&br.BankFeesCents,
			&br.InterestEarnedCents,
			&br.CurrencyCode,
			&reconciledBy,
			&reconciledAt,
			&approvedBy,
			&approvedAt,
			&br.Notes,
			&lastUpdatedBy,
			&br.CreatedAt,
			&br.UpdatedAt,
			&br.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		br.OrganizationID, _ = uuidv7.Parse(orgID)
		br.BankAccountID, _ = uuidv7.Parse(bankAccountID)
		br.AccountID, _ = uuidv7.Parse(accountID)
		br.LastUpdatedBy, _ = uuidv7.Parse(lastUpdatedBy)

		if reconciledBy.Valid {
			id, _ := uuidv7.Parse(reconciledBy.String)
			br.ReconciledBy = &id
		}
		if reconciledAt.Valid {
			br.ReconciledAt = &reconciledAt.Time
		}
		if approvedBy.Valid {
			id, _ := uuidv7.Parse(approvedBy.String)
			br.ApprovedBy = &id
		}
		if approvedAt.Valid {
			br.ApprovedAt = &approvedAt.Time
		}

		// Get reconciliation items
		items, err := r.getReconciliationItems(ctx, br.ID)
		if err != nil {
			return nil, err
		}
		br.Items = items

		reconciliations = append(reconciliations, &br)
	}

	return reconciliations, nil
}

func (r *reconciliationRepository) getReconciliationItems(ctx context.Context, reconciliationID uuidv7.UUID) ([]aggregate.Item, error) {
	query := `
        SELECT id, transaction_type, transaction_id, transaction_date,
               description, amount_cents, is_matched, matched_at, notes, created_at
        FROM accounting_bank_reconciliation_items
        WHERE reconciliation_id = $1
        ORDER BY created_at`

	rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, reconciliationID.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []aggregate.Item

	for rows.Next() {
		var item aggregate.Item
		var transactionID sql.NullString
		var matchedAt sql.NullTime

		err := rows.Scan(
			&item.ID,
			&item.TransactionType,
			&transactionID,
			&item.TransactionDate,
			&item.Description,
			&item.AmountCents,
			&item.IsMatched,
			&matchedAt,
			&item.Notes,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if transactionID.Valid {
			id, _ := uuidv7.Parse(transactionID.String)
			item.TransactionID = &id
		}
		if matchedAt.Valid {
			item.MatchedAt = &matchedAt.Time
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *reconciliationRepository) scanReconciliationsWithItems(rows *sqlx.Rows) ([]*aggregate.Reconciliation, error) {
	reconMap := make(map[string]*aggregate.Reconciliation)
	var reconOrder []string

	for rows.Next() {
		var reconID, orgID, bankAccountID, accountID, lastUpdatedBy string
		var br aggregate.Reconciliation
		var reconciledBy, approvedBy sql.NullString
		var reconciledAt, approvedAt sql.NullTime

		// Item fields (nullable because of LEFT JOIN)
		var itemID, transactionID sql.NullString
		var transactionType, itemDescription, itemNotes sql.NullString
		var transactionDate, itemCreatedAt sql.NullTime
		var amountCents sql.NullInt64
		var isMatched sql.NullBool
		var matchedAt sql.NullTime

		err := rows.Scan(
			&reconID,
			&br.Version,
			&orgID,
			&bankAccountID,
			&accountID,
			&br.ReconciliationDate,
			&br.StatementDate,
			&br.BankStatementBalanceCents,
			&br.BookBalanceCents,
			&br.Status,
			&br.OutstandingDepositsCents,
			&br.OutstandingChecksCents,
			&br.BankFeesCents,
			&br.InterestEarnedCents,
			&br.CurrencyCode,
			&reconciledBy,
			&reconciledAt,
			&approvedBy,
			&approvedAt,
			&br.Notes,
			&lastUpdatedBy,
			&br.CreatedAt,
			&br.UpdatedAt,
			// Item fields
			&itemID,
			&transactionType,
			&transactionID,
			&transactionDate,
			&itemDescription,
			&amountCents,
			&isMatched,
			&matchedAt,
			&itemNotes,
			&itemCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Check if we've seen this reconciliation before
		recon, exists := reconMap[reconID]
		if !exists {
			// New reconciliation
			br.ID, _ = uuidv7.Parse(reconID)
			br.OrganizationID, _ = uuidv7.Parse(orgID)
			br.BankAccountID, _ = uuidv7.Parse(bankAccountID)
			br.AccountID, _ = uuidv7.Parse(accountID)
			br.LastUpdatedBy, _ = uuidv7.Parse(lastUpdatedBy)

			if reconciledBy.Valid {
				id, _ := uuidv7.Parse(reconciledBy.String)
				br.ReconciledBy = &id
			}
			if reconciledAt.Valid {
				br.ReconciledAt = &reconciledAt.Time
			}
			if approvedBy.Valid {
				id, _ := uuidv7.Parse(approvedBy.String)
				br.ApprovedBy = &id
			}
			if approvedAt.Valid {
				br.ApprovedAt = &approvedAt.Time
			}

			br.Items = []aggregate.Item{}
			reconMap[reconID] = &br
			reconOrder = append(reconOrder, reconID)
			recon = &br
		}

		// Add item if exists (LEFT JOIN might return NULL)
		if itemID.Valid {
			item := aggregate.Item{
				ID:              uuidv7.MustParse(itemID.String),
				TransactionType: reconciliation.TransactionType(transactionType.String),
				TransactionDate: transactionDate.Time,
				Description:     itemDescription.String,
				AmountCents:     amountCents.Int64,
				IsMatched:       isMatched.Bool,
				Notes:           itemNotes.String,
				CreatedAt:       itemCreatedAt.Time,
			}
			if transactionID.Valid {
				id, _ := uuidv7.Parse(transactionID.String)
				item.TransactionID = &id
			}
			if matchedAt.Valid {
				item.MatchedAt = &matchedAt.Time
			}
			recon.Items = append(recon.Items, item)
		}
	}

	// Convert map to ordered slice
	var reconciliations []*aggregate.Reconciliation
	for _, id := range reconOrder {
		reconciliations = append(reconciliations, reconMap[id])
	}

	return reconciliations, nil
}

func uuidPtrToString(id *uuidv7.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

func timePtrToString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}
