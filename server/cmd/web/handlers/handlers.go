package handlers

import (
	"html/template"
	"log/slog"
)

type Handler struct {
	TemplateCache map[string]*template.Template
	Logger        *slog.Logger
}
