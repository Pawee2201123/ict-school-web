package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// Connect accepts a full DSN or falls back to env vars (same pattern you already used)
func Connect(dsn string) *sql.DB {
	if dsn == "" {
		host := getenv("PGHOST", "127.0.0.1")
		port := getenv("PGPORT", "5432")
		user := getenv("PGUSER", "postgres")
		password := getenv("PGPASSWORD", "")
		dbname := getenv("PGDATABASE", "postgres")
		sslmode := getenv("PGSSLMODE", "disable")
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("Connected to Postgres")
	return db
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
