package http

import (
	"errors"

	"github.com/basilex/promenade/internal/contexts/accounting/account"
	"github.com/basilex/promenade/internal/contexts/accounting/account/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/account/dto"
	"github.com/basilex/promenade/internal/contexts/accounting/account/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	uc usecase.IAccountUseCase
}

func NewAccountHandler(uc usecase.IAccountUseCase) *AccountHandler {
	return &AccountHandler{uc: uc}
}

// RegisterRoutes registers account routes
func (h *AccountHandler) RegisterRoutes(router *gin.RouterGroup) {
	accounts := router.Group("/accounts")
	{
		accounts.POST("", h.CreateAccount)
		accounts.GET("/:id", h.GetAccount)
		accounts.GET("/code/:code", h.GetAccountByCode)
		accounts.PUT("/:id", h.UpdateAccount)
		accounts.PUT("/:id/parent", h.SetParent)
		accounts.PUT("/:id/activate", h.ActivateAccount)
		accounts.PUT("/:id/deactivate", h.DeactivateAccount)
		accounts.DELETE("/:id", h.DeleteAccount)
		accounts.GET("", h.ListAccounts)
		accounts.GET("/children/:parentId", h.ListChildren)
	}
}

// CreateAccount creates a new account
// @Summary Create a new account
// @Tags accounts
// @Accept json
// @Produce json
// @Param request body dto.CreateAccountRequest true "Create account request"
// @Success 201 {object} dto.AccountResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/accounts [post]
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req dto.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Parse and validate account type
	accountType := aggregate.AccountType(req.AccountType)
	if !isValidAccountType(accountType) {
		response.BadRequest(c, "invalid account type")
		return
	}

	// Parse optional parent ID
	var parentID *uuidv7.UUID
	if req.ParentID != "" {
		parsed, err := uuidv7.Parse(req.ParentID)
		if err != nil {
			response.BadRequest(c, "invalid parent_id")
			return
		}
		parentID = &parsed
	}

	// Extract organization and user from context
	orgID, _ := c.Get("organization_id")
	organizationID, _ := uuidv7.Parse(orgID.(string))
	userID, _ := c.Get("user_id")
	userUUID, _ := uuidv7.Parse(userID.(string))

	acc, err := h.uc.CreateAccount(
		c.Request.Context(),
		organizationID,
		req.Code,
		req.Name,
		accountType,
		parentID,
		userUUID,
	)
	if err != nil {
		handleAccountError(c, err)
		return
	}

	response.Created(c, dto.ToAccountResponse(acc))
}

// GetAccount gets an account by ID
// @Summary Get an account
// @Tags accounts
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} dto.AccountResponse
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/accounts/{id} [get]
func (h *AccountHandler) GetAccount(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid account ID")
		return
	}

	acc, err := h.uc.GetAccountByID(c.Request.Context(), id)
	if err != nil {
		handleAccountError(c, err)
		return
	}

	response.Success(c, dto.ToAccountResponse(acc))
}

// GetAccountByCode gets an account by code
// @Summary Get an account by code
// @Tags accounts
// @Produce json
// @Param code path string true "Account code"
// @Success 200 {object} dto.AccountResponse
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/accounts/code/{code} [get]
func (h *AccountHandler) GetAccountByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "code is required")
		return
	}

	orgID, _ := c.Get("organization_id")
	organizationID, _ := uuidv7.Parse(orgID.(string))

	acc, err := h.uc.GetAccountByCode(c.Request.Context(), organizationID, code)
	if err != nil {
		handleAccountError(c, err)
		return
	}

	response.Success(c, dto.ToAccountResponse(acc))
}

// UpdateAccount updates an account
// @Summary Update an account
// @Tags accounts
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Param request body dto.UpdateAccountRequest true "Update account request"
// @Success 200 {object} dto.AccountResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/accounts/{id} [put]
func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid account ID")
		return
	}

	var req dto.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuidv7.Parse(userID.(string))

	acc, err := h.uc.UpdateAccount(
		c.Request.Context(),
		id,
		req.Name,
		userUUID,
	)
	if err != nil {
		handleAccountError(c, err)
		return
	}

	response.Success(c, dto.ToAccountResponse(acc))
}

// SetParent sets the parent of an account
// @Summary Set parent account
// @Tags accounts
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Param request body dto.SetParentRequest true "Set parent request"
// @Success 200 {object} dto.AccountResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/accounts/{id}/parent [put]
func (h *AccountHandler) SetParent(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid account ID")
		return
	}

	var req dto.SetParentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	parentID, err := uuidv7.Parse(req.ParentID)
	if err != nil {
		response.BadRequest(c, "invalid parent_id")
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuidv7.Parse(userID.(string))

	err = h.uc.SetParent(c.Request.Context(), id, parentID, userUUID)
	if err != nil {
		handleAccountError(c, err)
		return
	}

	response.SuccessWithMessage(c, "parent set successfully")
}

// ActivateAccount activates an account
// @Summary Activate an account
// @Tags accounts
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} dto.AccountResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/accounts/{id}/activate [put]
func (h *AccountHandler) ActivateAccount(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid account ID")
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuidv7.Parse(userID.(string))

	err = h.uc.ActivateAccount(c.Request.Context(), id, userUUID)
	if err != nil {
		handleAccountError(c, err)
		return
	}

	response.SuccessWithMessage(c, "account activated successfully")
}

// DeactivateAccount deactivates an account
// @Summary Deactivate an account
// @Tags accounts
// @Tags accounts
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} dto.AccountResponse
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/accounts/{id}/deactivate [put]
func (h *AccountHandler) DeactivateAccount(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid account ID")
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuidv7.Parse(userID.(string))

	err = h.uc.DeactivateAccount(c.Request.Context(), id, userUUID)
	if err != nil {
		handleAccountError(c, err)
		return
	}

	response.SuccessWithMessage(c, "account deactivated successfully")
}

// DeleteAccount deletes an account
// @Summary Delete an account
// @Tags accounts
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/accounts/{id} [delete]
func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid account ID")
		return
	}

	if err := h.uc.DeleteAccount(c.Request.Context(), id); err != nil {
		handleAccountError(c, err)
		return
	}

	response.SuccessWithMessage(c, "account deleted successfully")
}

// ListAccounts lists all accounts for an organization
// @Summary List accounts
// @Tags accounts
// @Produce json
// @Success 200 {array} dto.AccountResponse
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/accounts [get]
func (h *AccountHandler) ListAccounts(c *gin.Context) {
	orgID, _ := c.Get("organization_id")
	organizationID, _ := uuidv7.Parse(orgID.(string))

	includeInactive := c.Query("include_inactive") == "true"
	accounts, err := h.uc.ListAccountsByOrganization(c.Request.Context(), organizationID, includeInactive)
	if err != nil {
		response.InternalError(c, "failed to list accounts")
		return
	}

	response.Success(c, dto.ToAccountResponseList(accounts))
}

// ListChildren lists child accounts
// @Summary List child accounts
// @Tags accounts
// @Produce json
// @Param parentId path string true "Parent account ID"
// @Success 200 {array} dto.AccountResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/accounting/accounts/children/{parentId} [get]
func (h *AccountHandler) ListChildren(c *gin.Context) {
	parentID, err := uuidv7.Parse(c.Param("parentId"))
	if err != nil {
		response.BadRequest(c, "invalid parent ID")
		return
	}

	children, err := h.uc.ListChildAccounts(c.Request.Context(), parentID)
	if err != nil {
		response.InternalError(c, "failed to list child accounts")
		return
	}

	response.Success(c, dto.ToAccountResponseList(children))
}

// handleAccountError handles use case errors and returns appropriate HTTP responses
func handleAccountError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, account.ErrAccountNotFound):
		response.NotFound(c, "account not found")
	case errors.Is(err, account.ErrAccountCodeEmpty),
		errors.Is(err, account.ErrAccountNameEmpty),
		errors.Is(err, account.ErrAccountInvalidType),
		errors.Is(err, account.ErrAccountCodeAlreadyExists),
		errors.Is(err, account.ErrAccountParentNotFound),
		errors.Is(err, account.ErrAccountCircularReference),
		errors.Is(err, account.ErrAccountHasChildren):
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, "operation failed")
	}
}

// isValidAccountType validates the account type
func isValidAccountType(t aggregate.AccountType) bool {
	switch t {
	case aggregate.AccountTypeAsset,
		aggregate.AccountTypeLiability,
		aggregate.AccountTypeEquity,
		aggregate.AccountTypeRevenue,
		aggregate.AccountTypeExpense:
		return true
	default:
		return false
	}
}
