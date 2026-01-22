package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/accounting/journalentry"
	"github.com/basilex/promenade/internal/contexts/accounting/journalentry/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/journalentry/repository"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type JournalEntryRepository struct {
	db *sqlx.DB
}

func NewJournalEntryRepository(db *sqlx.DB) repository.IJournalEntryRepository {
	return &JournalEntryRepository{db: db}
}

func (r *JournalEntryRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

type journalEntryRow struct {
	ID             string         `db:"id"`
	OrganizationID string         `db:"organization_id"`
	EntryDate      time.Time      `db:"entry_date"`
	Description    string         `db:"description"`
	Status         string         `db:"status"`
	SourceType     string         `db:"source_type"`
	SourceID       sql.NullString `db:"source_id"`
	PostedBy       sql.NullString `db:"posted_by"`
	PostedAt       sql.NullTime   `db:"posted_at"`
	ReversedBy     sql.NullString `db:"reversed_by"`
	ReversedAt     sql.NullTime   `db:"reversed_at"`
	CreatedBy      string         `db:"created_by"`
	LastUpdatedBy  string         `db:"last_updated_by"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
	DeletedAt      sql.NullTime   `db:"deleted_at"`
}

type journalLineRow struct {
	ID           string    `db:"id"`
	JournalID    string    `db:"journal_entry_id"`
	AccountID    string    `db:"account_id"`
	DebitCents   int64     `db:"debit_cents"`
	CreditCents  int64     `db:"credit_cents"`
	CurrencyCode string    `db:"currency_code"`
	Description  string    `db:"description"`
	LineOrder    int       `db:"line_order"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (r *journalEntryRow) toEntity(lines []*aggregate.JournalEntryLine) (*aggregate.JournalEntry, error) {
	id, err := uuidv7.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid journal entry ID: %w", err)
	}

	orgID, err := uuidv7.Parse(r.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	lastUpdatedBy, err := uuidv7.Parse(r.LastUpdatedBy)
	if err != nil {
		return nil, fmt.Errorf("invalid last_updated_by: %w", err)
	}

	entry := &aggregate.JournalEntry{
		OrganizationID: orgID,
		EntryDate:      r.EntryDate,
		Description:    r.Description,
		Status:         aggregate.EntryStatus(r.Status),
		Lines:          lines,
		SourceType:     aggregate.SourceType(r.SourceType),
		LastUpdatedBy:  lastUpdatedBy,
	}

	entry.ID = id
	entry.CreatedAt = r.CreatedAt
	entry.UpdatedAt = r.UpdatedAt

	if r.SourceID.Valid {
		sourceID, _ := uuidv7.Parse(r.SourceID.String)
		entry.SourceID = &sourceID
	}
	if r.PostedAt.Valid {
		entry.PostedAt = &r.PostedAt.Time
	}
	if r.PostedBy.Valid {
		postedBy, _ := uuidv7.Parse(r.PostedBy.String)
		entry.PostedBy = &postedBy
	}
	if r.ReversedAt.Valid {
		entry.ReversedAt = &r.ReversedAt.Time
	}
	if r.ReversedBy.Valid {
		reversedBy, _ := uuidv7.Parse(r.ReversedBy.String)
		entry.ReversedBy = &reversedBy
	}

	return entry, nil
}

func (r *JournalEntryRepository) Create(ctx context.Context, entry *aggregate.JournalEntry) error {
	entryQuery := "INSERT INTO accounting_journal_entries (id,organization_id,entry_date,description,status,source_type,source_id,created_by,last_updated_by,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)"

	var sourceID *string
	if entry.SourceID != nil {
		sid := entry.SourceID.String()
		sourceID = &sid
	}

	_, err := r.getExecutor(ctx).ExecContext(ctx, entryQuery,
		entry.ID.String(),
		entry.OrganizationID.String(),
		entry.EntryDate,
		entry.Description,
		string(entry.Status),
		string(entry.SourceType),
		sourceID,
		entry.LastUpdatedBy.String(),
		entry.LastUpdatedBy.String(),
		entry.CreatedAt,
		entry.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create journal entry: %w", err)
	}

	lineQuery := "INSERT INTO accounting_journal_entry_lines (id,journal_entry_id,account_id,debit_cents,credit_cents,currency_code,description,line_order,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)"
	for _, line := range entry.Lines {
		_, err := r.getExecutor(ctx).ExecContext(ctx, lineQuery,
			line.ID.String(),
			entry.ID.String(),
			line.AccountID.String(),
			line.DebitCents,
			line.CreditCents,
			line.CurrencyCode,
			line.Description,
			line.LineOrder,
			time.Now(),
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to create journal line: %w", err)
		}
	}

	return nil
}

func (r *JournalEntryRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.JournalEntry, error) {
	entryQuery := "SELECT id,organization_id,entry_date,description,status,source_type,source_id,posted_by,posted_at,reversed_by,reversed_at,created_by,last_updated_by,created_at,updated_at,deleted_at FROM accounting_journal_entries WHERE id=$1 AND deleted_at IS NULL"

	var row journalEntryRow
	err := sqlx.GetContext(ctx, r.getExecutor(ctx), &row, entryQuery, id.String())
	if err == sql.ErrNoRows {
		return nil, journalentry.ErrJournalEntryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get journal entry: %w", err)
	}

	lineQuery := "SELECT id,journal_entry_id,account_id,debit_cents,credit_cents,currency_code,description,line_order,created_at,updated_at FROM accounting_journal_entry_lines WHERE journal_entry_id=$1 ORDER BY line_order"
	var lineRows []journalLineRow
	err = sqlx.SelectContext(ctx, r.getExecutor(ctx), &lineRows, lineQuery, id.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get journal lines: %w", err)
	}

	lines := make([]*aggregate.JournalEntryLine, 0, len(lineRows))
	for _, lr := range lineRows {
		lineID, _ := uuidv7.Parse(lr.ID)
		accountID, _ := uuidv7.Parse(lr.AccountID)
		lines = append(lines, &aggregate.JournalEntryLine{
			ID:           lineID,
			AccountID:    accountID,
			DebitCents:   lr.DebitCents,
			CreditCents:  lr.CreditCents,
			CurrencyCode: lr.CurrencyCode,
			Description:  lr.Description,
			LineOrder:    lr.LineOrder,
		})
	}

	return row.toEntity(lines)
}

func (r *JournalEntryRepository) GetByDocumentNumber(ctx context.Context, orgID uuidv7.UUID, docNum string) (*aggregate.JournalEntry, error) {
	return nil, fmt.Errorf("GetByDocumentNumber not implemented - document_number field does not exist in aggregate")
}

func (r *JournalEntryRepository) Update(ctx context.Context, entry *aggregate.JournalEntry) error {
	query := "UPDATE accounting_journal_entries SET description=$1,status=$2,posted_at=$3,posted_by=$4,reversed_at=$5,reversed_by=$6,last_updated_by=$7,updated_at=$8 WHERE id=$9 AND deleted_at IS NULL"

	var postedAt *time.Time
	var postedBy *string
	var reversedAt *time.Time
	var reversedBy *string

	if entry.PostedAt != nil {
		postedAt = entry.PostedAt
	}
	if entry.PostedBy != nil {
		pb := entry.PostedBy.String()
		postedBy = &pb
	}
	if entry.ReversedAt != nil {
		reversedAt = entry.ReversedAt
	}
	if entry.ReversedBy != nil {
		rb := entry.ReversedBy.String()
		reversedBy = &rb
	}

	nilUUID := "00000000-0000-0000-0000-000000000000"

	res, err := r.getExecutor(ctx).ExecContext(ctx, query,
		entry.Description,
		string(entry.Status),
		postedAt,
		postedBy,
		reversedAt,
		reversedBy,
		nilUUID,
		entry.UpdatedAt,
		entry.ID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update journal entry: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return journalentry.ErrJournalEntryNotFound
	}

	return nil
}

func (r *JournalEntryRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := "UPDATE accounting_journal_entries SET deleted_at=$1 WHERE id=$2 AND deleted_at IS NULL"

	res, err := r.getExecutor(ctx).ExecContext(ctx, query, time.Now(), id.String())
	if err != nil {
		return fmt.Errorf("failed to delete journal entry: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return journalentry.ErrJournalEntryNotFound
	}

	return nil
}

func (r *JournalEntryRepository) ListByOrganization(ctx context.Context, orgID uuidv7.UUID, status aggregate.EntryStatus) ([]*aggregate.JournalEntry, error) {
	query := "SELECT id,organization_id,entry_date,description,status,source_type,source_id,posted_by,posted_at,reversed_by,reversed_at,created_by,last_updated_by,created_at,updated_at,deleted_at FROM accounting_journal_entries WHERE organization_id=$1"

	args := []interface{}{orgID.String()}
	if status != "" {
		query += " AND status=$2"
		args = append(args, string(status))
	}
	query += " AND deleted_at IS NULL ORDER BY entry_date DESC"

	var rows []journalEntryRow
	err := sqlx.SelectContext(ctx, r.getExecutor(ctx), &rows, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list journal entries: %w", err)
	}

	entries := make([]*aggregate.JournalEntry, 0, len(rows))
	for _, row := range rows {
		entryID, _ := uuidv7.Parse(row.ID)

		lineQuery := "SELECT id,journal_entry_id,account_id,debit_cents,credit_cents,currency_code,description,line_order,created_at,updated_at FROM accounting_journal_entry_lines WHERE journal_entry_id=$1 ORDER BY line_order"
		var lineRows []journalLineRow
		err = sqlx.SelectContext(ctx, r.getExecutor(ctx), &lineRows, lineQuery, entryID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get journal lines: %w", err)
		}

		lines := make([]*aggregate.JournalEntryLine, 0, len(lineRows))
		for _, lr := range lineRows {
			lineID, _ := uuidv7.Parse(lr.ID)
			accountID, _ := uuidv7.Parse(lr.AccountID)
			lines = append(lines, &aggregate.JournalEntryLine{
				ID:           lineID,
				AccountID:    accountID,
				DebitCents:   lr.DebitCents,
				CreditCents:  lr.CreditCents,
				CurrencyCode: lr.CurrencyCode,
				Description:  lr.Description,
				LineOrder:    lr.LineOrder,
			})
		}

		entry, err := row.toEntity(lines)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *JournalEntryRepository) ListByPeriod(ctx context.Context, orgID uuidv7.UUID, periodID uuidv7.UUID) ([]*aggregate.JournalEntry, error) {
	return nil, fmt.Errorf("ListByPeriod not implemented - fiscal_period_id field does not exist in aggregate")
}
