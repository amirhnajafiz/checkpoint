package http

import (
	"time"

	"github.com/amirhnajafiz/mayigoo/internal/models"
)

// errorResponse is the body returned for any non-2xx response.
type errorResponse struct {
	Message string `json:"message"`
}

// loginResponse carries the signed user JWT back to the caller after a
// successful Google login.
type loginResponse struct {
	Token string `json:"token"`
}

// serviceAccountResponse is the API view of a service account together with its
// metadata (usage counters).
type serviceAccountResponse struct {
	ID          int32     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Active      bool      `json:"active"`
	UserEmail   string    `json:"user_email"`
	Usage       int       `json:"usage"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsed    time.Time `json:"last_used"`
}

// newServiceAccountResponse builds a response from a list row (account joined
// with its metadata).
func newServiceAccountResponse(a models.ListUserServiceAccountsRow) serviceAccountResponse {
	return serviceAccountResponse{
		ID:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		Active:      a.Active,
		UserEmail:   a.UserEmail,
		Usage:       int(a.Usage),
		CreatedAt:   a.CreatedAt,
		LastUsed:    nullTime(a.LastUsed),
	}
}

func newServiceAccountResponses(accounts []models.ListUserServiceAccountsRow) []serviceAccountResponse {
	out := make([]serviceAccountResponse, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, newServiceAccountResponse(a))
	}
	return out
}

// serviceAccountResponseFrom builds a response from a single account and its
// metadata row, used by the create/get/update endpoints.
func serviceAccountResponseFrom(a models.ServiceAccount, m models.ServiceAccountMetum) serviceAccountResponse {
	return serviceAccountResponse{
		ID:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		Active:      a.Active,
		UserEmail:   a.UserEmail,
		Usage:       int(m.Usage),
		CreatedAt:   a.CreatedAt,
		LastUsed:    nullTime(m.LastUsed),
	}
}

// serviceAccountTokenResponse is returned on creation: the account plus the
// freshly minted service JWT (to be stored in Redis in a future task).
type serviceAccountTokenResponse struct {
	serviceAccountResponse
	Token string `json:"token"`
}

func newServiceAccountTokenResponse(base serviceAccountResponse, tok string) serviceAccountTokenResponse {
	return serviceAccountTokenResponse{
		serviceAccountResponse: base,
		Token:                  tok,
	}
}
