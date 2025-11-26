
package handlers

import (
	"database/sql"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func EnsureAdmin(db *sql.DB) error {
	email := os.Getenv("ADMIN_EMAIL")
	pw := os.Getenv("ADMIN_PASSWORD")
	if email == "" || pw == "" {
		return nil // no admin setup
	}

	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// upsert admin user
	_, err = db.Exec(`
		INSERT INTO users (email, password_hash, is_admin)
		VALUES ($1, $2, true)
		ON CONFLICT (email)
		DO UPDATE SET password_hash = $2, is_admin = true
	`, email, string(hash))

	if err != nil {
		return err
	}

	log.Printf("✔ Admin user ensured: %s", email)
	return nil
}
