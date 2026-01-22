package accounting

import (
	"github.com/gin-gonic/gin"

	accounthttp "github.com/basilex/promenade/internal/contexts/accounting/account/adapter/http"
	accountusecase "github.com/basilex/promenade/internal/contexts/accounting/account/usecase"
	budgethttp "github.com/basilex/promenade/internal/contexts/accounting/budget/adapter/http"
	budgetusecase "github.com/basilex/promenade/internal/contexts/accounting/budget/usecase"
	costcenterhttp "github.com/basilex/promenade/internal/contexts/accounting/costcenter/adapter/http"
	costcenterusecase "github.com/basilex/promenade/internal/contexts/accounting/costcenter/usecase"
	fiscalperiodhttp "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/adapter/http"
	fiscalperiodusecase "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/usecase"
	journalentryhttp "github.com/basilex/promenade/internal/contexts/accounting/journalentry/adapter/http"
	journalentryusecase "github.com/basilex/promenade/internal/contexts/accounting/journalentry/usecase"
	reconciliationhttp "github.com/basilex/promenade/internal/contexts/accounting/reconciliation/adapter/http"
	reconciliationusecase "github.com/basilex/promenade/internal/contexts/accounting/reconciliation/usecase"
	taxcodehttp "github.com/basilex/promenade/internal/contexts/accounting/taxcode/adapter/http"
	taxcodeusecase "github.com/basilex/promenade/internal/contexts/accounting/taxcode/usecase"
)

// RegisterRoutes registers all accounting context routes
func RegisterRoutes(
	r *gin.RouterGroup,
	accountUC accountusecase.IAccountUseCase,
	journalEntryUC journalentryusecase.IJournalEntryUseCase,
	fiscalPeriodUC fiscalperiodusecase.IFiscalPeriodUseCase,
	taxCodeUC taxcodeusecase.ITaxCodeUseCase,
	budgetUC budgetusecase.IBudgetUseCase,
	costCenterUC costcenterusecase.ICostCenterUseCase,
	reconciliationUC reconciliationusecase.IReconciliationUseCase,
) {
	accounting := r.Group("/accounting")
	{
		// Initialize handlers
		accountHandler := accounthttp.NewAccountHandler(accountUC)
		journalEntryHandler := journalentryhttp.NewJournalEntryHandler(journalEntryUC)
		fiscalPeriodHandler := fiscalperiodhttp.NewFiscalPeriodHandler(fiscalPeriodUC)
		taxCodeHandler := taxcodehttp.NewTaxCodeHandler(taxCodeUC)
		budgetHandler := budgethttp.NewBudgetHandler(budgetUC)
		costCenterHandler := costcenterhttp.NewCostCenterHandler(costCenterUC)
		reconciliationHandler := reconciliationhttp.NewReconciliationHandler(reconciliationUC)

		// Register all routes
		accountHandler.RegisterRoutes(accounting)
		journalEntryHandler.RegisterRoutes(accounting)
		fiscalPeriodHandler.RegisterRoutes(accounting)
		taxCodeHandler.RegisterRoutes(accounting)
		budgetHandler.RegisterRoutes(accounting)
		costCenterHandler.RegisterRoutes(accounting)
		reconciliationHandler.RegisterRoutes(accounting)
	}
}
