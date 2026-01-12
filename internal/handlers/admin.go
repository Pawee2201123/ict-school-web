
package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"example.com/myapp/internal/models"
)

// Make sure you import: "database/sql", "time", "example.com/myapp/internal/models"
func (h *Handler) AdminPage(w http.ResponseWriter, r *http.Request) {
	h.tpl.Render(w, "admin_index.html", nil)
}

// AdminConfig handles GET (show form) and POST (save data)
func (h *Handler) AdminConfig(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        // 1. Process Form Submit
        d1 := r.FormValue("event_day1")
        d2 := r.FormValue("event_day2")
        
        err := models.UpdateEventDates(h.db, d1, d2)
        if err != nil {
            http.Error(w, "Failed to save", http.StatusInternalServerError)
            return
        }
        
        // Redirect back to Admin Home after save
        http.Redirect(w, r, "/admin", http.StatusSeeOther)
        return
    }

    // 2. Render Page (GET)
    dates, err := models.GetEventDates(h.db)
    if err != nil {
        http.Error(w, "DB Error", http.StatusInternalServerError)
        return
    }
    
    // Render the template with current dates
    h.tpl.Render(w, "admin_config_edit.html", dates)
}


func (h *Handler) AdminCreateClass(w http.ResponseWriter, r *http.Request) {
	// GET: Show basic form
	if r.Method == http.MethodGet {
		h.tpl.Render(w, "admin_class_edit.html", nil)
		return
	}

	// POST: Create Class Only
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Form error", http.StatusBadRequest)
		return
	}

	// 1. Basic Parsing
	// (Skipping file upload logic for brevity, keeping it empty/placeholder)
	pdfName := "" 
	
	layout := "2006-01-02T15:04"
	start, _ := time.Parse(layout, r.FormValue("reception_start"))
	end, _ := time.Parse(layout, r.FormValue("reception_end"))

	class := models.Class{
		ClassName:           r.FormValue("class_name"),
		SyllabusPDFURL:      pdfName,
		RoomNumber:          r.FormValue("room_number"),
		RoomName:            r.FormValue("room_name"),
		RegistrationStartAt: start,
		RegistrationEndAt:   end,
	}

	// 2. Instructors
	teachers := []string{r.FormValue("teacher_name_1")}
	if t2 := r.FormValue("teacher_name_2"); t2 != "" {
		teachers = append(teachers, t2)
	}

	// 3. Save Class
	classID, err := models.CreateClassWithInstructors(h.db, class, teachers)
	if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// CHANGE: Redirect to the "Session Management" page for this new class
	// This is the UX improvement: Create -> Immediately start adding sessions
	http.Redirect(w, r, fmt.Sprintf("/admin/classes/detail?id=%d", classID), http.StatusSeeOther)
}

// VIEW: Shows class info + existing sessions + add form
func (h *Handler) AdminClassDetail(w http.ResponseWriter, r *http.Request) {
	// 1. Get ID from URL query ?id=1
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	// 2. Fetch Data
	class, err := models.GetClassByID(h.db, id)
	if err != nil {
		http.Error(w, "Class not found", http.StatusNotFound)
		return
	}
	sessions, err := models.GetSessionsByClassID(h.db, id)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	
	// 3. Prepare Data for Template
	data := map[string]any{
		"Class":    class,
		"Sessions": sessions,
	}
	h.tpl.Render(w, "admin_class_detail.html", data)
}

func combineDateTime(dateStr, timeStr string) (time.Time, error) {
	fullStr := dateStr + " " + timeStr
	return time.Parse("2006-01-02 15:04", fullStr)
}
// ACTION: Adds a single session to a class
func (h *Handler) AdminAddSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	classID, _ := strconv.Atoi(r.FormValue("class_id"))
	daySeq, _ := strconv.Atoi(r.FormValue("day_sequence"))
	capacity, _ := strconv.Atoi(r.FormValue("capacity"))
	
	// Get "Day 1" or "Day 2" date from DB to combine with time
	eventDates, _ := models.GetEventDates(h.db)
	targetDate := eventDates.Day1
	if daySeq == 2 {
		targetDate = eventDates.Day2
	}

	// Combine Date + Time input
	startAt, _ := combineDateTime(targetDate, r.FormValue("start_time"))
	endAt, _ := combineDateTime(targetDate, r.FormValue("end_time"))

	sess := models.Session{
		ClassID:     classID,
		DaySequence: daySeq,
		StartAt:     startAt,
		EndAt:       endAt,
		Capacity:    capacity, // Per session capacity!
	}

	if err := models.CreateSession(h.db, sess); err != nil {
		http.Error(w, "Failed to add session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect back to the detail page to see the new list
	http.Redirect(w, r, fmt.Sprintf("/admin/classes/detail?id=%d", classID), http.StatusSeeOther)
}

// AdminClassList shows all classes so admin can select one to manage
func (h *Handler) AdminClassList(w http.ResponseWriter, r *http.Request) {
	classes, err := models.GetAllClasses(h.db)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	
	h.tpl.Render(w, "admin_class_list.html", classes)
}
