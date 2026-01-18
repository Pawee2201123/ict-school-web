package template

import (
    "html/template"
    "io"
    "log"
    "path/filepath"
    "os"
    "net/http"
)

type Renderer struct {
    templates *template.Template
}

func Load(dir string) *Renderer {
    // 1. Create a Base Template with functions (if needed)
    tmpl := template.New("")
    
    // 2. Walk the directory and parse ALL .html files (Recursive)
    // This finds web/templates/admin/admin_index.html AND web/templates/layout.html
    err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil { return err }
        if !info.IsDir() && filepath.Ext(path) == ".html" {
            _, err = tmpl.ParseFiles(path)
            if err != nil { log.Printf("Error parsing %s: %v", path, err) }
        }
        return nil
    })

    if err != nil {
        log.Println("Template Load Error:", err)
        return nil
    }

    return &Renderer{templates: tmpl}
}

func (t *Renderer) Render(w io.Writer, name string, data any) {
    // 1. SAFE LOOKUP
    // If the template is not found, Lookup returns nil.
    tmpl := t.templates.Lookup(name)
    if tmpl == nil {
        log.Printf("CRITICAL: Template '%s' not found!", name)
        // Log all available templates to help debug
        log.Printf("Available templates: %s", t.templates.DefinedTemplates())
        http.Error(w.(http.ResponseWriter), "Template Missing: "+name, http.StatusInternalServerError)
        return
    }

    // 2. Execute
    err := tmpl.Execute(w, data)
    if err != nil {
        log.Printf("Template Execution Error (%s): %v", name, err)
    }
}
