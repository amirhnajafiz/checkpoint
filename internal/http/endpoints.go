package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	oauth "github.com/amirhnajafiz/mayigoo/internal/auth"
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

	// Mint a JWT for the service account.
	// TODO: persist this token in Redis (future task).
	serviceToken, err := h.jwtManager.Generate(strconv.FormatInt(int64(account.ID), 10), oauth.JWTKindService)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, newServiceAccountTokenResponse(serviceAccountResponseFrom(account, meta), serviceToken))
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

	meta, err := h.store.GetServiceAccountMeta(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, serviceAccountResponseFrom(account, meta))
}

func (h *Handler) deleteAccount(c echo.Context) error {
	id, err := pathID(c, "id")
	if err != nil {
		return err
	}

	if _, err := h.ownedAccount(c, id); err != nil {
		return err
	}

	// service_account_meta and service_account_kv cascade on delete.
	if err := h.store.DeleteServiceAccount(c.Request().Context(), id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
