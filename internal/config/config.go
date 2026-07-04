// Package config loads application settings, layering sources from lowest to
// highest precedence: built-in defaults, an optional YAML file, an optional
// .env file, and finally MAYIGOO_-prefixed environment variables.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"

	"github.com/amirhnajafiz/mayigoo/internal/cache"
	"github.com/amirhnajafiz/mayigoo/internal/db"
)

const (
	// EnvPrefix is the required prefix for environment-variable overrides;
	// e.g. MAYIGOO_DB__HOST maps to the "db.host" key.
	EnvPrefix = "MAYIGOO_"
	// envConfigFile names an environment variable pointing at a YAML file.
	envConfigFile = "MAYIGOO_CONFIG_FILE"
	// defaultConfigFile is loaded automatically when present in the cwd.
	defaultConfigFile = "config.yaml"
	// dotEnvFile is loaded automatically when present in the cwd.
	dotEnvFile = ".env"
	// structTag is the field tag koanf reads for keys.
	structTag = "koanf"
)

// Config is the fully-resolved application configuration.
type Config struct {
	HTTP   HTTPConfig   `koanf:"http"`
	DB     db.Config    `koanf:"db"`
	Redis  cache.Config `koanf:"redis"`
	JWT    JWTConfig    `koanf:"jwt"`
	Google GoogleConfig `koanf:"oauth"`
}

// HTTPConfig configures the HTTP server.
type HTTPConfig struct {
	Addr string `koanf:"addr"`
	Port int    `koanf:"port"`
}

// JWTConfig configures token signing.
type JWTConfig struct {
	Secret string        `koanf:"secret"`
	TTL    time.Duration `koanf:"ttl"`
}

// GoogleConfig configures the Google OAuth client.
type GoogleConfig struct {
	ClientID     string `koanf:"client_id"`
	ClientSecret string `koanf:"client_secret"`
	RedirectURL  string `koanf:"redirect_url"`
}

// Load resolves configuration from all sources in precedence order and
// unmarshals the result into a Config.
func Load() (*Config, error) {
	k := koanf.New(".")

	// Built-in defaults, read straight from the Default() struct.
	if err := k.Load(structs.Provider(Default(), structTag), nil); err != nil {
		return nil, fmt.Errorf("load defaults: %w", err)
	}

	// Optional YAML file (nested sections: http, db, redis, jwt, oauth).
	if path := yamlPath(); path != "" {
		if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
			return nil, fmt.Errorf("load yaml %q: %w", path, err)
		}
	}

	// Optional .env file: keys are flat and case-insensitive, sections split on
	// a double underscore (DB__HOST -> db.host, OAUTH__CLIENT_ID -> oauth.client_id).
	if exists(dotEnvFile) {
		if err := k.Load(file.Provider(dotEnvFile), dotenv.ParserEnv("", ".", nestKey)); err != nil {
			return nil, fmt.Errorf("load .env: %w", err)
		}
	}

	// MAYIGOO_-prefixed environment variables override everything.
	envCb := func(s string) string {
		return nestKey(strings.TrimPrefix(s, EnvPrefix))
	}
	if err := k.Load(env.Provider(EnvPrefix, ".", envCb), nil); err != nil {
		return nil, fmt.Errorf("load env: %w", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}

// nestKey turns a flat env-style key into a dotted koanf path. A double
// underscore separates sections while a single underscore is a word break
// within a field name: DB__SSL_MODE -> db.ssl_mode, OAUTH__CLIENT_ID ->
// oauth.client_id.
func nestKey(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), "__", ".")
}

// yamlPath returns the YAML file to load: MAYIGOO_CONFIG_FILE if set, otherwise
// config.yaml when it exists, otherwise "" (no YAML source).
func yamlPath() string {
	if p := os.Getenv(envConfigFile); p != "" {
		return p
	}
	if exists(defaultConfigFile) {
		return defaultConfigFile
	}
	return ""
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
