package config

import (
	"flag"
	"os"
)

// Config holds all configuration for LucX Core.
type Config struct {
	ListenAddr string
	DBPath     string
	JWTSecret  string
}

// Parse reads configuration from command-line flags and returns a populated Config.
// Uses a local FlagSet to avoid conflicts with go test flags.
func Parse() *Config {
	cfg := &Config{}
	fs := flag.NewFlagSet("lucx-core", flag.ExitOnError)
	fs.StringVar(&cfg.ListenAddr, "listen", ":8744", "API listen address")
	fs.StringVar(&cfg.DBPath, "db", "./lucx.db", "SQLite database path")
	fs.StringVar(&cfg.JWTSecret, "jwt-secret", "", "JWT signing secret (generate if empty)")
	fs.Parse(os.Args[1:])
	return cfg
}
