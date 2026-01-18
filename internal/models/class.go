package models

import (
	"database/sql"
	"time"
)

type Class struct {
	ID                  int
	ClassName           string
	SyllabusPDFURL      string
	RoomNumber          string
	RoomName            string
	RegistrationStartAt time.Time
	RegistrationEndAt   time.Time
}

// CreateClassWithInstructors (No changes needed here, this was already correct)
func CreateClassWithInstructors(db *sql.DB, c Class, teacherNames []string) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var classID int
	err = tx.QueryRow(`
		INSERT INTO classes (
			class_name, syllabus_pdf_url, room_number, room_name,
			registration_start_at, registration_end_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING class_id
	`,
		c.ClassName, c.SyllabusPDFURL, c.RoomNumber, c.RoomName,
		c.RegistrationStartAt, c.RegistrationEndAt,
	).Scan(&classID)

	if err != nil {
		return 0, err
	}

	for _, name := range teacherNames {
		if name == "" { continue }
		var instructorID int
		err = tx.QueryRow("SELECT instructor_id FROM instructors WHERE name = $1", name).Scan(&instructorID)
		if err == sql.ErrNoRows {
			err = tx.QueryRow("INSERT INTO instructors (name) VALUES ($1) RETURNING instructor_id", name).Scan(&instructorID)
			if err != nil { return 0, err }
		} else if err != nil { return 0, err }

		_, err = tx.Exec(`
			INSERT INTO class_instructors (class_id, instructor_id)
			VALUES ($1, $2)
			ON CONFLICT (class_id, instructor_id) DO NOTHING
		`, classID, instructorID)
		if err != nil { return 0, err }
	}

	return classID, nil
}

// GetClassByID: Fixed to include Syllabus and Room Number
func GetClassByID(db *sql.DB, id int) (*Class, error) {
	c := &Class{}
	// 👇 ADDED: syllabus_pdf_url, room_number
	err := db.QueryRow(`
		SELECT 
			class_id, 
			class_name, 
			COALESCE(syllabus_pdf_url, ''), 
			room_number, 
			room_name, 
			registration_start_at, 
			registration_end_at 
		FROM classes WHERE class_id = $1`, id).Scan(
		&c.ID, 
		&c.ClassName, 
		&c.SyllabusPDFURL, // 👈 Scan this
		&c.RoomNumber,     // 👈 Scan this
		&c.RoomName, 
		&c.RegistrationStartAt, 
		&c.RegistrationEndAt,
	)
	return c, err
}

// GetAllClasses: Fixed to include Syllabus and Room Number
func GetAllClasses(db *sql.DB) ([]Class, error) {
	// 👇 ADDED: syllabus_pdf_url, room_number
	rows, err := db.Query(`
		SELECT 
			class_id, 
			class_name, 
			COALESCE(syllabus_pdf_url, ''), 
			room_number, 
			room_name, 
			registration_start_at, 
			registration_end_at
		FROM classes 
		ORDER BY class_id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []Class
	for rows.Next() {
		var c Class
		// 👇 Updated Scan to match the SELECT
		if err := rows.Scan(
			&c.ID, 
			&c.ClassName, 
			&c.SyllabusPDFURL, // 👈 Critical Fix
			&c.RoomNumber,     // 👈 Critical Fix
			&c.RoomName, 
			&c.RegistrationStartAt, 
			&c.RegistrationEndAt,
		); err != nil {
			return nil, err
		}
		classes = append(classes, c)
	}
	return classes, nil
}
