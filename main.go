package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"

	"github.com/amirhnajafiz/mayigoo/internal/auth"
	"github.com/amirhnajafiz/mayigoo/internal/cache"
	"github.com/amirhnajafiz/mayigoo/internal/config"
	"github.com/amirhnajafiz/mayigoo/internal/daemons"
	"github.com/amirhnajafiz/mayigoo/internal/db"
	httpapi "github.com/amirhnajafiz/mayigoo/internal/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Database: connect and apply migrations as a pre-execution phase.
	conn, err := db.New(cfg.DB)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer func() { _ = conn.Close() }()

	if err := db.Migrate(conn); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// Redis: the service-token cache.
	tokenCache, err := cache.New(cfg.Redis)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer func() { _ = tokenCache.Close() }()

	store := db.NewStore(conn)

	// Auth: JWT signer and Google OAuth client.
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.TTL)
	googleOAuth := auth.NewGoogleOAuth(cfg.Google.ClientID, cfg.Google.ClientSecret, cfg.Google.RedirectURL)
	// This exact string must be registered as an Authorized redirect URI on the
	// Google OAuth client, or Google returns redirect_uri_mismatch.
	log.Printf("google oauth redirect_uri: %q", cfg.Google.RedirectURL)

	// Daemons: aggregate validation usage and monitor dependency health.
	usageDaemon := daemons.NewUsageDaemon(store, cfg.Daemons.UsageFlushInterval, cfg.Daemons.UsageBufferSize)
	healthDaemon := daemons.NewHealthDaemon(cfg.Daemons.HealthPingInterval,
		daemons.Checker{Name: "postgres", Check: conn.PingContext},
		daemons.Checker{Name: "redis", Check: tokenCache.Ping},
	)
	manager := daemons.NewManager(usageDaemon, healthDaemon)

	// HTTP: wire the handler (which talks to the daemons over channels).
	handler := httpapi.NewHandler(store, jwtManager, googleOAuth, tokenCache, usageDaemon, healthDaemon)

	e := echo.New()
	e.HideBanner = true
	// Connection controls on the underlying HTTP server.
	e.Server.ReadTimeout = cfg.HTTP.ReadTimeout
	e.Server.ReadHeaderTimeout = cfg.HTTP.ReadHeaderTimeout
	e.Server.WriteTimeout = cfg.HTTP.WriteTimeout
	e.Server.IdleTimeout = cfg.HTTP.IdleTimeout
	handler.Register(e)

	// A single context cancelled on SIGINT/SIGTERM drives shutdown of both the
	// daemons and the HTTP server.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Addr, cfg.HTTP.Port)

	g, gctx := errgroup.WithContext(ctx)

	// Background daemons; the manager stops them all when gctx is cancelled.
	g.Go(func() error {
		return manager.Run(gctx)
	})

	// HTTP server with graceful shutdown tied to the group context.
	g.Go(func() error {
		go func() {
			<-gctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
			defer cancel()
			_ = e.Shutdown(shutdownCtx)
		}()

		log.Printf("mayigoo listening on %s", addr)
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("server: %v", err)
	}

	log.Print("shutdown complete")
}
