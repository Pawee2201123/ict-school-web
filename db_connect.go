package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Read connection info from env (safe & flexible)
	// Example env variables you can set:
	//   export PGHOST=127.0.0.1
	//   export PGPORT=5432
	//   export PGUSER=postgres
	//   export PGPASSWORD=yourpassword
	//   export PGDATABASE=postgres
	//   export PGSSLMODE=disable
	host := getenv("PGHOST", "127.0.0.1")
	port := getenv("PGPORT", "5432")
	user := getenv("PGUSER", "postgres")
	password := getenv("PGPASSWORD", "")
	dbname := getenv("PGDATABASE", "postgres")
	sslmode := getenv("PGSSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Error opening DB: %v", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging DB: %v", err)
	}
	fmt.Println("✅ Connected to Postgres successfully")

	// Check if 'users' table exists and show row count (safe read)
	var exists bool
	err = db.QueryRow(`
        SELECT EXISTS (
          SELECT 1 FROM information_schema.tables
          WHERE table_schema = 'public' AND table_name = 'users'
        );
    `).Scan(&exists)
	if err != nil {
		log.Fatalf("Error checking for users table: %v", err)
	}

	if !exists {
		fmt.Println("ℹ️  'users' table does not exist in schema public.")
		fmt.Println("Run the CREATE TABLE statement provided earlier to create it.")
		return
	}

	var count int
	err = db.QueryRow(`SELECT count(*) FROM public.users;`).Scan(&count)
	if err != nil {
		log.Fatalf("Error counting users: %v", err)
	}
	fmt.Printf("✅ 'users' table exists — row count = %d\n", count)
}

// helper
func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
