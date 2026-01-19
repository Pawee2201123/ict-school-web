package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// AdminResetPage shows the reset confirmation page
func (h *Handler) AdminResetPage(w http.ResponseWriter, r *http.Request) {
	h.tpl.Render(w, "admin_reset.html", nil)
}

// AdminResetExecute performs the complete system reset
func (h *Handler) AdminResetExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin/reset", http.StatusSeeOther)
		return
	}

	// Verify confirmation keyword
	keyword := r.FormValue("confirm_keyword")
	if keyword != "削除を実行する" {
		http.Error(w, "確認キーワードが正しくありません", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := h.db.Begin()
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		http.Error(w, "システムエラーが発生しました", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // Will be no-op if committed

	// 1. Delete all enrollments
	if _, err := tx.Exec("DELETE FROM session_enrollments"); err != nil {
		log.Printf("Failed to delete enrollments: %v", err)
		http.Error(w, "エラー: 申込データの削除に失敗しました", http.StatusInternalServerError)
		return
	}

	// 2. Delete all sessions
	if _, err := tx.Exec("DELETE FROM class_sessions"); err != nil {
		log.Printf("Failed to delete sessions: %v", err)
		http.Error(w, "エラー: セッションデータの削除に失敗しました", http.StatusInternalServerError)
		return
	}

	// 3. Delete all class-instructor relationships
	if _, err := tx.Exec("DELETE FROM class_instructors"); err != nil {
		log.Printf("Failed to delete class_instructors: %v", err)
		http.Error(w, "エラー: 授業-講師関係の削除に失敗しました", http.StatusInternalServerError)
		return
	}

	// 4. Delete all classes
	if _, err := tx.Exec("DELETE FROM classes"); err != nil {
		log.Printf("Failed to delete classes: %v", err)
		http.Error(w, "エラー: 授業データの削除に失敗しました", http.StatusInternalServerError)
		return
	}

	// 5. Delete all instructors
	if _, err := tx.Exec("DELETE FROM instructors"); err != nil {
		log.Printf("Failed to delete instructors: %v", err)
		http.Error(w, "エラー: 講師データの削除に失敗しました", http.StatusInternalServerError)
		return
	}

	// 6. Delete all user profiles (students only - admin profiles don't exist typically)
	if _, err := tx.Exec("DELETE FROM user_profiles WHERE user_id IN (SELECT id FROM users WHERE is_admin = FALSE)"); err != nil {
		log.Printf("Failed to delete user profiles: %v", err)
		http.Error(w, "エラー: ユーザープロファイルの削除に失敗しました", http.StatusInternalServerError)
		return
	}

	// 7. Delete all non-admin users
	if _, err := tx.Exec("DELETE FROM users WHERE is_admin = FALSE"); err != nil {
		log.Printf("Failed to delete users: %v", err)
		http.Error(w, "エラー: ユーザーデータの削除に失敗しました", http.StatusInternalServerError)
		return
	}

	// 8. Reset system settings to defaults
	if _, err := tx.Exec("DELETE FROM system_settings"); err != nil {
		log.Printf("Failed to delete system_settings: %v", err)
		http.Error(w, "エラー: システム設定の削除に失敗しました", http.StatusInternalServerError)
		return
	}

	// Insert default settings
	if _, err := tx.Exec(`
		INSERT INTO system_settings (setting_key, setting_value) VALUES
		('event_date_1', '2025-08-01'),
		('event_date_2', '2025-08-02')
	`); err != nil {
		log.Printf("Failed to insert default settings: %v", err)
		http.Error(w, "エラー: デフォルト設定の追加に失敗しました", http.StatusInternalServerError)
		return
	}

	// Commit database changes
	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		http.Error(w, "エラー: データベースの変更をコミットできませんでした", http.StatusInternalServerError)
		return
	}

	// 9. Delete all files in uploads directory
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./web/static/uploads"
	}

	if err := clearUploadDirectory(uploadDir); err != nil {
		log.Printf("Warning: Failed to clear upload directory: %v", err)
		// Don't fail the entire operation if file deletion fails
	}

	log.Println("System reset completed successfully")

	// Redirect to admin home with success message
	// You could also render a success page here
	http.Redirect(w, r, "/admin?reset=success", http.StatusSeeOther)
}

// clearUploadDirectory removes all files from the uploads directory
func clearUploadDirectory(uploadDir string) error {
	// Check if directory exists
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		// Directory doesn't exist, nothing to clear
		return nil
	}

	// Read all files in directory
	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		return err
	}

	// Delete each file
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}

		filePath := filepath.Join(uploadDir, entry.Name())
		if err := os.Remove(filePath); err != nil {
			log.Printf("Warning: Failed to delete file %s: %v", filePath, err)
			// Continue deleting other files even if one fails
		}
	}

	log.Printf("Cleared %d files from upload directory", len(entries))
	return nil
}
