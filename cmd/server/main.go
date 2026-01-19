package main

import (
	"log"
	"net/http"
	"os"

	"example.com/myapp/internal/config"
	"example.com/myapp/internal/database"
	"example.com/myapp/internal/handlers"
	"example.com/myapp/internal/template"
)

func main() {
	cfg := config.Load()

	db := database.Connect(cfg.DB_DSN)
	defer db.Close()

	// set admin 
	if err := handlers.EnsureAdmin(db); err != nil {
		log.Fatal("Failed to ensure admin:", err)
	}

	tpl := template.Load("web/templates")

	h := handlers.New(db, tpl, cfg)

	mux := http.NewServeMux()
	// public
	mux.HandleFunc("/signup", h.Signup)
	mux.HandleFunc("/login", h.Login)

	// 1. Determine where uploaded files live
	// Default to local folder for development
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./web/static/uploads"
	}
	// Ensure the directory exists
	os.MkdirAll(uploadDir, 0755)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))

	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))

	// protected
	mux.HandleFunc("/", h.RequireLogin(h.Home))
	mux.HandleFunc("/logout", h.RequireLogin(h.Logout))

	mux.HandleFunc("/lesson", h.RequireLogin(h.StudentLessonList))

	mux.HandleFunc("/application", h.RequireLogin(h.StudentApplication))


	protectAdmin := func(next http.HandlerFunc) http.HandlerFunc {
		return h.RequireLogin(h.RequireAdmin(next))
	}


	mux.HandleFunc("/admin", protectAdmin(h.AdminPage))

	mux.HandleFunc("/admin/config", protectAdmin(h.AdminConfig))

	mux.HandleFunc("/admin/classes/new", protectAdmin(h.AdminCreateClass))

	// 1. View Detail Page
	mux.HandleFunc("/admin/classes/detail", protectAdmin(h.AdminClassDetail))

	// 2. Action: Add Session

	mux.HandleFunc("/admin/sessions/add", protectAdmin(h.AdminAddSession))

	mux.HandleFunc("/admin/classes", protectAdmin(h.AdminClassList))

	mux.HandleFunc("/admin/data", protectAdmin(h.AdminDataPage))

	mux.HandleFunc("/admin/data/download", protectAdmin(h.AdminDownloadCSV))

	mux.HandleFunc("/admin/data/download/classes", protectAdmin(h.AdminDownloadClasses))

	mux.HandleFunc("/admin/reset", protectAdmin(h.AdminResetPage))

	mux.HandleFunc("/admin/reset/execute", protectAdmin(h.AdminResetExecute))


	addr := cfg.ListenAddr
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}
