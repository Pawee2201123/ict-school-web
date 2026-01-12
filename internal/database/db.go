package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

// Connect now ONLY accepts a DSN. It does not look at Env vars.
func Connect(dsn string) *sql.DB {
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
