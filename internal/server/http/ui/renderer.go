package ui

import (
	"html/template"
	"net/http"

	"github.com/turtacn/QuantaID/internal/server/http/middleware"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/web"
)

type TemplateData struct {
	CSRFToken string
	User      *types.User
	Error     string
	Data      interface{}
}

type Renderer struct {
	templates *template.Template
}

func NewRenderer() (*Renderer, error) {
	// Parse ONLY shared layout templates.
	// Specific pages will be parsed on-demand in Render to ensure correct block overriding.
	t, err := template.ParseFS(web.TemplateFS, "templates/layout.html")
	if err != nil {
		return nil, err
	}
	return &Renderer{templates: t}, nil
}

func (r *Renderer) Render(w http.ResponseWriter, req *http.Request, tmplName string, data interface{}) {
	templateData := TemplateData{
		CSRFToken: middleware.GetCSRFToken(req),
		Data:      data,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	t, err := r.templates.Clone()
	if err != nil {
		http.Error(w, "failed to clone templates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	pattern := "templates/" + tmplName
	_, err = t.ParseFS(web.TemplateFS, pattern)
	if err != nil {
		http.Error(w, "failed to parse template file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	execName := tmplName
	for i := len(tmplName) - 1; i >= 0; i-- {
		if tmplName[i] == '/' {
			execName = tmplName[i+1:]
			break
		}
	}

	err = t.ExecuteTemplate(w, execName, templateData)
	if err != nil {
		http.Error(w, "failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
