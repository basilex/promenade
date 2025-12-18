package router

import (
	"github.com/gin-gonic/gin"
)

// V1Router is the main router for API v1
// It aggregates all module-specific routers (health, auth, etc.)
type V1Router struct {
	healthRouter      *HealthRouter
	authRouter        *AuthRouter
	countryRouter     *CountryRouter
	currencyRouter    *CurrencyRouter
	userContactRouter *UserContactRouter
	userProfileRouter *UserProfileRouter
	userPostRouter    *UserPostRouter
	postCommentRouter *PostCommentRouter
	rbacRouter        *RBACRouter
	// notificationRouter *NotificationRouter  // Future: notifications
}

// NewV1Router creates a new V1Router with all sub-routers
func NewV1Router(
	healthRouter *HealthRouter,
	authRouter *AuthRouter,
	countryRouter *CountryRouter,
	currencyRouter *CurrencyRouter,
	userContactRouter *UserContactRouter,
	userProfileRouter *UserProfileRouter,
	userPostRouter *UserPostRouter,
	postCommentRouter *PostCommentRouter,
	rbacRouter *RBACRouter,
	// Add other routers here as needed
) *V1Router {
	return &V1Router{
		healthRouter:      healthRouter,
		authRouter:        authRouter,
		countryRouter:     countryRouter,
		currencyRouter:    currencyRouter,
		userContactRouter: userContactRouter,
		userProfileRouter: userProfileRouter,
		userPostRouter:    userPostRouter,
		postCommentRouter: postCommentRouter,
		rbacRouter:        rbacRouter,
	}
}

// Setup registers all v1 routes by delegating to module-specific routers
func (r *V1Router) Setup(rg *gin.RouterGroup) {
	// Each module router handles its own routes under its prefix
	r.healthRouter.Setup(rg)
	r.authRouter.Setup(rg)
	r.userPostRouter.Setup(rg)
	r.postCommentRouter.Setup(rg)
	r.userContactRouter.Setup(rg)
	r.userProfileRouter.Setup(rg)
	r.countryRouter.Setup(rg)
	r.currencyRouter.Setup(rg)
	r.rbacRouter.Setup(rg)

	// Future modules:
	// r.notificationRouter.Setup(rg)  // Will handle /notifications/*
}
