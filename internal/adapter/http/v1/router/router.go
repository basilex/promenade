package router

import (
	"github.com/gin-gonic/gin"
)

// V1Router is the main router for API v1
// It aggregates all module-specific routers (auth, rbac, etc.)
type V1Router struct {
	authRouter        *AuthRouter
	countryRouter     *CountryRouter
	currencyRouter    *CurrencyRouter
	userContactRouter *UserContactRouter
	// rbacRouter *RBACRouter  // Future: role-based access control
	// profileRouter *ProfileRouter  // Future: user profiles
	// notificationRouter *NotificationRouter  // Future: notifications
}

// NewV1Router creates a new V1Router with all sub-routers
func NewV1Router(
	authRouter *AuthRouter,
	countryRouter *CountryRouter,
	currencyRouter *CurrencyRouter,
	userContactRouter *UserContactRouter,
	// Add other routers here as needed
) *V1Router {
	return &V1Router{
		authRouter:        authRouter,
		countryRouter:     countryRouter,
		currencyRouter:    currencyRouter,
		userContactRouter: userContactRouter,
	}
}

// Setup registers all v1 routes by delegating to module-specific routers
func (r *V1Router) Setup(rg *gin.RouterGroup) {
	// Each module router handles its own routes under its prefix
	r.authRouter.Setup(rg)
	r.userContactRouter.Setup(rg)
	r.countryRouter.Setup(rg)
	r.currencyRouter.Setup(rg)

	// Future modules:
	// r.rbacRouter.Setup(rg)  // Will handle /rbac/*
	// r.profileRouter.Setup(rg)  // Will handle /profiles/*
	// r.notificationRouter.Setup(rg)  // Will handle /notifications/*
}
