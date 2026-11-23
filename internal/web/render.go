package web

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// Renderer renders full pages: the shared layout plus one page template each.
type Renderer struct {
	pages map[string]*template.Template
}

// NewRenderer parses templates/layout.html with every templates/pages/*.html from fsys.
func NewRenderer(fsys fs.FS) (*Renderer, error) {
	files, err := fs.Glob(fsys, "templates/pages/*.html")
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no page templates found")
	}
	r := &Renderer{pages: make(map[string]*template.Template, len(files))}
	for _, file := range files {
		tmpl, err := template.ParseFS(fsys, "templates/layout.html", file)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", file, err)
		}
		name := strings.TrimSuffix(path.Base(file), ".html")
		r.pages[name] = tmpl
	}
	return r, nil
}

// Render writes the named page, buffering so a template error never yields a half page.
func (r *Renderer) Render(w http.ResponseWriter, name string, data any) error {
	tmpl, ok := r.pages[name]
	if !ok {
		return fmt.Errorf("unknown page template %q", name)
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		return fmt.Errorf("render %s: %w", name, err)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := buf.WriteTo(w)
	return err
}
