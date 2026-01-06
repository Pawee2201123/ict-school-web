package models

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type User struct {
	ID           int
	Email        string
	PasswordHash string
	CreatedAt    string
}

type UserProfile struct {
    ID           int
    UserID       int
    StudentName  sql.NullString
    SchoolName   sql.NullString
    Grade        sql.NullString
    GuardianName sql.NullString
}

var ErrUserExists = errors.New("user already exists")

func CreateUser(db *sql.DB, email, passwordHash string) (int, error) {
	var id int
	err := db.QueryRow(
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		email, passwordHash,
	).Scan(&id)
	if err != nil {
		// detect unique violation
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return 0, ErrUserExists
		}
		return 0, err
	}
	return id, nil
}

func CreateUserProfile(db *sql.DB, userID int, name, school, grade, guardian string) error {
    _, err := db.Exec(`
        INSERT INTO user_profiles (user_id, student_name, school_name, grade, guardian_name)
        VALUES ($1, $2, $3, $4, $5)
    `, userID, name, school, grade, guardian)
    return err
}

func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	u := &User{}
	err := db.QueryRow(
		`SELECT id, email, password_hash, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}
func GetUserProfile(db *sql.DB, userID int) (*UserProfile, error) {
    p := &UserProfile{}
    err := db.QueryRow(`
        SELECT id, user_id, student_name, school_name, grade, guardian_name
        FROM user_profiles
        WHERE user_id = $1
        LIMIT 1
    `, userID).Scan(&p.ID, &p.UserID, &p.StudentName, &p.SchoolName, &p.Grade, &p.GuardianName)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil // no profile yet
        }
        return nil, err
    }
    return p, nil
}
