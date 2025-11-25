
package template

import (
	"html/template"
	"net/http"
	"path/filepath"
)

type Renderer struct {
	t *template.Template
}

func Load(dir string) *Renderer {
	pattern := filepath.Join(dir, "*.html")
	t := template.Must(template.ParseGlob(pattern))
	return &Renderer{t: t}
}

func (r *Renderer) Render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = r.t.ExecuteTemplate(w, name, data)
}
