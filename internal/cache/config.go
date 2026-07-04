package cache

import "net"

// Config holds the settings needed to reach a Redis instance.
type Config struct {
	Host     string `koanf:"host"`
	Port     string `koanf:"port"`
	Password string `koanf:"password"`
	DB       int    `koanf:"db"`
}

// Addr renders the host:port Redis address.
func (c Config) Addr() string {
	return net.JoinHostPort(c.Host, c.Port)
}
