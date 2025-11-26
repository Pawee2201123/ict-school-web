
package handlers

import (
	"net/http"
)

func (h *Handler) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// First ensure logged in (reuse your session logic)
		c, err := r.Cookie(h.sess.Key)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		var data map[string]any
		if err := h.sess.Secure.Decode(h.sess.Key, c.Value, &data); err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Extract user id
		uidF := data["user_id"]
		var uid int
		switch v := uidF.(type) {
		case float64:
			uid = int(v)
		case int:
			uid = v
		default:
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// check is_admin = true
		var isAdmin bool
		err = h.db.QueryRow("SELECT is_admin FROM users WHERE id=$1", uid).Scan(&isAdmin)
		if err != nil || !isAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}
