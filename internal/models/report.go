package models

import (
	"database/sql"
	"fmt"
	"time"
)

// 1. Struct for Class Status (Monitor Table)
type ClassStatusReport struct {
	ClassName    string
	SessionTime  string // Formatted "Day X 10:00~11:00"
	Capacity     int
	Count        int
	Instructors  string
	RoomName     string
}

// 2. Struct for Applicants (CSV Export)
type ApplicantReport struct {
	UserID       int
	StudentName  string
	GuardianName string
	SchoolName   string
	Grade        string
	Email        string
	RegDate      time.Time
	ClassName    string
	SessionTime  string
}

// GetClassStatusReport fetches data for the "Live Monitor" and Class Info CSV
func GetClassStatusReport(db *sql.DB, classID int, sessionID int) ([]ClassStatusReport, error) {
	query := `
		SELECT 
			c.class_name, 
			s.day_sequence, s.start_at, s.end_at, 
			s.capacity, COALESCE(s.current_enrolled_count, 0),
			c.room_name,
			COALESCE(string_agg(i.name, ', '), '') as instructors
		FROM classes c
		JOIN class_sessions s ON c.class_id = s.class_id
		LEFT JOIN class_instructors ci ON c.class_id = ci.class_id
		LEFT JOIN instructors i ON ci.instructor_id = i.instructor_id
		WHERE 1=1
	`
	
	// Dynamic Filtering
	var args []any
	argCounter := 1

	if classID > 0 {
		query += fmt.Sprintf(" AND c.class_id = $%d", argCounter)
		args = append(args, classID)
		argCounter++
	}

	if sessionID > 0 {
		query += fmt.Sprintf(" AND s.session_id = $%d", argCounter)
		args = append(args, sessionID)
		argCounter++
	}

	query += `
		GROUP BY 
			c.class_id, c.class_name, c.room_name, 
			s.session_id, s.day_sequence, s.start_at, s.end_at, s.capacity, s.current_enrolled_count
		ORDER BY c.class_id, s.start_at
	`

	rows, err := db.Query(query, args...)
	if err != nil { return nil, err }
	defer rows.Close()

	var reports []ClassStatusReport
	for rows.Next() {
		var r ClassStatusReport
		var daySeq int
		var start, end time.Time
		
		err := rows.Scan(
			&r.ClassName, &daySeq, &start, &end, 
			&r.Capacity, &r.Count, &r.RoomName, &r.Instructors,
		)
		if err != nil { return nil, err }

		r.SessionTime = fmt.Sprintf("Day %d %s~%s", daySeq, start.Format("15:04"), end.Format("15:04"))
		reports = append(reports, r)
	}
	return reports, nil
}
// GetApplicantsReport fetches the main list for CSV Export
func GetApplicantsReport(db *sql.DB, classID int, sessionID int) ([]ApplicantReport, error) {
	// Base Query
	query := `
		SELECT 
			u.id, u.email, e.registered_at,
			up.student_name, up.guardian_name, up.school_name, up.grade,
			c.class_name, s.day_sequence, s.start_at, s.end_at
		FROM session_enrollments e
		JOIN user_profiles up ON e.user_profile_id = up.id
		JOIN users u ON up.user_id = u.id
		JOIN class_sessions s ON e.session_id = s.session_id
		JOIN classes c ON s.class_id = c.class_id
		WHERE 1=1 
	`
	
	// Dynamic Filtering
	var args []any
	argCounter := 1

	if classID > 0 {
		query += fmt.Sprintf(" AND c.class_id = $%d", argCounter)
		args = append(args, classID)
		argCounter++
	}

	if sessionID > 0 {
		query += fmt.Sprintf(" AND s.session_id = $%d", argCounter)
		args = append(args, sessionID)
		argCounter++
	}

	query += ` ORDER BY s.start_at, u.id`

	rows, err := db.Query(query, args...)
	if err != nil { return nil, err }
	defer rows.Close()

	var reports []ApplicantReport
	for rows.Next() {
		var r ApplicantReport
		var daySeq int
		var start, end time.Time
		var createdAt time.Time

		err := rows.Scan(
			&r.UserID, &r.Email, &createdAt,
			&r.StudentName, &r.GuardianName, &r.SchoolName, &r.Grade,
			&r.ClassName, &daySeq, &start, &end,
		)
		if err != nil { return nil, err }

		r.RegDate = createdAt
		r.SessionTime = fmt.Sprintf("Day %d %s-%s", daySeq, start.Format("15:04"), end.Format("15:04"))
		reports = append(reports, r)
	}
	return reports, nil
}

// NEW Helper: Need to fetch all sessions to populate the dropdown
type SessionOption struct {
    ID          int
    ClassID     int
    DisplayName string
}

func GetAllSessionsForDropdown(db *sql.DB) ([]SessionOption, error) {
    query := `
        SELECT s.session_id, s.class_id, s.day_sequence, s.start_at, s.end_at
        FROM class_sessions s
        ORDER BY s.class_id, s.start_at
    `
    rows, err := db.Query(query)
    if err != nil { return nil, err }
    defer rows.Close()

    var opts []SessionOption
    for rows.Next() {
        var s SessionOption
        var day int
        var start, end time.Time
        rows.Scan(&s.ID, &s.ClassID, &day, &start, &end)
        s.DisplayName = fmt.Sprintf("Day %d %s-%s", day, start.Format("15:04"), end.Format("15:04"))
        opts = append(opts, s)
    }
    return opts, nil
}
