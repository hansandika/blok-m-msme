package config

import (
	"os"
	"strings"
)

type Config struct {
	DatabaseURL string
	Addr        string
	AutoSeed    bool
	AdminToken  string
}

func Load() Config {
	addr := getenv("API_ADDR", ":8080")
	if !strings.HasPrefix(addr, ":") && !strings.Contains(addr, ":") {
		addr = ":" + addr
	}
	return Config{
		DatabaseURL: getenv("DATABASE_URL", "postgres://blokm:blokm@localhost:5432/blokm?sslmode=disable"),
		Addr:        addr,
		AutoSeed:    getenv("AUTO_SEED", "1") != "0",
		AdminToken:  os.Getenv("ADMIN_TOKEN"),
	}
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
