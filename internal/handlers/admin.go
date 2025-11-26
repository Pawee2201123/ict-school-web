
package handlers

import (
	"net/http"
)

func (h *Handler) AdminPage(w http.ResponseWriter, r *http.Request) {
	h.tpl.Render(w, "admin_index.html", nil)
}
