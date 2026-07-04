package http

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	oauth "github.com/amirhnajafiz/mayigoo/internal/auth"
	"github.com/amirhnajafiz/mayigoo/internal/db"
)

// Handler holds the dependencies shared by every endpoint.
type Handler struct {
	store       *db.Store
	jwtManager  *oauth.JWTManager
	googleOAuth *oauth.GoogleOAuth
}

// NewHandler builds a Handler backed by the store, JWT manager, and Google
// OAuth client.
func NewHandler(store *db.Store, jwtm *oauth.JWTManager, googleoa *oauth.GoogleOAuth) *Handler {
	return &Handler{store: store, jwtManager: jwtm, googleOAuth: googleoa}
}

// Register configures middlewares, the request validator, the JSON error
// handler, and wires every route onto the given Echo instance.
func (h *Handler) Register(e *echo.Echo) {
	e.Validator = newValidator()
	e.HTTPErrorHandler = errorHandler

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod: true,
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(_ echo.Context, v middleware.RequestLoggerValues) error {
			log.Printf("%s %s %d", v.Method, v.URI, v.Status)
			return nil
		},
	}))
	e.Use(middleware.Recover())

	api := e.Group("/api")

	// Users: Google OAuth login flow (public).
	users := api.Group("/users")
	users.GET("/login", h.login)
	users.GET("/callback", h.callback)

	// Service accounts: authenticated and scoped to the caller.
	accounts := api.Group("/accounts")
	accounts.Use(h.authMiddleware)
	accounts.POST("", h.createAccount)
	accounts.GET("", h.listAccounts)
	accounts.GET("/:id", h.getAccount)
	accounts.PUT("/:id", h.updateAccount)
	accounts.DELETE("/:id", h.deleteAccount)
}
