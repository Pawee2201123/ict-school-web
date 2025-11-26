
package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
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
	// data contains session info
	email, _ := data["email"].(string)
	fmt.Fprintf(w, "<h1>ようこそ %s</h1><p><a href=\"/logout\">ログアウト</a></p>", email)
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
	email := r.PostForm.Get("email")
	pw := r.PostForm.Get("password")
	if email == "" || pw == "" || len(pw) < 8 {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(pw)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	_, err = models.CreateUser(h.db, email, hash)
	if errors.Is(err, models.ErrUserExists) {
		http.Error(w, "email already registered", http.StatusConflict)
		return
	}
	if err != nil {
		log.Printf("create user error: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

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
