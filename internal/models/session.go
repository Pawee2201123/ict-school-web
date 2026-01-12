package models

import (
	"database/sql"
	"time"
)

type Session struct {
	ID                   int
	ClassID              int
	DaySequence          int       // 1 or 2
	StartAt              time.Time
	EndAt                time.Time
	Capacity             int
	CurrentEnrolledCount int
}
type SessionDetail struct {
	SessionID            int
	ClassName            string
	RoomNumber           string
	RoomName             string
	TeacherName          string // Simplified for display
	SyllabusPDF          string
	StartAt              time.Time
	EndAt                time.Time
	Capacity             int
	CurrentEnrolledCount int
	RemainingSeats       int 
}

// CreateSession inserts one specific time slot
func CreateSession(db *sql.DB, s Session) error {
	_, err := db.Exec(`
		INSERT INTO class_sessions (
			class_id, day_sequence, start_at, end_at, capacity, current_enrolled_count
		)
		VALUES ($1, $2, $3, $4, $5, 0)
	`, 
		s.ClassID, s.DaySequence, s.StartAt, s.EndAt, s.Capacity,
	)
	return err
}

func GetSessionsByClassID(db *sql.DB, classID int) ([]Session, error) {
	rows, err := db.Query(`
		SELECT session_id, day_sequence, start_at, end_at, capacity, current_enrolled_count
		FROM class_sessions 
		WHERE class_id = $1 
		ORDER BY start_at ASC
	`, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.DaySequence, &s.StartAt, &s.EndAt, &s.Capacity, &s.CurrentEnrolledCount); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}
func GetSessionDetail(db *sql.DB, sessionID int) (*SessionDetail, error) {
	// Join Sessions with Classes to get the full picture
	query := `
		SELECT 
			cs.session_id, c.class_name, c.room_number, c.room_name, c.syllabus_pdf_url,
			cs.start_at, cs.end_at, cs.capacity, cs.current_enrolled_count,
            COALESCE(string_agg(i.name, ', '), '') as teachers
		FROM class_sessions cs
		JOIN classes c ON cs.class_id = c.class_id
        LEFT JOIN class_instructors ci ON c.class_id = ci.class_id
        LEFT JOIN instructors i ON ci.instructor_id = i.instructor_id
		WHERE cs.session_id = $1
        GROUP BY cs.session_id, c.class_id
	`
	var s SessionDetail
	err := db.QueryRow(query, sessionID).Scan(
		&s.SessionID, &s.ClassName, &s.RoomNumber, &s.RoomName, &s.SyllabusPDF,
		&s.StartAt, &s.EndAt, &s.Capacity, &s.CurrentEnrolledCount, &s.TeacherName,
	)
	if err != nil {
		return nil, err
	}
	// 2. Calculate it here
	s.RemainingSeats = s.Capacity - s.CurrentEnrolledCount

	// Optional: Prevent negative numbers if data is messy
	if s.RemainingSeats < 0 {
		s.RemainingSeats = 0
	}
	return &s, nil
}
