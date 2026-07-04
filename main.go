package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/amirhnajafiz/mayigoo/internal/auth"
	"github.com/amirhnajafiz/mayigoo/internal/cache"
	"github.com/amirhnajafiz/mayigoo/internal/db"
	httpapi "github.com/amirhnajafiz/mayigoo/internal/http"
)

func main() {
	// Database: connect and apply migrations as a pre-execution phase.
	conn, err := db.New(db.Config{
		Host:            env("DB_HOST", "localhost"),
		Port:            env("DB_PORT", "5432"),
		User:            env("DB_USER", "mayigoo"),
		Password:        env("DB_PASSWORD", "mayigoo"),
		Name:            env("DB_NAME", "mayigoo"),
		SSLMode:         env("DB_SSLMODE", "disable"),
		MaxOpenConns:    25,
		MaxIdleConns:    25,
		ConnMaxLifetime: 5 * time.Minute,
	})
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer func() { _ = conn.Close() }()

	if err := db.Migrate(conn); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// Redis: the service-token cache.
	tokenCache, err := cache.New(cache.Config{
		Host:     env("REDIS_HOST", "localhost"),
		Port:     env("REDIS_PORT", "6379"),
		Password: env("REDIS_PASSWORD", ""),
		DB:       envInt("REDIS_DB", 0),
	})
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer func() { _ = tokenCache.Close() }()

	// Auth: JWT signer and Google OAuth client.
	jwtManager := auth.NewJWTManager(
		env("JWT_SECRET", "dev-secret-change-me"),
		envDuration("JWT_TTL", 24*time.Hour),
	)
	googleRedirect := env("GOOGLE_REDIRECT_URL", "http://localhost:5000/api/users/callback")
	googleOAuth := auth.NewGoogleOAuth(
		env("GOOGLE_CLIENT_ID", ""),
		env("GOOGLE_CLIENT_SECRET", ""),
		googleRedirect,
	)
	// This exact string must be registered as an Authorized redirect URI on the
	// Google OAuth client, or Google returns redirect_uri_mismatch.
	log.Printf("google oauth redirect_uri: %q", googleRedirect)

	// HTTP: wire the handler and start Echo.
	handler := httpapi.NewHandler(db.NewStore(conn), jwtManager, googleOAuth, tokenCache)

	e := echo.New()
	e.HideBanner = true
	handler.Register(e)

	addr := env("HTTP_ADDR", ":5000")
	log.Printf("mayigoo listening on %s", addr)
	if err := e.Start(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
