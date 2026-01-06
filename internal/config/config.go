package config

import (
	"fmt"
	"os"
)

type Config struct {
	DB_DSN      string
	CookieHash  string
	CookieBlock string
	ListenAddr  string
}

func Load() Config {
	return Config{
		// Centralized Logic: Try DATABASE_URL first, otherwise build it manually
		DB_DSN:      getDatabaseDSN(),
		CookieHash:  getEnv("COOKIE_HASH_KEY", ""),
		CookieBlock: getEnv("COOKIE_BLOCK_KEY", ""),
		ListenAddr:  getEnv("LISTEN_ADDR", ":8080"),
	}
}

func getDatabaseDSN() string {
	// 1. Try simple URL first
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}

	// 2. Fallback to individual vars (moved from db.go)
	host := getEnv("PGHOST", "127.0.0.1")
	port := getEnv("PGPORT", "5432")
	user := getEnv("PGUSER", "postgres")
	password := getEnv("PGPASSWORD", "")
	dbname := getEnv("PGDATABASE", "postgres")
	sslmode := getEnv("PGSSLMODE", "disable")

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
