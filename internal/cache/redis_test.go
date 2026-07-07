package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/amirhnajafiz/mayigoo/internal/cache"
	"github.com/amirhnajafiz/mayigoo/internal/testsupport"
)

func TestServiceToken_SetGetDelete(t *testing.T) {
	client, _ := testsupport.NewMockCache(t)
	ctx := context.Background()

	if err := client.SetServiceToken(ctx, 42, "secret-token", time.Minute); err != nil {
		t.Fatalf("set: %v", err)
	}

	got, err := client.GetServiceToken(ctx, 42)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != "secret-token" {
		t.Fatalf("got %q, want %q", got, "secret-token")
	}

	if err := client.DeleteServiceToken(ctx, 42); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if _, err := client.GetServiceToken(ctx, 42); !errors.Is(err, cache.ErrNotFound) {
		t.Fatalf("after delete: got err %v, want ErrNotFound", err)
	}
}

func TestServiceToken_Expires(t *testing.T) {
	client, mr := testsupport.NewMockCache(t)
	ctx := context.Background()

	if err := client.SetServiceToken(ctx, 1, "tok", time.Minute); err != nil {
		t.Fatalf("set: %v", err)
	}

	// miniredis lets us advance its clock to trigger TTL expiry deterministically.
	mr.FastForward(2 * time.Minute)

	if _, err := client.GetServiceToken(ctx, 1); !errors.Is(err, cache.ErrNotFound) {
		t.Fatalf("after expiry: got err %v, want ErrNotFound", err)
	}
}

func TestGetServiceToken_Missing(t *testing.T) {
	client, _ := testsupport.NewMockCache(t)

	if _, err := client.GetServiceToken(context.Background(), 999); !errors.Is(err, cache.ErrNotFound) {
		t.Fatalf("got err %v, want ErrNotFound", err)
	}
}
