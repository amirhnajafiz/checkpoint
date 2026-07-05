package logger

// Config selects the log encoding and verbosity.
type Config struct {
	// Format is "json" for structured logs or "text" for a human-readable
	// console encoding.
	Format string `koanf:"format"`
	// Level is one of debug, info, warn, error.
	Level string `koanf:"level"`
}
