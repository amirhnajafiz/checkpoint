package cache

import "net"

// Config holds the settings needed to reach a Redis instance.
type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// Addr renders the host:port Redis address.
func (c Config) Addr() string {
	return net.JoinHostPort(c.Host, c.Port)
}
