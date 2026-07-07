//go:build integration

// End-to-end API tests: real HTTP requests are driven through the Echo router
// against a real PostgreSQL (testcontainers) and a real Redis (miniredis), with
// real JWT auth. The API is treated as a black box — requests and responses are
// plain JSON over the wire, so these tests exercise routing, middleware,
// validation, the error handler, the store, and the cache together.
//
// Run with: go test -tags=integration ./internal/http/...
package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/amirhnajafiz/mayigoo/internal/auth"
	"github.com/amirhnajafiz/mayigoo/internal/daemons"
	apphttp "github.com/amirhnajafiz/mayigoo/internal/http"
	"github.com/amirhnajafiz/mayigoo/internal/testsupport"
)

const testUser = "owner@example.com"

// api bundles the wired-up router and the seams a test wants to poke.
type api struct {
	e         *echo.Echo
	jwt       *auth.JWTManager
	userToken string               // a valid user JWT for testUser
	redis     *miniredis.Miniredis // to simulate cache drops / expiry
}

// setupAPI builds a Handler backed by a real store and cache, registers the
// routes on a fresh Echo, seeds the owning user, and mints a user token.
// googleOAuth and health are nil because the endpoints under test never touch
// them; usage is a real (unstarted) daemon whose Record is a safe no-op.
func setupAPI(t *testing.T) api {
	t.Helper()

	store := testsupport.NewRealStore(t)
	cacheClient, mr := testsupport.NewMockCache(t)
	jwtm := auth.NewJWTManager("test-secret", time.Hour)
	usage := daemons.NewUsageDaemon(store, zap.NewNop(), time.Minute, 16)

	h := apphttp.NewHandler(store, jwtm, nil, cacheClient, usage, nil, zap.NewNop())
	e := echo.New()
	h.Register(e)

	// service_accounts.user_email is a FK to users, so the owner must exist.
	if _, err := store.UpsertUser(context.Background(), testUser); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	token, err := jwtm.Generate(testUser, auth.JWTKindUser, nil)
	if err != nil {
		t.Fatalf("generate user token: %v", err)
	}

	return api{e: e, jwt: jwtm, userToken: token, redis: mr}
}

// do issues one request through the router and returns the recorder plus the
// JSON body decoded into a generic map (empty when there is no body).
func (a api) do(t *testing.T, method, path, token string, body any) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	if token != "" {
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	a.e.ServeHTTP(rec, req)

	parsed := map[string]any{}
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
			// Some bodies (e.g. list endpoints) are arrays, not objects; callers
			// that need those decode rec.Body themselves. Leave parsed empty.
			parsed = map[string]any{}
		}
	}
	return rec, parsed
}

// createAccount POSTs a new service account and returns its id and cached token.
func (a api) createAccount(t *testing.T, name string, kv map[string]string) (int32, string) {
	t.Helper()

	rec, body := a.do(t, http.MethodPost, "/api/accounts", a.userToken, map[string]any{
		"name": name,
		"kv":   kv,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create account: status %d, body %s", rec.Code, rec.Body.String())
	}

	id, ok := body["id"].(float64)
	if !ok {
		t.Fatalf("create account: missing id in %v", body)
	}
	token, _ := body["token"].(string)
	if token == "" {
		t.Fatalf("create account: missing token in %v", body)
	}
	return int32(id), token
}

// TestAPI_AccountLifecycle walks create → get → get-token → delete → and then
// the two "empty state" reads: the deleted row is gone (404) and its cached
// token was evicted (404).
func TestAPI_AccountLifecycle(t *testing.T) {
	a := setupAPI(t)

	id, createdToken := a.createAccount(t, "billing", map[string]string{"env": "prod"})

	// GET the account: real row, labels round-tripped.
	rec, body := a.do(t, http.MethodGet, fmt.Sprintf("/api/accounts/%d", id), a.userToken, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get account: status %d, body %s", rec.Code, rec.Body.String())
	}
	if body["name"] != "billing" {
		t.Fatalf("get account: name = %v, want billing", body["name"])
	}
	if kv, _ := body["kv"].(map[string]any); kv["env"] != "prod" {
		t.Fatalf("get account: kv = %v, want env=prod", body["kv"])
	}

	// GET the cached token: the issuer can retrieve exactly what was minted.
	rec, body = a.do(t, http.MethodGet, fmt.Sprintf("/api/accounts/%d/token", id), a.userToken, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get token: status %d, body %s", rec.Code, rec.Body.String())
	}
	if body["token"] != createdToken {
		t.Fatalf("get token: returned a different token than was created")
	}

	// DELETE the account.
	rec, _ = a.do(t, http.MethodDelete, fmt.Sprintf("/api/accounts/%d", id), a.userToken, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete account: status %d, body %s", rec.Code, rec.Body.String())
	}

	// The row is really gone: sql.ErrNoRows → 404 via the error handler.
	rec, _ = a.do(t, http.MethodGet, fmt.Sprintf("/api/accounts/%d", id), a.userToken, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get after delete: status %d, want 404", rec.Code)
	}

	// The cached token was evicted on delete: 404.
	rec, _ = a.do(t, http.MethodGet, fmt.Sprintf("/api/accounts/%d/token", id), a.userToken, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get token after delete: status %d, want 404", rec.Code)
	}
}

// TestAPI_RequiresAuth confirms the auth middleware rejects an unauthenticated
// request to a protected route.
func TestAPI_RequiresAuth(t *testing.T) {
	a := setupAPI(t)

	rec, _ := a.do(t, http.MethodGet, "/api/accounts", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token: status %d, want 401", rec.Code)
	}
}

// TestAPI_OwnershipScoping confirms one user cannot read another user's account:
// the response is 404 (not 403) so existence isn't leaked across users.
func TestAPI_OwnershipScoping(t *testing.T) {
	a := setupAPI(t)

	id, _ := a.createAccount(t, "private", nil)

	// A token for a different user (no DB row needed — the middleware only parses
	// the JWT; ownership is checked against the account's user_email).
	intruderToken, err := a.jwt.Generate("intruder@example.com", auth.JWTKindUser, nil)
	if err != nil {
		t.Fatalf("generate intruder token: %v", err)
	}

	rec, _ := a.do(t, http.MethodGet, fmt.Sprintf("/api/accounts/%d", id), intruderToken, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross-user get: status %d, want 404", rec.Code)
	}
}

// TestAPI_ValidateServiceToken exercises the cache↔JWT interplay end-to-end: a
// freshly issued service token validates, and after the cache is dropped (your
// "manually drop redis" scenario) the same token is rejected as inactive.
func TestAPI_ValidateServiceToken(t *testing.T) {
	a := setupAPI(t)

	_, serviceToken := a.createAccount(t, "worker", nil)

	// The open validate endpoint accepts the current cached token.
	rec, body := a.do(t, http.MethodGet, "/api/services/validate", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("validate without token: status %d, want 401", rec.Code)
	}

	rec, body = a.do(t, http.MethodGet, "/api/services/validate", serviceToken, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("validate: status %d, body %s", rec.Code, rec.Body.String())
	}
	if body["kind"] != string(auth.JWTKindService) {
		t.Fatalf("validate: kind = %v, want service", body["kind"])
	}

	// Manually drop the whole Redis cache; the (still-signed) token must now be
	// rejected because it is no longer the account's current cached token.
	a.redis.FlushAll()

	rec, _ = a.do(t, http.MethodGet, "/api/services/validate", serviceToken, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("validate after cache drop: status %d, want 401", rec.Code)
	}
}
