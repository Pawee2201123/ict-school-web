package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// 1. Define the Context Key (Private to this file/package)
type contextKey string
const sessionKey contextKey = "session_data"

// ---------------------------------------------------------
// Middleware 1: Require Login (The Producer)
// Decodes cookie -> Saves to Context
// ---------------------------------------------------------
func (h *Handler) RequireLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// A. Get Cookie
		c, err := r.Cookie(h.sess.Key)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// B. Decode Data
		var data map[string]any
		if err := h.sess.Secure.Decode(h.sess.Key, c.Value, &data); err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// C. Check Expiry
		if !isSessionValid(data) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// D. Save to Context
		ctx := context.WithValue(r.Context(), sessionKey, data)
		next(w, r.WithContext(ctx))
	}
}

// ---------------------------------------------------------
// Middleware 2: Require Admin (The Consumer)
// Reads Context -> Checks Database -> Allows/Blocks
// ---------------------------------------------------------
func (h *Handler) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// A. Retrieve data from Context (Must be logged in first!)
		data, ok := r.Context().Value(sessionKey).(map[string]any)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// B. Extract User ID
		var uid int
		switch v := data["user_id"].(type) {
		case float64:
			uid = int(v)
		case int:
			uid = v
		default:
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// C. Check Admin Status in DB
		var isAdmin bool
		err := h.db.QueryRow("SELECT is_admin FROM users WHERE id=$1", uid).Scan(&isAdmin)
		if err != nil || !isAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// D. Pass through
		next(w, r)
	}
}

// ---------------------------------------------------------
// Helper Functions
// ---------------------------------------------------------
func isSessionValid(data map[string]any) bool {
	expVal, ok := data["exp"]
	if !ok {
		return false
	}
	var exp int64
	switch v := expVal.(type) {
	case int64:
		exp = v
	case float64:
		exp = int64(v)
	case string:
		exp, _ = strconv.ParseInt(v, 10, 64)
	}
	return time.Now().Unix() < exp
}
