package http

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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

	var account models.ServiceAccount
	var meta models.ServiceAccountMetum
	err := h.store.ExecTx(ctx, func(q *models.Queries) error {
		a, err := q.CreateServiceAccount(ctx, models.CreateServiceAccountParams{
			Name:        req.Name,
			Description: req.Description,
			Active:      boolOrDefault(req.Active, true),
			UserEmail:   userEmail(c),
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

	// Mint a JWT for the service account (carrying its labels) and cache it.
	serviceToken, err := h.issueServiceToken(ctx, account.ID, labels)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, newServiceAccountTokenResponse(serviceAccountResponseFrom(account, meta, labels), serviceToken))
}

// issueServiceToken mints a fresh JWT for a service account (embedding its
// labels) and stores it in the cache (keyed by account id) so the issuer can
// retrieve it later and the open validate endpoint can confirm it is current.
func (h *Handler) issueServiceToken(ctx context.Context, accountID int32, labels map[string]string) (string, error) {
	serviceToken, err := h.jwtManager.Generate(strconv.FormatInt(int64(accountID), 10), auth.JWTKindService, labels)
	if err != nil {
		return "", err
	}
	if err := h.cache.SetServiceToken(ctx, accountID, serviceToken, h.jwtManager.TTL()); err != nil {
		return "", err
	}
	return serviceToken, nil
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

	var account models.ServiceAccount
	err = h.store.ExecTx(ctx, func(q *models.Queries) error {
		a, err := q.UpdateServiceAccount(ctx, models.UpdateServiceAccountParams{
			ID:          id,
			Name:        req.Name,
			Description: req.Description,
			Active:      boolOrDefault(req.Active, true),
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

	// Updating an account rotates its token: mint and cache a fresh one
	// (with the updated labels), superseding the previous token.
	serviceToken, err := h.issueServiceToken(ctx, id, labels)
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
