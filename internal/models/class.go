package models

import (
	"database/sql"
	"time"
)

// Matches your new "classes" table
type Class struct {
	ID                  int
	ClassName           string
	SyllabusPDFURL      string
	RoomNumber          string
	RoomName            string
	RegistrationStartAt time.Time
	RegistrationEndAt   time.Time
}

// CreateClassWithInstructors inserts the class and links it to instructors
func CreateClassWithInstructors(db *sql.DB, c Class, teacherNames []string) (int, error) {
	// 1. Start Transaction
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	
	// Safety: Rollback if anything fails, Commit if everything succeeds
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 2. Insert into 'classes'
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

	// 3. Insert/Find Instructors and Link them
	for _, name := range teacherNames {
		if name == "" {
			continue
		}

		// A. Check if instructor exists
		var instructorID int
		err = tx.QueryRow("SELECT instructor_id FROM instructors WHERE name = $1", name).Scan(&instructorID)

		if err == sql.ErrNoRows {
			// B. If not, create new instructor
			err = tx.QueryRow("INSERT INTO instructors (name) VALUES ($1) RETURNING instructor_id", name).Scan(&instructorID)
			if err != nil {
				return 0, err
			}
		} else if err != nil {
			return 0, err
		}

		// C. Link in 'class_instructors'
		_, err = tx.Exec(`
			INSERT INTO class_instructors (class_id, instructor_id)
			VALUES ($1, $2)
			ON CONFLICT (class_id, instructor_id) DO NOTHING
		`, classID, instructorID)
		if err != nil {
			return 0, err
		}
	}

	return classID, nil
}
func GetClassByID(db *sql.DB, id int) (*Class, error) {
	c := &Class{}
	// Simple query (omitting instructors join for brevity, but you can add it)
	err := db.QueryRow(`
	SELECT class_id, class_name, room_name, registration_start_at, registration_end_at 
	FROM classes WHERE class_id = $1`, id).Scan(
		&c.ID, &c.ClassName, &c.RoomName, &c.RegistrationStartAt, &c.RegistrationEndAt,
	)
	return c, err
}
// GetAllClasses fetches all classes for the admin list
func GetAllClasses(db *sql.DB) ([]Class, error) {
	rows, err := db.Query(`
		SELECT class_id, class_name, room_name, registration_start_at, registration_end_at
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
		if err := rows.Scan(&c.ID, &c.ClassName, &c.RoomName, &c.RegistrationStartAt, &c.RegistrationEndAt); err != nil {
			return nil, err
		}
		classes = append(classes, c)
	}
	return classes, nil
}
