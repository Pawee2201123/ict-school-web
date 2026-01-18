package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"example.com/myapp/internal/models"
)

func (h *Handler) AdminDataPage(w http.ResponseWriter, r *http.Request) {
	// A. Dropdown Data
	classes, _ := models.GetAllClasses(h.db)
	sessions, _ := models.GetAllSessionsForDropdown(h.db)

	// B. Get Filters from URL
	classID, _ := strconv.Atoi(r.URL.Query().Get("class_id"))
	sessionID, _ := strconv.Atoi(r.URL.Query().Get("session_id"))

	// C. Fetch BOTH Reports using the SAME filters
	// Table 1: Participants
	previewData, _ := models.GetApplicantsReport(h.db, classID, sessionID)
	
	// Table 2: Class Info (Now Dynamic!)
	statuses, _ := models.GetClassStatusReport(h.db, classID, sessionID)

	data := map[string]any{
		"Classes":       classes,
		"Sessions":      sessions,
		"PreviewData":   previewData, // Participants
		"Statuses":      statuses,    // Class Info
		"SelectedClass": classID,
		"SelectedSess":  sessionID,
	}

	h.tpl.Render(w, "admin_data_list.html", data)
}

func (h *Handler) AdminDownloadCSV(w http.ResponseWriter, r *http.Request) {
	classID, _ := strconv.Atoi(r.URL.Query().Get("class_id"))
	sessionID, _ := strconv.Atoi(r.URL.Query().Get("session_id"))

	data, _ := models.GetApplicantsReport(h.db, classID, sessionID)
	
	setCSVHeaders(w, "participants_list.csv")
	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{"ID", "中学生氏名", "保護者", "メール", "登録日時", "授業名", "実施回"})
	for _, row := range data {
		writer.Write([]string{
			strconv.Itoa(row.UserID), row.StudentName, row.GuardianName, row.Email, 
			row.RegDate.Format("2006-01-02 15:04"), row.ClassName, row.SessionTime,
		})
	}
}
func (h *Handler) AdminDownloadClasses(w http.ResponseWriter, r *http.Request) {
	classID, _ := strconv.Atoi(r.URL.Query().Get("class_id"))
	sessionID, _ := strconv.Atoi(r.URL.Query().Get("session_id"))

	data, _ := models.GetClassStatusReport(h.db, classID, sessionID)

	setCSVHeaders(w, "class_info.csv")
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header
	writer.Write([]string{"模擬授業名", "実施回(日時)", "最大受入人数", "現在申込数", "担当教職員", "実施場所"})

	// Rows
	for _, row := range data {
		writer.Write([]string{
			row.ClassName,
			row.SessionTime,
			strconv.Itoa(row.Capacity),
			strconv.Itoa(row.Count), // Added Current Count
			row.Instructors,
			row.RoomName,
		})
	}
}
func setCSVHeaders(w http.ResponseWriter, filename string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Write([]byte{0xEF, 0xBB, 0xBF}) // BOM for Excel
}
