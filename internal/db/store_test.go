//go:build integration

package db_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/amirhnajafiz/mayigoo/internal/models"
	"github.com/amirhnajafiz/mayigoo/internal/testsupport"
)

// makeAccount inserts a user (service_accounts.user_email is a FK to users) and
// a service account owned by it, returning the created account.
func makeAccount(t *testing.T, ctx context.Context, store interface {
	UpsertUser(context.Context, string) (models.User, error)
	CreateServiceAccount(context.Context, models.CreateServiceAccountParams) (models.ServiceAccount, error)
}) models.ServiceAccount {
	t.Helper()

	if _, err := store.UpsertUser(ctx, "owner@example.com"); err != nil {
		t.Fatalf("UpsertUser: %v", err)
	}
	sa, err := store.CreateServiceAccount(ctx, models.CreateServiceAccountParams{
		Name:      "svc",
		Active:    true,
		UserEmail: "owner@example.com",
	})
	if err != nil {
		t.Fatalf("CreateServiceAccount: %v", err)
	}
	return sa
}

// Happy path: a row that was inserted can be read back.
func TestGetServiceAccount_Exists(t *testing.T) {
	store := testsupport.NewRealStore(t)
	ctx := context.Background()

	sa := makeAccount(t, ctx, store)

	got, err := store.GetServiceAccount(ctx, sa.ID)
	if err != nil {
		t.Fatalf("GetServiceAccount: %v", err)
	}
	if got.Name != "svc" || got.UserEmail != "owner@example.com" {
		t.Fatalf("unexpected account: %+v", got)
	}
}

// Your scenario: manually remove a row, then read it — real empty state yields
// sql.ErrNoRows, exercising the not-found branch of your code.
func TestGetServiceAccount_AfterDelete(t *testing.T) {
	store := testsupport.NewRealStore(t)
	ctx := context.Background()

	sa := makeAccount(t, ctx, store)

	if err := store.DeleteServiceAccount(ctx, sa.ID); err != nil {
		t.Fatalf("DeleteServiceAccount: %v", err)
	}

	_, err := store.GetServiceAccount(ctx, sa.ID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("after delete: got err %v, want sql.ErrNoRows", err)
	}
}

// Your scenario: query an id that never existed.
func TestGetServiceAccount_Missing(t *testing.T) {
	store := testsupport.NewRealStore(t)

	_, err := store.GetServiceAccount(context.Background(), 999999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("got err %v, want sql.ErrNoRows", err)
	}
}

// Real-state transaction semantics: a rollback must leave no row behind.
func TestExecTx_Rollback_LeavesNoRow(t *testing.T) {
	store := testsupport.NewRealStore(t)
	ctx := context.Background()

	if _, err := store.UpsertUser(ctx, "owner@example.com"); err != nil {
		t.Fatalf("UpsertUser: %v", err)
	}

	wantErr := errors.New("boom")
	err := store.ExecTx(ctx, func(q *models.Queries) error {
		if _, err := q.CreateServiceAccount(ctx, models.CreateServiceAccountParams{
			Name:      "doomed",
			Active:    true,
			UserEmail: "owner@example.com",
		}); err != nil {
			return err
		}
		return wantErr // force rollback
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("ExecTx: got %v, want %v", err, wantErr)
	}

	// The account created inside the rolled-back tx must not persist.
	accounts, err := store.ListUserServiceAccounts(ctx, "owner@example.com")
	if err != nil {
		t.Fatalf("ListUserServiceAccounts: %v", err)
	}
	if len(accounts) != 0 {
		t.Fatalf("rollback left %d accounts, want 0", len(accounts))
	}
}
