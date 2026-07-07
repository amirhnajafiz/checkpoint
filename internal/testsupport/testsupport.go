package testsupport

import (
	"testing"

	"github.com/alicebob/miniredis/v2"

	"github.com/amirhnajafiz/mayigoo/internal/cache"
)

// NewMockCache spins up an in-process miniredis server and returns a real
// *cache.Client connected to it, plus the underlying *miniredis.Miniredis for
// direct inspection/manipulation (Set, Get, Del, FlushAll, FastForward for TTL
// testing, etc.).
//
// The server and client are closed automatically when the test finishes.
//
//	client, mr := testsupport.NewMockCache(t)
//	_ = client.SetServiceToken(ctx, 1, "tok", time.Minute)
//	mr.FastForward(2 * time.Minute) // expire it
//	mr.FlushAll()                   // or drop the whole cache
func NewMockCache(t *testing.T) (*cache.Client, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}

	client, err := cache.New(cache.Config{
		Host: mr.Host(),
		Port: mr.Port(),
	})
	if err != nil {
		mr.Close()
		t.Fatalf("connect cache to miniredis: %v", err)
	}

	t.Cleanup(func() {
		_ = client.Close()
		mr.Close()
	})

	return client, mr
}
