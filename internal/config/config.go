package config

import (
	"fmt"
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
	// EnvPrefix is the required prefix for environment-variable overrides.
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
	HTTP    HTTPConfig    `koanf:"http"`
	DB      db.Config     `koanf:"db"`
	Redis   cache.Config  `koanf:"redis"`
	JWT     JWTConfig     `koanf:"jwt"`
	Google  GoogleConfig  `koanf:"oauth"`
	Daemons DaemonsConfig `koanf:"daemons"`
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

// DaemonsConfig tunes the background daemons.
type DaemonsConfig struct {
	// UsageFlushInterval is how often batched validation events are written to
	// the database.
	UsageFlushInterval time.Duration `koanf:"usage_flush_interval"`
	// UsageBufferSize is the capacity of the validation-event channel; events
	// are dropped rather than blocking the request path when it is full.
	UsageBufferSize int `koanf:"usage_buffer_size"`
	// HealthPingInterval is how often dependency health is probed.
	HealthPingInterval time.Duration `koanf:"health_ping_interval"`
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
	// a double underscore.
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
