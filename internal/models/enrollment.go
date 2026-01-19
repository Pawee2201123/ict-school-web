package models

import (
	"database/sql"
	"errors"
	"time"
)

// Define errors we can check for later
var (
	ErrAlreadyEnrolled    = errors.New("user is already enrolled in this session")
	ErrSessionFull        = errors.New("class session is full")
	ErrDayLimitExceeded   = errors.New("cannot enroll in more than 2 classes on the same day")
	ErrTotalLimitExceeded = errors.New("cannot enroll in more than 3 classes total")
)

// EnrolledSession represents a class the user has joined (for MyPage)
type EnrolledSession struct {
    SessionID   int
    ClassName   string
    StartAt     time.Time
    EndAt       time.Time
}

// EnrollUser adds a student to a class session
func EnrollUser(db *sql.DB, sessionID, userID int) error {
	// 1. Check Capacity (Simplified)
	var current, cap int
	err := db.QueryRow("SELECT current_enrolled_count, capacity FROM class_sessions WHERE session_id = $1", sessionID).Scan(&current, &cap)
	if err != nil {
		return err
	}
	if current >= cap {
		return ErrSessionFull
	}

	// 2. Insert Enrollment
    // We assume you have the user_id, so we find the profile ID via subquery
	const query = `
		INSERT INTO session_enrollments (session_id, user_profile_id)
		SELECT $1, id FROM user_profiles WHERE user_id = $2
		ON CONFLICT (session_id, user_profile_id) DO NOTHING
		RETURNING enrollment_id
	`
	
	var enrollmentID int
	err = db.QueryRow(query, sessionID, userID).Scan(&enrollmentID)
	
	if err == sql.ErrNoRows {
		return ErrAlreadyEnrolled
	}
    
    // 3. Update the counter in class_sessions
    if err == nil {
        _, _ = db.Exec("UPDATE class_sessions SET current_enrolled_count = current_enrolled_count + 1 WHERE session_id = $1", sessionID)
    }
    
	return err
}

// HasUserJoined checks if a user is already in a session
func HasUserJoined(db *sql.DB, sessionID, userID int) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM session_enrollments se
			JOIN user_profiles up ON se.user_profile_id = up.id
			WHERE se.session_id = $1 AND up.user_id = $2
		)
	`
	err := db.QueryRow(query, sessionID, userID).Scan(&exists)
	return exists, err
}

// GetUserEnrollments fetches the list of classes a student has joined
func GetUserEnrollments(db *sql.DB, userID int) ([]EnrolledSession, error) {
    query := `
        SELECT cs.session_id, c.class_name, cs.start_at, cs.end_at
        FROM session_enrollments se
        JOIN class_sessions cs ON se.session_id = cs.session_id
        JOIN classes c ON cs.class_id = c.class_id
        JOIN user_profiles up ON se.user_profile_id = up.id
        WHERE up.user_id = $1
        ORDER BY cs.start_at DESC
    `
    rows, err := db.Query(query, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var sessions []EnrolledSession
    for rows.Next() {
        var s EnrolledSession
        if err := rows.Scan(&s.SessionID, &s.ClassName, &s.StartAt, &s.EndAt); err != nil {
            return nil, err
        }
        sessions = append(sessions, s)
    }
    return sessions, nil
}

// CheckEnrollmentLimits verifies that the user hasn't exceeded enrollment limits
// Rules: Max 2 classes per day, Max 3 classes total
func CheckEnrollmentLimits(db *sql.DB, userID, newSessionID int) error {
	// Get the day_sequence of the session the user wants to enroll in
	var newDaySequence int
	err := db.QueryRow(`
		SELECT day_sequence
		FROM class_sessions
		WHERE session_id = $1
	`, newSessionID).Scan(&newDaySequence)
	if err != nil {
		return err
	}

	// Count enrollments per day and total
	query := `
		SELECT
			cs.day_sequence,
			COUNT(*) as count
		FROM session_enrollments se
		JOIN class_sessions cs ON se.session_id = cs.session_id
		JOIN user_profiles up ON se.user_profile_id = up.id
		WHERE up.user_id = $1
		GROUP BY cs.day_sequence
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	dayCounts := make(map[int]int)
	totalCount := 0

	for rows.Next() {
		var daySeq, count int
		if err := rows.Scan(&daySeq, &count); err != nil {
			return err
		}
		dayCounts[daySeq] = count
		totalCount += count
	}

	// Check total limit (max 3 classes)
	if totalCount >= 3 {
		return ErrTotalLimitExceeded
	}

	// Check day limit (max 2 classes per day)
	if dayCounts[newDaySequence] >= 2 {
		return ErrDayLimitExceeded
	}

	return nil
}
