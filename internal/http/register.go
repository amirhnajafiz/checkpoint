package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/amirhnajafiz/mayigoo/internal/auth"
	"github.com/amirhnajafiz/mayigoo/internal/cache"
	"github.com/amirhnajafiz/mayigoo/internal/daemons"
	"github.com/amirhnajafiz/mayigoo/internal/db"
)

// Handler holds the dependencies shared by every endpoint.
type Handler struct {
	store       *db.Store
	jwtManager  *auth.JWTManager
	googleOAuth *auth.GoogleOAuth
	cache       *cache.Client
	usage       *daemons.UsageDaemon
	health      *daemons.HealthDaemon
	logger      *zap.Logger
}

// NewHandler builds a Handler backed by the store, JWT manager, Google OAuth
// client, token cache, the background daemons it talks to over channels, and the
// logger.
func NewHandler(store *db.Store, jwtm *auth.JWTManager, googleoa *auth.GoogleOAuth, tokenCache *cache.Client, usage *daemons.UsageDaemon, health *daemons.HealthDaemon, logger *zap.Logger) *Handler {
	return &Handler{
		store:       store,
		jwtManager:  jwtm,
		googleOAuth: googleoa,
		cache:       tokenCache,
		usage:       usage,
		health:      health,
		logger:      logger,
	}
}

// Register configures middlewares, the request validator, the JSON error
// handler, and wires every route onto the given Echo instance.
func (h *Handler) Register(e *echo.Echo) {
	e.Validator = newValidator()
	e.Renderer = newTemplateRenderer()
	e.HTTPErrorHandler = errorHandler

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod: true,
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(_ echo.Context, v middleware.RequestLoggerValues) error {
			h.logger.Info("request",
				zap.String("method", v.Method),
				zap.String("uri", v.URI),
				zap.Int("status", v.Status))
			return nil
		},
	}))
	e.Use(middleware.Recover())

	// Server-rendered client pages.
	e.GET("/", h.loginPage)
	e.GET("/app", h.appPage)

	api := e.Group("/api")

	// Health: the latest dependency-health snapshot from the health daemon (public).
	api.GET("/health", h.healthCheck)

	// Users: Google OAuth login flow (public).
	users := api.Group("/users")
	users.GET("/login", h.login)
	users.GET("/callback", h.callback)

	// Services: open endpoint to validate & unmarshal a service token supplied
	// via the Authorization: Bearer header.
	services := api.Group("/services")
	services.GET("/validate", h.validateService)

	// Service accounts: authenticated and scoped to the caller.
	accounts := api.Group("/accounts")
	accounts.Use(h.authMiddleware)
	accounts.POST("", h.createAccount)
	accounts.GET("", h.listAccounts)
	accounts.GET("/:id", h.getAccount)
	accounts.GET("/:id/token", h.getAccountToken)
	accounts.PUT("/:id", h.updateAccount)
	accounts.DELETE("/:id", h.deleteAccount)
}
