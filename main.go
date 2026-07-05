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
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/amirhnajafiz/mayigoo/internal/auth"
	"github.com/amirhnajafiz/mayigoo/internal/cache"
	"github.com/amirhnajafiz/mayigoo/internal/config"
	"github.com/amirhnajafiz/mayigoo/internal/daemons"
	"github.com/amirhnajafiz/mayigoo/internal/db"
	httpapi "github.com/amirhnajafiz/mayigoo/internal/http"
	"github.com/amirhnajafiz/mayigoo/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		// The logger is not built yet, so fall back to the standard logger.
		log.Fatalf("config: %v", err)
	}

	l, err := logger.New(cfg.Logger)
	if err != nil {
		log.Fatalf("logger: %v", err)
	}
	defer func() { _ = l.Sync() }()

	// Database: connect and apply migrations as a pre-execution phase.
	conn, err := db.New(cfg.DB)
	if err != nil {
		l.Fatal("database", zap.Error(err))
	}
	defer func() { _ = conn.Close() }()

	if err := db.Migrate(conn); err != nil {
		l.Fatal("migrate", zap.Error(err))
	}

	// Redis: the service-token cache.
	tokenCache, err := cache.New(cfg.Redis)
	if err != nil {
		l.Fatal("redis", zap.Error(err))
	}
	defer func() { _ = tokenCache.Close() }()

	store := db.NewStore(conn)

	// Auth: JWT signer and Google OAuth client.
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.TTL)
	googleOAuth := auth.NewGoogleOAuth(cfg.Google.ClientID, cfg.Google.ClientSecret, cfg.Google.RedirectURL)
	// This exact string must be registered as an Authorized redirect URI on the
	// Google OAuth client, or Google returns redirect_uri_mismatch.
	l.Info("google oauth configured", zap.String("redirect_uri", cfg.Google.RedirectURL))

	// Daemons: aggregate validation usage and monitor dependency health.
	usageDaemon := daemons.NewUsageDaemon(store, l, cfg.Daemons.UsageFlushInterval, cfg.Daemons.UsageBufferSize)
	healthDaemon := daemons.NewHealthDaemon(cfg.Daemons.HealthPingInterval,
		daemons.Checker{Name: "postgres", Check: conn.PingContext},
		daemons.Checker{Name: "redis", Check: tokenCache.Ping},
	)
	manager := daemons.NewManager(l, usageDaemon, healthDaemon)

	// HTTP: wire the handler (which talks to the daemons over channels).
	handler := httpapi.NewHandler(store, jwtManager, googleOAuth, tokenCache, usageDaemon, healthDaemon, l)

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

		l.Info("http server starting", zap.String("addr", addr))
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		l.Fatal("server", zap.Error(err))
	}

	l.Info("shutdown complete")
}
