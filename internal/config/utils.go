package config

import (
	"os"
	"strings"
)

// nestKey turns a flat env-style key into a dotted koanf path.
func nestKey(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), "__", ".")
}

// yamlPath returns the YAML file to load.
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
