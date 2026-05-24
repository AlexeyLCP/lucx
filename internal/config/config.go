package config

import "flag"

type Config struct {
	ListenAddr string
	DBPath     string
	JWTSecret  string
}

func Parse() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.ListenAddr, "listen", ":8744", "API listen address")
	flag.StringVar(&cfg.DBPath, "db", "./lucx.db", "SQLite database path")
	flag.StringVar(&cfg.JWTSecret, "jwt-secret", "", "JWT signing secret (generate if empty)")
	flag.Parse()
	return cfg
}
