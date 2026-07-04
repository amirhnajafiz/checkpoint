package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	oauth "github.com/amirhnajafiz/mayigoo/internal/auth"
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

	signed, err := h.jwtManager.Generate(user.Email, oauth.JWTKindUser)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, loginResponse{Token: signed})
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
		m, err := q.CreateServiceAccountMeta(ctx, a.ID)
		if err != nil {
			return err
		}
		account = a
		meta = m
		return nil
	})
	if err != nil {
		return err
	}

	// Mint a JWT for the service account and cache it in Redis.
	serviceToken, err := h.issueServiceToken(ctx, account.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, newServiceAccountTokenResponse(serviceAccountResponseFrom(account, meta), serviceToken))
}

// issueServiceToken mints a fresh JWT for a service account and stores it in the
// cache (keyed by account id) so the issuer can retrieve it later and the open
// validate endpoint can confirm it is the current token.
func (h *Handler) issueServiceToken(ctx context.Context, accountID int32) (string, error) {
	serviceToken, err := h.jwtManager.Generate(strconv.FormatInt(int64(accountID), 10), oauth.JWTKindService)
	if err != nil {
		return "", err
	}
	if err := h.cache.SetServiceToken(ctx, accountID, serviceToken, h.jwtManager.TTL()); err != nil {
		return "", err
	}
	return serviceToken, nil
}

// listAccounts returns the service accounts owned by the caller.
func (h *Handler) listAccounts(c echo.Context) error {
	list, err := h.store.ListUserServiceAccounts(c.Request().Context(), userEmail(c))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, newServiceAccountResponses(list))
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

	meta, err := h.store.GetServiceAccountMeta(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, serviceAccountResponseFrom(account, meta))
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

	account, err := h.store.UpdateServiceAccount(c.Request().Context(), models.UpdateServiceAccountParams{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Active:      boolOrDefault(req.Active, true),
	})
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	meta, err := h.store.GetServiceAccountMeta(ctx, id)
	if err != nil {
		return err
	}

	// Updating an account rotates its token: mint and cache a fresh one,
	// superseding the previous token.
	serviceToken, err := h.issueServiceToken(ctx, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, newServiceAccountTokenResponse(serviceAccountResponseFrom(account, meta), serviceToken))
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
	var req validateServiceRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	claims, err := h.jwtManager.Parse(req.Token)
	if err != nil || claims.JWTKind != oauth.JWTKindService {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid service token")
	}

	accountID, err := strconv.ParseInt(claims.Subject, 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid service token")
	}

	// Confirm the presented token is the account's current cached token; an
	// older (rotated) or evicted token is treated as invalid.
	cached, err := h.cache.GetServiceToken(c.Request().Context(), int32(accountID))
	if errors.Is(err, cache.ErrNotFound) || (err == nil && cached != req.Token) {
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
