
package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"example.com/myapp/internal/auth"
	"example.com/myapp/internal/models"
	"example.com/myapp/internal/template"
	"example.com/myapp/internal/config"

	"crypto/rand"
	"encoding/hex"
)


type Handler struct {
	db     *sql.DB
	tpl    *template.Renderer
	cfg    config.Config
	sess   *auth.Session
}

func New(db *sql.DB, tpl *template.Renderer, cfg config.Config) *Handler {
	// if cookie keys not provided, generate ephemeral keys (dev only)
	hash := cfg.CookieHash
	if hash == "" {
		hash = string(authRandom(32))
	}
	block := cfg.CookieBlock

	return &Handler{
		db:   db,
		tpl:  tpl,
		cfg:  cfg,
		sess: auth.NewSecureCookie(hash, block),
	}
}

// Home - protected
func (h *Handler) Home(w http.ResponseWriter, r *http.Request, data map[string]any) {
    // 1) extract user_id from session data (cookie-decoded map)
    var userID int
    switch v := data["user_id"].(type) {
    case int:
        userID = v
    case float64:
        userID = int(v) // securecookie/json may decode numbers as float64
    default:
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    // 2) fetch profile from DB
    profile, err := models.GetUserProfile(h.db, userID)
    if err != nil {
        http.Error(w, "server error", http.StatusInternalServerError)
        return
    }

    // 3) prepare view data (avoid modifying original map if you prefer)
    view := map[string]any{
        "Email":       data["email"],
        "StudentName": "",
        "SchoolName":  "",
        "Grade":       "",
        "GuardianName":    "",
    }
    if profile != nil {
        if profile.StudentName.Valid {
            view["StudentName"] = profile.StudentName.String
        }
        if profile.SchoolName.Valid {
            view["SchoolName"] = profile.SchoolName.String

        }
        if profile.Grade.Valid {
            view["Grade"] = profile.Grade.String
        }
        if profile.GuardianName.Valid {
            view["GuardianName"] = profile.GuardianName.String

        }
    }

    // 4) render template with view data
    h.tpl.Render(w, "mypage.html", view)
}
// Signup: GET shows form; POST creates user
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        h.tpl.Render(w, "signup.html", nil)
        return
    }

    if err := r.ParseForm(); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }

    email := r.PostForm.Get("Email")
    pw := r.PostForm.Get("password")
    studentName := r.PostForm.Get("student_name")
    schoolName := r.PostForm.Get("school_name")
    grade := r.PostForm.Get("grade")
    guardianName := r.PostForm.Get("guardian_name")

    if email == "" || pw == "" || studentName == "" || schoolName == "" || grade == "" || guardianName == "" {
        http.Error(w, "invalid input", http.StatusBadRequest)
        return
    }

    hashed, err := auth.HashPassword(pw)
    if err != nil {
        http.Error(w, "server error", http.StatusInternalServerError)
        return
    }

    // insert into users table
    var userID int
    err = h.db.QueryRow(
        `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
        email, hashed,
    ).Scan(&userID)
    if err != nil {
        http.Error(w, "email already exists", http.StatusConflict)
        return
    }

    // insert into user_profiles table
    _, err = h.db.Exec(`
        INSERT INTO user_profiles
        (user_id, student_name, school_name, grade, guardian_name)
        VALUES ($1, $2, $3, $4, $5)
    `, userID, studentName, schoolName, grade, guardianName)
    if err != nil {
        http.Error(w, "server error", http.StatusInternalServerError)
        return
    }

    // redirect to login page after successful signup
    http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Login: GET shows form; POST authenticates
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.tpl.Render(w, "login.html", nil)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	email := r.PostForm.Get("email")
	pw := r.PostForm.Get("password")
	if email == "" || pw == "" {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	u, err := models.GetUserByEmail(h.db, email)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := auth.CompareHash(u.PasswordHash, pw); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// determine admin flag from DB (ensure models.User or DB has this column)
	var isAdmin bool
	if err := h.db.QueryRow("SELECT is_admin FROM users WHERE id = $1", u.ID).Scan(&isAdmin); err != nil {
		// If this fails, treat as server error rather than letting login succeed silently
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	// create session payload
	payload := map[string]any{
		"user_id": u.ID,
		"email":   u.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	encoded, err := h.sess.Secure.Encode(h.sess.Key, payload)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	c := &http.Cookie{
		Name:     h.sess.Key,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
		// Secure: true, // enable in production with HTTPS
	}
	http.SetCookie(w, c)

	// redirect based on role
	if isAdmin {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// Logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request, _ map[string]any) {
	c := &http.Cookie{
		Name:     h.sess.Key,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	}
	http.SetCookie(w, c)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}


// authRandom generates a random byte slice of length n, returns hex bytes
func authRandom(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return []byte(hex.EncodeToString(b))
}
