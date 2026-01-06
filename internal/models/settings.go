package models

import (
	"database/sql"
)

type EventDates struct {
	Day1 string // Format: "YYYY-MM-DD"
	Day2 string
}

func GetEventDates(db *sql.DB) (EventDates, error) {
	dates := EventDates{}
	
	// We fetch both in one go, or individual queries. 
	// Simple separate queries for clarity:
	err := db.QueryRow("SELECT setting_value FROM system_settings WHERE setting_key='event_date_1'").Scan(&dates.Day1)
	if err != nil { return dates, err }

	err = db.QueryRow("SELECT setting_value FROM system_settings WHERE setting_key='event_date_2'").Scan(&dates.Day2)
	// Day 2 might be empty (optional), so we ignore error if needed, 
    // but for now let's assume it exists.
	return dates, nil
}

func UpdateEventDates(db *sql.DB, d1, d2 string) error {
	_, err := db.Exec(`
		INSERT INTO system_settings (setting_key, setting_value) 
		VALUES ('event_date_1', $1), ('event_date_2', $2)
		ON CONFLICT (setting_key) 
		DO UPDATE SET setting_value = EXCLUDED.setting_value
	`, d1, d2)
	return err
}
