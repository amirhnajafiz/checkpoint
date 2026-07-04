package config

import (
	"time"

	"github.com/amirhnajafiz/mayigoo/internal/cache"
	"github.com/amirhnajafiz/mayigoo/internal/db"
)

// Default returns the built-in configuration. It is the lowest-precedence
// layer, overlaid in order by an optional YAML file, an optional .env file,
// and finally MAYIGOO_-prefixed environment variables.
func Default() Config {
	return Config{
		HTTP: HTTPConfig{
			Addr: ":5000",
			Port: 5000,
		},
		DB: db.Config{
			Host:            "localhost",
			Port:            "5432",
			User:            "mayigoo",
			Password:        "mayigoo",
			Name:            "mayigoo",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    25,
			ConnMaxLifetime: 5 * time.Minute,
		},
		Redis: cache.Config{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			DB:       0,
		},
		JWT: JWTConfig{
			Secret: "dev-secret-change-me",
			TTL:    24 * time.Hour,
		},
		Google: GoogleConfig{
			ClientID:     "",
			ClientSecret: "",
			RedirectURL:  "http://localhost:5000/api/users/callback",
		},
	}
}
