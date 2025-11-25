package config

import "os"

type Config struct {
	DB_DSN      string
	CookieHash  string
	CookieBlock string
	ListenAddr  string
}

func Load() Config {
	return Config{
		DB_DSN:      getEnv("DATABASE_URL", ""),
		CookieHash:  getEnv("COOKIE_HASH_KEY", ""),
		CookieBlock: getEnv("COOKIE_BLOCK_KEY", ""),
		ListenAddr:  getEnv("LISTEN_ADDR", ":8080"),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
