package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/amirhnajafiz/mayigoo/internal/models"
)

// --- Users ---

func (h *Handler) createUser(c echo.Context) error {
	var req createUserRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	user, err := h.store.CreateUser(c.Request().Context(), req.Email)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, newUserResponse(user))
}

// --- Workspaces ---

func (h *Handler) createWorkspace(c echo.Context) error {
	var req createWorkspaceRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	ws, err := h.store.CreateWorkspace(c.Request().Context(), req.UserEmail)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, newWorkspaceResponse(ws))
}

// listWorkspaces returns all workspaces, or those of a user when the
// ?user_email= query parameter is supplied.
func (h *Handler) listWorkspaces(c echo.Context) error {
	ctx := c.Request().Context()

	if email := c.QueryParam("user_email"); email != "" {
		list, err := h.store.ListWorkspacesByUser(ctx, email)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, newWorkspaceResponses(list))
	}

	list, err := h.store.ListWorkspaces(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, newWorkspaceResponses(list))
}

func (h *Handler) getWorkspace(c echo.Context) error {
	id, err := pathID(c, "id")
	if err != nil {
		return err
	}

	ws, err := h.store.GetWorkspace(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, newWorkspaceResponse(ws))
}

func (h *Handler) updateWorkspace(c echo.Context) error {
	id, err := pathID(c, "id")
	if err != nil {
		return err
	}

	var req updateWorkspaceRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	ws, err := h.store.UpdateWorkspace(c.Request().Context(), models.UpdateWorkspaceParams{
		ID:        id,
		UserEmail: req.UserEmail,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, newWorkspaceResponse(ws))
}

func (h *Handler) deleteWorkspace(c echo.Context) error {
	id, err := pathID(c, "id")
	if err != nil {
		return err
	}

	if err := h.store.DeleteWorkspace(c.Request().Context(), id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// --- Roles ---

func (h *Handler) createRole(c echo.Context) error {
	var req createRoleRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	role, err := h.store.CreateRole(c.Request().Context(), models.CreateRoleParams{
		Name:        req.Name,
		Description: req.Description,
		WorkspaceID: req.WorkspaceID,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, newRoleResponse(role))
}

// listRoles returns roles filtered by ?workspace_id= or ?account_id=. One of
// the two filters is required.
func (h *Handler) listRoles(c echo.Context) error {
	ctx := c.Request().Context()

	if accountID, ok, err := queryInt64(c, "account_id"); err != nil {
		return err
	} else if ok {
		list, err := h.store.ListRolesByAccount(ctx, accountID)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, newRoleResponses(list))
	}

	workspaceID, ok, err := queryInt64(c, "workspace_id")
	if err != nil {
		return err
	}
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "workspace_id or account_id query parameter is required")
	}

	list, err := h.store.ListRolesByWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, newRoleResponses(list))
}

func (h *Handler) updateRole(c echo.Context) error {
	id, err := pathID(c, "id")
	if err != nil {
		return err
	}

	var req updateRoleRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	role, err := h.store.UpdateRole(c.Request().Context(), models.UpdateRoleParams{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, newRoleResponse(role))
}

func (h *Handler) deleteRole(c echo.Context) error {
	id, err := pathID(c, "id")
	if err != nil {
		return err
	}

	if err := h.store.DeleteRole(c.Request().Context(), id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// --- Accounts ---

func (h *Handler) createAccount(c echo.Context) error {
	var req createAccountRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	account, err := h.store.CreateAccount(c.Request().Context(), models.CreateAccountParams{
		Name:        req.Name,
		Description: req.Description,
		WorkspaceID: req.WorkspaceID,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, newAccountResponse(account))
}

// listAccounts returns accounts filtered by ?workspace_id= or ?role_id=. One of
// the two filters is required.
func (h *Handler) listAccounts(c echo.Context) error {
	ctx := c.Request().Context()

	if roleID, ok, err := queryInt64(c, "role_id"); err != nil {
		return err
	} else if ok {
		list, err := h.store.ListAccountsByRole(ctx, roleID)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, newAccountResponses(list))
	}

	workspaceID, ok, err := queryInt64(c, "workspace_id")
	if err != nil {
		return err
	}
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "workspace_id or role_id query parameter is required")
	}

	list, err := h.store.ListAccountsByWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, newAccountResponses(list))
}

func (h *Handler) updateAccount(c echo.Context) error {
	id, err := pathID(c, "id")
	if err != nil {
		return err
	}

	var req updateAccountRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	account, err := h.store.UpdateAccount(c.Request().Context(), models.UpdateAccountParams{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, newAccountResponse(account))
}

func (h *Handler) deleteAccount(c echo.Context) error {
	id, err := pathID(c, "id")
	if err != nil {
		return err
	}

	if err := h.store.DeleteAccount(c.Request().Context(), id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// --- Role bindings (account_roles) ---

func (h *Handler) bindRole(c echo.Context) error {
	var req bindingRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	binding, err := h.store.BindAccountRole(c.Request().Context(), models.BindAccountRoleParams{
		RoleID:    req.RoleID,
		AccountID: req.AccountID,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, newBindingResponse(binding))
}

func (h *Handler) listBindings(c echo.Context) error {
	bindings, err := h.store.ListAccountRoles(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, newBindingResponses(bindings))
}

// unbindRole removes a binding identified by the ?role_id= and ?account_id=
// query parameters.
func (h *Handler) unbindRole(c echo.Context) error {
	roleID, ok, err := queryInt64(c, "role_id")
	if err != nil {
		return err
	}
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "role_id query parameter is required")
	}

	accountID, ok, err := queryInt64(c, "account_id")
	if err != nil {
		return err
	}
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "account_id query parameter is required")
	}

	if err := h.store.UnbindAccountRole(c.Request().Context(), models.UnbindAccountRoleParams{
		RoleID:    roleID,
		AccountID: accountID,
	}); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// Services.

func (h *Handler) parseServiceAccount(c echo.Context) error {
	return nil
}
