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
	// Parse all templates; names will be "layout.html", "login.html", etc.
	t, err := template.ParseFS(web.TemplateFS, "templates/*.html")
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

	// **IMPORTANT: Use tmplName exactly as "login.html"**
	err := r.templates.ExecuteTemplate(w, tmplName, templateData)
	if err != nil {
		http.Error(w, "failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
