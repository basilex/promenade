package taxcode_test

import (
	"context"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/jmoiron/sqlx"
)

// createTestGLAccount creates a test GL account for tax codes
func createTestGLAccount(ctx context.Context, tx *sqlx.Tx, orgID, userID uuidv7.UUID, code, name, accountType string) (uuidv7.UUID, error) {
	accountID := uuidv7.New()

	query := `
        INSERT INTO accounting_chart_of_accounts (
            id, version, organization_id, code, name, type,
            level, is_active, currency_code, created_by, last_updated_by,
            created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := tx.ExecContext(
		ctx, query,
		accountID.String(),
		1,
		orgID.String(),
		code,
		name,
		accountType,
		3,
		true,
		"UAH",
		userID.String(),
		userID.String(),
		time.Now(),
		time.Now(),
	)

	return accountID, err
}
