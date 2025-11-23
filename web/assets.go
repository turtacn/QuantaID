package web

import "embed"

// TemplateFS holds the embedded HTML templates from the 'templates' directory.
//go:embed templates/*.html templates/*/*.html
var TemplateFS embed.FS

// StaticFS holds the embedded static assets (JS, CSS) from the 'static' directory.
//go:embed all:static
var StaticFS embed.FS
