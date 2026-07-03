package http

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/amirhnajafiz/mayigoo/internal/db"
)

// Handler holds the dependencies shared by every endpoint.
type Handler struct {
	store *db.Store
}

// NewHandler builds a Handler backed by the given store.
func NewHandler(store *db.Store) *Handler {
	return &Handler{store: store}
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

	// Handler groups.
	api := e.Group("/api")
	auth := api.Group("/auth")
	services := api.Group("/services")
	workspaces := api.Group("/workspaces")
	roles := workspaces.Group("/roles")
	accounts := workspaces.Group("/accounts")
	bindings := workspaces.Group("/role-bindings")

	// Users.
	auth.POST("/", h.createUser)

	// Services.
	services.GET("/", h.parseServiceAccount)

	// Workspaces.
	workspaces.POST("/", h.createWorkspace)
	workspaces.GET("/", h.listWorkspaces)
	workspaces.GET("/:id", h.getWorkspace)
	workspaces.PUT("/:id", h.updateWorkspace)
	workspaces.DELETE("/:id", h.deleteWorkspace)

	// Roles.
	roles.POST("/", h.createRole)
	roles.GET("/", h.listRoles)
	roles.PUT("/:id", h.updateRole)
	roles.DELETE("/:id", h.deleteRole)

	// Accounts.
	accounts.POST("/", h.createAccount)
	accounts.GET("/", h.listAccounts)
	accounts.PUT("/:id", h.updateAccount)
	accounts.DELETE("/:id", h.deleteAccount)

	// Role bindings.
	bindings.POST("/", h.bindRole)
	bindings.GET("/", h.listBindings)
	bindings.DELETE("/:id", h.unbindRole)
}
