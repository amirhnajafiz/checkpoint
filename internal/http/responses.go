package http

import (
	"time"

	"github.com/amirhnajafiz/mayigoo/internal/models"
)

// errorResponse is the body returned for any non-2xx response.
type errorResponse struct {
	Message string `json:"message"`
}

// userResponse is the API view of a models.User.
type userResponse struct {
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func newUserResponse(u models.User) userResponse {
	return userResponse{
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

func newUserResponses(users []models.User) []userResponse {
	out := make([]userResponse, 0, len(users))
	for _, u := range users {
		out = append(out, newUserResponse(u))
	}
	return out
}

// workspaceResponse is the API view of a models.Workspace.
type workspaceResponse struct {
	ID        int32  `json:"id"`
	UserEmail string `json:"user_email"`
}

func newWorkspaceResponse(w models.Workspace) workspaceResponse {
	return workspaceResponse{
		ID:        w.ID,
		UserEmail: w.UserEmail,
	}
}

func newWorkspaceResponses(workspaces []models.Workspace) []workspaceResponse {
	out := make([]workspaceResponse, 0, len(workspaces))
	for _, w := range workspaces {
		out = append(out, newWorkspaceResponse(w))
	}
	return out
}

// roleResponse is the API view of a models.Role.
type roleResponse struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	WorkspaceID int64  `json:"workspace_id"`
}

func newRoleResponse(r models.Role) roleResponse {
	return roleResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		WorkspaceID: r.WorkspaceID,
	}
}

func newRoleResponses(roles []models.Role) []roleResponse {
	out := make([]roleResponse, 0, len(roles))
	for _, r := range roles {
		out = append(out, newRoleResponse(r))
	}
	return out
}

// accountResponse is the API view of a models.Account.
type accountResponse struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	WorkspaceID int64  `json:"workspace_id"`
}

func newAccountResponse(a models.Account) accountResponse {
	return accountResponse{
		ID:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		WorkspaceID: a.WorkspaceID,
	}
}

func newAccountResponses(accounts []models.Account) []accountResponse {
	out := make([]accountResponse, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, newAccountResponse(a))
	}
	return out
}

// bindingResponse is the API view of a models.AccountRole binding.
type bindingResponse struct {
	RoleID    int64 `json:"role_id"`
	AccountID int64 `json:"account_id"`
}

func newBindingResponse(b models.AccountRole) bindingResponse {
	return bindingResponse{
		RoleID:    b.RoleID,
		AccountID: b.AccountID,
	}
}

func newBindingResponses(bindings []models.AccountRole) []bindingResponse {
	out := make([]bindingResponse, 0, len(bindings))
	for _, b := range bindings {
		out = append(out, newBindingResponse(b))
	}
	return out
}
