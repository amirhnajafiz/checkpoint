package http

import (
	"time"

	"github.com/amirhnajafiz/mayigoo/internal/auth"
	"github.com/amirhnajafiz/mayigoo/internal/models"
)

// errorResponse is the body returned for any non-2xx response.
type errorResponse struct {
	Message string `json:"message"`
}

// loginResponse carries the signed user JWT back to API callers after a
// successful Google login.
type loginResponse struct {
	Token string `json:"token"`
}

// tokenResponse carries a single token back to the caller (e.g. the cached
// service token retrieved by its issuer).
type tokenResponse struct {
	Token string `json:"token"`
}

// serviceClaimsResponse is the unmarshaled result of validating a service
// token at the open validate endpoint.
type serviceClaimsResponse struct {
	AccountID int32             `json:"account_id"`
	Kind      string            `json:"kind"`
	Labels    map[string]string `json:"labels"`
	ExpiresAt time.Time         `json:"expires_at"`
}

func newServiceClaimsResponse(accountID int32, claims *auth.JWTClaims) serviceClaimsResponse {
	var expiresAt time.Time
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	}
	return serviceClaimsResponse{
		AccountID: accountID,
		Kind:      string(claims.JWTKind),
		Labels:    orEmptyLabels(claims.Labels),
		ExpiresAt: expiresAt,
	}
}

// serviceAccountResponse is the API view of a service account together with its
// metadata (usage counters) and key/value labels.
type serviceAccountResponse struct {
	ID          int32             `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Active      bool              `json:"active"`
	UserEmail   string            `json:"user_email"`
	Usage       int               `json:"usage"`
	KV          map[string]string `json:"kv"`
	CreatedAt   time.Time         `json:"created_at"`
	LastUsed    time.Time         `json:"last_used"`
}

// newServiceAccountResponse builds a response from a list row (account joined
// with its metadata) plus its labels.
func newServiceAccountResponse(a models.ListUserServiceAccountsRow, kv map[string]string) serviceAccountResponse {
	return serviceAccountResponse{
		ID:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		Active:      a.Active,
		UserEmail:   a.UserEmail,
		Usage:       int(a.Usage),
		KV:          orEmptyLabels(kv),
		CreatedAt:   a.CreatedAt,
		LastUsed:    nullTime(a.LastUsed),
	}
}

func newServiceAccountResponses(accounts []models.ListUserServiceAccountsRow, kvByAccount map[int32]map[string]string) []serviceAccountResponse {
	out := make([]serviceAccountResponse, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, newServiceAccountResponse(a, kvByAccount[a.ID]))
	}
	return out
}

// serviceAccountResponseFrom builds a response from a single account, its
// metadata row, and its labels, used by the create/get/update endpoints.
func serviceAccountResponseFrom(a models.ServiceAccount, m models.ServiceAccountMetum, kv map[string]string) serviceAccountResponse {
	return serviceAccountResponse{
		ID:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		Active:      a.Active,
		UserEmail:   a.UserEmail,
		Usage:       int(m.Usage),
		KV:          orEmptyLabels(kv),
		CreatedAt:   a.CreatedAt,
		LastUsed:    nullTime(m.LastUsed),
	}
}

// kvMap collapses a service account's key/value rows into a plain map.
func kvMap(rows []models.ServiceAccountKv) map[string]string {
	m := make(map[string]string, len(rows))
	for _, r := range rows {
		m[r.Xkey] = r.Xvalue
	}
	return m
}

// orEmptyLabels guarantees a non-nil map so labels serialize as {} not null.
func orEmptyLabels(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return m
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
