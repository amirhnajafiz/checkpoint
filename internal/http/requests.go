package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// createUserRequest is the JSON body for POST /api/users.
type createUserRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// createWorkspaceRequest is the JSON body for POST /api/workspaces.
type createWorkspaceRequest struct {
	UserEmail string `json:"user_email" validate:"required,email"`
}

// updateWorkspaceRequest is the JSON body for PUT /api/workspaces/:id.
type updateWorkspaceRequest struct {
	UserEmail string `json:"user_email" validate:"required,email"`
}

// createRoleRequest is the JSON body for POST /api/roles.
type createRoleRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	WorkspaceID int64  `json:"workspace_id" validate:"required"`
}

// updateRoleRequest is the JSON body for PUT /api/roles/:id.
type updateRoleRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

// createAccountRequest is the JSON body for POST /api/accounts.
type createAccountRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	WorkspaceID int64  `json:"workspace_id" validate:"required"`
}

// updateAccountRequest is the JSON body for PUT /api/accounts/:id.
type updateAccountRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

// bindingRequest is the JSON body for POST /api/role-bindings.
type bindingRequest struct {
	RoleID    int64 `json:"role_id" validate:"required"`
	AccountID int64 `json:"account_id" validate:"required"`
}

// bindAndValidate binds the JSON body into req and runs struct validation,
// returning a 400 HTTPError on failure.
func bindAndValidate(c echo.Context, req any) error {
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return err
	}
	return nil
}

// pathID reads an integer path parameter (e.g. /:id) as an int32.
func pathID(c echo.Context, name string) (int32, error) {
	v, err := strconv.ParseInt(c.Param(name), 10, 32)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+name)
	}
	return int32(v), nil
}

// queryInt64 reads an integer query parameter. found is false when the
// parameter is absent; a present-but-invalid value returns a 400 HTTPError.
func queryInt64(c echo.Context, name string) (value int64, found bool, err error) {
	raw := c.QueryParam(name)
	if raw == "" {
		return 0, false, nil
	}
	v, parseErr := strconv.ParseInt(raw, 10, 64)
	if parseErr != nil {
		return 0, false, echo.NewHTTPError(http.StatusBadRequest, "invalid "+name)
	}
	return v, true, nil
}
