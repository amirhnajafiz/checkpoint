package main

import (
	"fmt"
	"log"

	"github.com/labstack/echo/v4"

	"github.com/amirhnajafiz/mayigoo/internal/auth"
	"github.com/amirhnajafiz/mayigoo/internal/cache"
	"github.com/amirhnajafiz/mayigoo/internal/config"
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

	// Auth: JWT signer and Google OAuth client.
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.TTL)
	googleOAuth := auth.NewGoogleOAuth(cfg.Google.ClientID, cfg.Google.ClientSecret, cfg.Google.RedirectURL)
	// This exact string must be registered as an Authorized redirect URI on the
	// Google OAuth client, or Google returns redirect_uri_mismatch.
	log.Printf("google oauth redirect_uri: %q", cfg.Google.RedirectURL)

	// HTTP: wire the handler and start Echo.
	handler := httpapi.NewHandler(db.NewStore(conn), jwtManager, googleOAuth, tokenCache)

	e := echo.New()
	e.HideBanner = true
	handler.Register(e)

	log.Printf("mayigoo listening on %s:%d", cfg.HTTP.Addr, cfg.HTTP.Port)
	if err := e.Start(fmt.Sprintf("%s:%d", cfg.HTTP.Addr, cfg.HTTP.Port)); err != nil {
		log.Fatalf("server: %v", err)
	}
}
