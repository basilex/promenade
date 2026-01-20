package banking

import (
	"github.com/gin-gonic/gin"

	accounthttp "github.com/basilex/promenade/internal/contexts/banking/bankaccount/adapter/http"
	accountusecase "github.com/basilex/promenade/internal/contexts/banking/bankaccount/usecase"
	txhttp "github.com/basilex/promenade/internal/contexts/banking/banktransaction/adapter/http"
	txusecase "github.com/basilex/promenade/internal/contexts/banking/banktransaction/usecase"
)

// RegisterRoutes registers all banking context routes
func RegisterRoutes(
	r *gin.RouterGroup,
	accountUC accountusecase.IBankAccountUseCase,
	txUC txusecase.IBankTransactionUseCase,
) {
	banking := r.Group("/banking")
	{
		// Initialize handlers
		accountHandler := accounthttp.NewBankAccountHandler(accountUC)
		txHandler := txhttp.NewBankTransactionHandler(txUC)

		// Bank Account routes
		accounts := banking.Group("/accounts")
		{
			accounts.POST("/manual", accountHandler.CreateManual)
			accounts.POST("/connect", accountHandler.ConnectProvider)
			accounts.GET("", accountHandler.List)
			accounts.GET("/:id", accountHandler.GetByID)
			accounts.PUT("/:id/details", accountHandler.UpdateDetails)
			accounts.PUT("/:id/balance", accountHandler.UpdateBalance)
			accounts.POST("/:id/sync", accountHandler.RecordSync)
			accounts.POST("/:id/activate", accountHandler.Activate)
			accounts.POST("/:id/deactivate", accountHandler.Deactivate)
			accounts.POST("/:id/archive", accountHandler.Archive)
			accounts.DELETE("/:id", accountHandler.Delete)
		}

		// Bank Transaction routes
		transactions := banking.Group("/transactions")
		{
			transactions.POST("", txHandler.Record)
			transactions.GET("", txHandler.ListByAccount)
			transactions.GET("/unmatched", txHandler.ListUnmatched)
			transactions.GET("/:id", txHandler.GetByID)
			transactions.POST("/:id/book", txHandler.Book)
			transactions.POST("/:id/cancel", txHandler.Cancel)
			transactions.POST("/:id/match", txHandler.Match)
			transactions.POST("/:id/unmatch", txHandler.Unmatch)
			transactions.PUT("/:id/counterparty", txHandler.SetCounterparty)
			transactions.DELETE("/:id", txHandler.Delete)
		}
	}
}
