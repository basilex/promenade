package router

import (
	"github.com/gin-gonic/gin"
)

// V1Router is the main router for API v1
// It aggregates all core module routers (health, auth, etc.)
// Business logic modules (posts, profiles, contacts) are registered dynamically
type V1Router struct {
	healthRouter   *HealthRouter
	authRouter     *AuthRouter
	countryRouter  *CountryRouter
	currencyRouter *CurrencyRouter
	languageRouter *LanguageRouter
	timezoneRouter *TimezoneRouter
	regionRouter   *RegionRouter
	cityRouter     *CityRouter
	rbacRouter     *RBACRouter
	adminRouter    *AdminRouter
}

// NewV1Router creates a new V1Router with all core routers
func NewV1Router(
	healthRouter *HealthRouter,
	authRouter *AuthRouter,
	countryRouter *CountryRouter,
	currencyRouter *CurrencyRouter,
	languageRouter *LanguageRouter,
	timezoneRouter *TimezoneRouter,
	regionRouter *RegionRouter,
	cityRouter *CityRouter,
	rbacRouter *RBACRouter,
	adminRouter *AdminRouter,
) *V1Router {
	return &V1Router{
		healthRouter:   healthRouter,
		authRouter:     authRouter,
		countryRouter:  countryRouter,
		currencyRouter: currencyRouter,
		languageRouter: languageRouter,
		timezoneRouter: timezoneRouter,
		regionRouter:   regionRouter,
		cityRouter:     cityRouter,
		rbacRouter:     rbacRouter,
		adminRouter:    adminRouter,
	}
}

// Setup registers all v1 core routes
// Business modules (posts, profiles, contacts, etc.) register their routes dynamically
func (r *V1Router) Setup(rg *gin.RouterGroup) {
	// Core infrastructure routes only
	r.healthRouter.Setup(rg)
	r.authRouter.Setup(rg)
	r.countryRouter.Setup(rg)
	r.currencyRouter.Setup(rg)
	r.languageRouter.Setup(rg)
	r.timezoneRouter.Setup(rg)
	r.regionRouter.RegisterRoutes(rg)
	r.cityRouter.RegisterRoutes(rg)
	r.rbacRouter.Setup(rg)
	r.adminRouter.Setup(rg)

	// Business modules register routes dynamically via module system:
	// - posts module: /posts/*, /comments/*
	// - profiles module: /profiles/*, /contacts/*
	// - warehouse module: /warehouse/* (commercial, requires license)
}
