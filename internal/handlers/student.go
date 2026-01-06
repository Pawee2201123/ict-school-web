package handlers

import (
	"net/http"
	"strconv"

	"example.com/myapp/internal/models"
)

// ClassView is a helper struct just for the Template
// It combines the static Class data with the dynamic Session list
type ClassView struct {
	Class    models.Class
	Sessions []SessionView
}

type SessionView struct {
	Session        models.Session
	IsFull         bool
	IsEnrolled     bool
	ButtonLabel    string // e.g. "受付中" (Open), "満席" (Full), "申込済" (Joined)
	ButtonDisabled bool
}

// StudentLessonList handles the main catalog page
func (h *Handler) StudentLessonList(w http.ResponseWriter, r *http.Request) {
	// 1. Get current User ID from Context (to check "Already Joined")
	// If not logged in, userID will be 0, which is fine (just shows all open)
	userID := 0
	if val := r.Context().Value(sessionKey); val != nil {
		data := val.(map[string]any)
		if uid, ok := data["user_id"].(int); ok {
			userID = uid
		} else if uidFloat, ok := data["user_id"].(float64); ok {
			userID = int(uidFloat)
		}
	}

	// 2. Fetch All Classes
	classes, err := models.GetAllClasses(h.db)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	// 3. Build the View Data
	var viewData []ClassView

	for _, c := range classes {
		// Fetch sessions for this class
		sessions, err := models.GetSessionsByClassID(h.db, c.ID)
		if err != nil {
			continue
		}

		var sessViews []SessionView
		for _, s := range sessions {
			// A. Check Capacity
			isFull := s.CurrentEnrolledCount >= s.Capacity

			// B. Check if User is Enrolled (Only if logged in)
			isEnrolled := false
			if userID > 0 {
				isEnrolled, _ = models.HasUserJoined(h.db, s.ID, userID)
			}

			// C. Determine Button State
			label := "受付中" // Open
			disabled := false

			if isEnrolled {
				label = "申込済" // Already Joined
				disabled = true
			} else if isFull {
				label = "満席" // Full
				disabled = true
			} else if s.CurrentEnrolledCount >= s.Capacity-5 {
				label = "残りわずか" // Low stock
			}

			// Format Time nicely for display (Optional)
			// You can do this in HTML too, but Go is safer

			sessViews = append(sessViews, SessionView{
				Session:        s,
				IsFull:         isFull,
				IsEnrolled:     isEnrolled,
				ButtonLabel:    label,
				ButtonDisabled: disabled,
			})
		}

		viewData = append(viewData, ClassView{
			Class:    c,
			Sessions: sessViews,
		})
	}

	// 4. Render
	// We pass 'viewData' which contains everything the HTML needs
	h.tpl.Render(w, "lesson_list.html", viewData)
}
// Application Page: 
// GET: Shows confirmation form
// POST: Executes "Join"
func (h *Handler) StudentApplication(w http.ResponseWriter, r *http.Request) {
    // 1. Get Session ID
    sessIDStr := r.FormValue("session_id")
    if sessIDStr == "" {
        sessIDStr = r.URL.Query().Get("session_id")
    }
    sessID, _ := strconv.Atoi(sessIDStr)

    // 2. Get User ID (SAFELY)
    // First, check if session data exists at all
    ctxVal := r.Context().Value(sessionKey)
    if ctxVal == nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }
    
    data, ok := ctxVal.(map[string]any)
    if !ok {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    // --- FIX IS HERE: Type Switch ---
    var userID int
    switch v := data["user_id"].(type) {
    case int:
        userID = v            // It's already an int (Clean!)
    case float64:
        userID = int(v)       // It's a float (JSON style), convert it
    default:
        // user_id is missing or weird type -> Crash prevented
        http.Error(w, "Session Error: Invalid User ID", http.StatusInternalServerError)
        return
    }

    // --- POST: PROCESS APPLICATION ---
    if r.Method == http.MethodPost {
        err := models.EnrollUser(h.db, sessID, userID)
        if err != nil {
            // Handle specific errors for better UX
            if err.Error() == "user is already enrolled in this session" {
                 http.Error(w, "You have already joined this class.", http.StatusConflict)
            } else if err.Error() == "class session is full" {
                 http.Error(w, "Class is full.", http.StatusConflict)
            } else {
                 http.Error(w, "Enrollment failed: "+err.Error(), http.StatusInternalServerError)
            }
            return
        }
        // Success! Redirect to MyPage
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    // --- GET: SHOW CONFIRMATION PAGE ---
    detail, err := models.GetSessionDetail(h.db, sessID)
    if err != nil {
        http.Error(w, "Session not found", http.StatusNotFound)
        return
    }

    profile, err := models.GetUserProfile(h.db, userID)
    if err != nil {
        http.Error(w, "Profile not found", http.StatusInternalServerError)
        return
    }

    viewData := map[string]any{
        "Session": detail,
        "User":    profile,
        "Email":   data["email"],
    }

    h.tpl.Render(w, "application.html", viewData)
}
