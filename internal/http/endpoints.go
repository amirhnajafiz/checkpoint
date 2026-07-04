package http

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/amirhnajafiz/mayigoo/internal/auth"
	"github.com/amirhnajafiz/mayigoo/internal/cache"
	"github.com/amirhnajafiz/mayigoo/internal/models"
)

const oauthStateCookie = "oauth_state"

// --- Authentication ---

// login starts the Google OAuth flow: it stores a CSRF state in a cookie and
// redirects the browser to Google's consent screen.
func (h *Handler) login(c echo.Context) error {
	state, err := randomState()
	if err != nil {
		return err
	}

	c.SetCookie(&http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return c.Redirect(http.StatusFound, h.googleOAuth.AuthCodeURL(state))
}

// callback completes the Google OAuth flow: it validates the state, exchanges
// the code for the user's email, upserts the user, and returns a signed JWT.
func (h *Handler) callback(c echo.Context) error {
	cookie, err := c.Cookie(oauthStateCookie)
	if err != nil || cookie.Value == "" || cookie.Value != c.QueryParam("state") {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid oauth state")
	}

	code := c.QueryParam("code")
	if code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing authorization code")
	}

	ctx := c.Request().Context()

	oauthToken, err := h.googleOAuth.Exchange(ctx, code)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "oauth code exchange failed")
	}

	email, err := h.googleOAuth.Email(ctx, oauthToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "failed to resolve google account email")
	}

	user, err := h.store.UpsertUser(ctx, email)
	if err != nil {
		return err
	}

	signed, err := h.jwtManager.Generate(user.Email, auth.JWTKindUser, nil)
	if err != nil {
		return err
	}

	// API clients (Accept: application/json) get the token as JSON; browsers
	// are redirected to the dashboard with the token in the URL fragment
	// (never sent to the server) so the client can store it in localStorage.
	if strings.Contains(c.Request().Header.Get("Accept"), "application/json") {
		return c.JSON(http.StatusOK, loginResponse{Token: signed})
	}

	return c.Redirect(http.StatusFound, "/app#token="+url.QueryEscape(signed))
}

// --- Service accounts (authenticated; scoped to the caller) ---

// createAccount creates a service account owned by the caller, together with
// its metadata row, then mints a service JWT for it.
func (h *Handler) createAccount(c echo.Context) error {
	var req createServiceAccountRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	labels := orEmptyLabels(req.KV)
	active := boolOrDefault(req.Active, true)
	ttlSeconds, err := parseTTL(req.TTL)
	if err != nil {
		return err
	}

	var account models.ServiceAccount
	var meta models.ServiceAccountMetum
	err = h.store.ExecTx(ctx, func(q *models.Queries) error {
		a, err := q.CreateServiceAccount(ctx, models.CreateServiceAccountParams{
			Name:        req.Name,
			Description: req.Description,
			Active:      active,
			UserEmail:   userEmail(c),
			TtlSeconds:  ttlSeconds,
		})
		if err != nil {
			return err
		}
		if _, err := q.CreateServiceAccountMeta(ctx, a.ID); err != nil {
			return err
		}
		if err := setAccountKV(ctx, q, a.ID, labels); err != nil {
			return err
		}
		account = a
		meta, err = q.GetServiceAccountMeta(ctx, a.ID)
		return err
	})
	if err != nil {
		return err
	}

	// Cache a fresh token only if the account is active; an inactive account
	// has no usable token. The token lives for the account's TTL (or the default).
	serviceToken, err := h.syncServiceToken(ctx, account.ID, active, labels, h.resolveTTL(account.TtlSeconds))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, newServiceAccountTokenResponse(serviceAccountResponseFrom(account, meta, labels), serviceToken))
}

// syncServiceToken reconciles the cached token with the account's active state.
// An active account gets a freshly minted, cached token (rotating any previous
// one) that lives for ttl; an inactive account has its cached token removed so
// it can no longer validate. The returned token is empty for an inactive account.
func (h *Handler) syncServiceToken(ctx context.Context, accountID int32, active bool, labels map[string]string, ttl time.Duration) (string, error) {
	if !active {
		if err := h.cache.DeleteServiceToken(ctx, accountID); err != nil {
			return "", err
		}
		return "", nil
	}

	return h.issueServiceToken(ctx, accountID, labels, ttl)
}

// issueServiceToken mints a fresh JWT for a service account (embedding its
// labels, valid for ttl) and stores it in the cache (keyed by account id, with a
// matching expiry) so the issuer can retrieve it later and the open validate
// endpoint can confirm it is current.
func (h *Handler) issueServiceToken(ctx context.Context, accountID int32, labels map[string]string, ttl time.Duration) (string, error) {
	serviceToken, err := h.jwtManager.GenerateWithTTL(strconv.FormatInt(int64(accountID), 10), auth.JWTKindService, labels, ttl)
	if err != nil {
		return "", err
	}
	if err := h.cache.SetServiceToken(ctx, accountID, serviceToken, ttl); err != nil {
		return "", err
	}
	return serviceToken, nil
}

// resolveTTL picks the token lifetime for an account: its per-account TTL when
// set, otherwise the manager's default fallback TTL.
func (h *Handler) resolveTTL(ttlSeconds sql.NullInt64) time.Duration {
	if d, ok := ttlDuration(ttlSeconds); ok {
		return d
	}
	return h.jwtManager.TTL()
}

// listAccounts returns the service accounts owned by the caller, each with its
// labels.
func (h *Handler) listAccounts(c echo.Context) error {
	ctx := c.Request().Context()
	email := userEmail(c)

	list, err := h.store.ListUserServiceAccounts(ctx, email)
	if err != nil {
		return err
	}

	kvRows, err := h.store.ListUserServiceAccountKV(ctx, email)
	if err != nil {
		return err
	}
	kvByAccount := make(map[int32]map[string]string)
	for _, r := range kvRows {
		m := kvByAccount[r.AccountID]
		if m == nil {
			m = make(map[string]string)
			kvByAccount[r.AccountID] = m
		}
		m[r.Xkey] = r.Xvalue
	}

	return c.JSON(http.StatusOK, newServiceAccountResponses(list, kvByAccount))
}

func (h *Handler) getAccount(c echo.Context) error {
	id, err := pathID(c, "id")
	if err != nil {
		return err
	}

	account, err := h.ownedAccount(c, id)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	meta, err := h.store.GetServiceAccountMeta(ctx, id)
	if err != nil {
		return err
	}

	kvRows, err := h.store.ListServiceAccountKV(ctx, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, serviceAccountResponseFrom(account, meta, kvMap(kvRows)))
}

func (h *Handler) updateAccount(c echo.Context) error {
	id, err := pathID(c, "id")
	if err != nil {
		return err
	}

	if _, err := h.ownedAccount(c, id); err != nil {
		return err
	}

	var req updateServiceAccountRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	labels := orEmptyLabels(req.KV)
	active := boolOrDefault(req.Active, true)
	ttlSeconds, err := parseTTL(req.TTL)
	if err != nil {
		return err
	}

	var account models.ServiceAccount
	err = h.store.ExecTx(ctx, func(q *models.Queries) error {
		a, err := q.UpdateServiceAccount(ctx, models.UpdateServiceAccountParams{
			ID:          id,
			Name:        req.Name,
			Description: req.Description,
			Active:      active,
			TtlSeconds:  ttlSeconds,
		})
		if err != nil {
			return err
		}
		account = a
		return setAccountKV(ctx, q, id, labels)
	})
	if err != nil {
		return err
	}

	meta, err := h.store.GetServiceAccountMeta(ctx, id)
	if err != nil {
		return err
	}

	// Reconcile the cached token with the new active state: an active account
	// gets a freshly rotated token (with the updated labels and TTL), while an
	// inactive account has its token evicted so it stops validating.
	serviceToken, err := h.syncServiceToken(ctx, id, active, labels, h.resolveTTL(account.TtlSeconds))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, newServiceAccountTokenResponse(serviceAccountResponseFrom(account, meta, labels), serviceToken))
}

func (h *Handler) deleteAccount(c echo.Context) error {
	id, err := pathID(c, "id")
	if err != nil {
		return err
	}

	if _, err := h.ownedAccount(c, id); err != nil {
		return err
	}

	ctx := c.Request().Context()

	// service_account_meta and service_account_kv cascade on delete.
	if err := h.store.DeleteServiceAccount(ctx, id); err != nil {
		return err
	}

	// Drop the cached token; ignore a missing entry.
	if err := h.cache.DeleteServiceToken(ctx, id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// getAccountToken returns the cached JWT for one of the caller's service
// accounts. Only the issuer (owner) can retrieve it.
func (h *Handler) getAccountToken(c echo.Context) error {
	id, err := pathID(c, "id")
	if err != nil {
		return err
	}

	if _, err := h.ownedAccount(c, id); err != nil {
		return err
	}

	serviceToken, err := h.cache.GetServiceToken(c.Request().Context(), id)
	if errors.Is(err, cache.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "no token cached for this account")
	}
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, tokenResponse{Token: serviceToken})
}

// validateService is an open endpoint: given a service token it verifies the
// signature, confirms it is the account's current cached token, and returns the
// unmarshaled claims.
func (h *Handler) validateService(c echo.Context) error {
	raw := bearerToken(c)
	if raw == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "missing bearer token")
	}

	claims, err := h.jwtManager.Parse(raw)
	if err != nil || claims.JWTKind != auth.JWTKindService {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid service token")
	}

	accountID, err := strconv.ParseInt(claims.Subject, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid service token")
	}

	// Confirm the presented token is the account's current cached token; an
	// older (rotated) or evicted token is treated as invalid.
	cached, err := h.cache.GetServiceToken(c.Request().Context(), int32(accountID))
	if errors.Is(err, cache.ErrNotFound) || (err == nil && cached != raw) {
		return echo.NewHTTPError(http.StatusUnauthorized, "service token is not active")
	}
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, newServiceClaimsResponse(int32(accountID), claims))
}

// ownedAccount fetches a service account and confirms it belongs to the caller,
// returning 404 otherwise so account existence is not leaked across users.
func (h *Handler) ownedAccount(c echo.Context, id int32) (models.ServiceAccount, error) {
	account, err := h.store.GetServiceAccount(c.Request().Context(), id)
	if err != nil {
		return models.ServiceAccount{}, err
	}

	if account.UserEmail != userEmail(c) {
		return models.ServiceAccount{}, echo.NewHTTPError(http.StatusNotFound, "resource not found")
	}

	return account, nil
}
