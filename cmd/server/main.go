package main

import (
	"log"
	"net/http"

	"example.com/myapp/internal/config"
	"example.com/myapp/internal/database"
	"example.com/myapp/internal/handlers"
	"example.com/myapp/internal/template"
)

func main() {
	cfg := config.Load()

	db := database.Connect(cfg.DB_DSN)
	defer db.Close()

	tpl := template.Load("web/templates")

	h := handlers.New(db, tpl, cfg)

	mux := http.NewServeMux()
	// public
	mux.HandleFunc("/signup", h.Signup)
	mux.HandleFunc("/login", h.Login)
	// static (optional)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// protected
	mux.HandleFunc("/", h.RequireLogin(h.Home))
	mux.HandleFunc("/logout", h.RequireLogin(h.Logout))

	addr := cfg.ListenAddr
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}
