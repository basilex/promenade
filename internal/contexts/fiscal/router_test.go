package fiscal

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/fiscal/checkbox"
)

func TestNewRouter_WithPDFOnly(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	router := NewRouter(sqlxDB, nil, t.TempDir())

	require.NotNil(t, router)
	require.NotNil(t, router.cashRegisterHandler)
	require.NotNil(t, router.receiptHandler)
}

func TestNewRouter_WithCheckboxAndPDF(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	client := &checkbox.Client{}
	router := NewRouter(sqlxDB, client, t.TempDir())

	require.NotNil(t, router)
	require.NotNil(t, router.cashRegisterHandler)
	require.NotNil(t, router.receiptHandler)
}

func TestRouter_RegisterRoutes(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	router := NewRouter(sqlxDB, nil, "")

	ginRouter := gin.New()
	api := ginRouter.Group("/api/v1")
	router.RegisterRoutes(api)

	routes := ginRouter.Routes()
	require.NotEmpty(t, routes)

	paths := make(map[string]bool)
	for _, r := range routes {
		paths[r.Method+" "+r.Path] = true
	}

	require.True(t, paths["POST /api/v1/fiscal/cash-registers"])
	require.True(t, paths["GET /api/v1/fiscal/cash-registers/:id"])
	require.True(t, paths["GET /api/v1/fiscal/cash-registers"])
	require.True(t, paths["DELETE /api/v1/fiscal/cash-registers/:id"])
	require.True(t, paths["POST /api/v1/fiscal/cash-registers/:id/activate"])
	require.True(t, paths["POST /api/v1/fiscal/cash-registers/:id/deactivate"])

	require.True(t, paths["POST /api/v1/fiscal/receipts"])
	require.True(t, paths["GET /api/v1/fiscal/receipts/:id"])
	require.True(t, paths["GET /api/v1/fiscal/receipts"])
	require.True(t, paths["POST /api/v1/fiscal/receipts/:id/print"])
	require.True(t, paths["POST /api/v1/fiscal/receipts/:id/cancel"])
	require.True(t, paths["DELETE /api/v1/fiscal/receipts/:id"])
}
