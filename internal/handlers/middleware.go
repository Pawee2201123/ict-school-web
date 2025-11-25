
package handlers

import (
	"net/http"
	"strconv"
	"time"
)

// RequireLogin wraps a handler that expects decoded session data map[string]any
func (h *Handler) RequireLogin(next func(http.ResponseWriter, *http.Request, map[string]any)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// allow public routes
		if r.URL.Path == "/signup" || r.URL.Path == "/login" {
			// forward to handlers
			if r.URL.Path == "/signup" {
				h.Signup(w, r)
				return
			}
			h.Login(w, r)
			return
		}

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
		// check expiry
		expVal, ok := data["exp"]
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
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
		if time.Now().Unix() > exp {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next(w, r, data)
	}
}
