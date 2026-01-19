package handlers

import (
	"log"
	"net/http"
	"strconv"

	"example.com/myapp/internal/email"
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

    // --- GET: Fetch data needed for both GET and error cases ---
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
        "Error":   "",
    }

    // --- POST: PROCESS APPLICATION ---
    if r.Method == http.MethodPost {
        var errorMsg string

        // Check enrollment limits before allowing enrollment
        if err := models.CheckEnrollmentLimits(h.db, userID, sessID); err != nil {
            if err == models.ErrDayLimitExceeded {
                errorMsg = "申込数の上限を超えています。同じ日に申し込める授業は2つまでです。"
            } else if err == models.ErrTotalLimitExceeded {
                errorMsg = "申込数の上限を超えています。申し込める授業は全体で3つまでです。"
            } else {
                errorMsg = "エラーが発生しました: " + err.Error()
            }
            viewData["Error"] = errorMsg
            h.tpl.Render(w, "application.html", viewData)
            return
        }

        // Proceed with enrollment
        err := models.EnrollUser(h.db, sessID, userID)
        if err != nil {
            // Handle specific errors for better UX
            if err == models.ErrAlreadyEnrolled {
                errorMsg = "この授業には既に申し込んでいます。"
            } else if err == models.ErrSessionFull {
                errorMsg = "この授業は満席です。"
            } else {
                errorMsg = "申込に失敗しました: " + err.Error()
            }
            viewData["Error"] = errorMsg
            h.tpl.Render(w, "application.html", viewData)
            return
        }

        // Send confirmation email (asynchronously to avoid blocking)
        go func() {
            if err := h.sendEnrollmentEmail(userID, sessID, data["email"].(string)); err != nil {
                log.Printf("Failed to send enrollment email to user %d: %v", userID, err)
            }
        }()

        // Success! Redirect to MyPage
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    // --- GET: SHOW CONFIRMATION PAGE ---
    h.tpl.Render(w, "application.html", viewData)
}

// sendEnrollmentEmail sends a confirmation email after successful enrollment
func (h *Handler) sendEnrollmentEmail(userID, sessionID int, userEmail string) error {
	// Get session details
	sessionDetail, err := models.GetSessionDetail(h.db, sessionID)
	if err != nil {
		return err
	}

	// Get user profile
	profile, err := models.GetUserProfile(h.db, userID)
	if err != nil {
		return err
	}

	// Extract student name with fallback
	studentName := "Student"
	if profile != nil && profile.StudentName.Valid {
		studentName = profile.StudentName.String
	}

	// Prepare email data
	emailData := email.EnrollmentData{
		StudentName: studentName,
		ClassName:   sessionDetail.ClassName,
		RoomNumber:  sessionDetail.RoomNumber,
		RoomName:    sessionDetail.RoomName,
		TeacherName: sessionDetail.TeacherName,
		StartAt:     sessionDetail.StartAt,
		EndAt:       sessionDetail.EndAt,
	}

	// Generate email content
	subject := email.GetEnrollmentSubject()
	body := email.GenerateEnrollmentConfirmation(emailData)

	// Send email
	return h.mailer.Send(userEmail, subject, body)
}
