package handlers

import (
	"html/template"
	"log/slog"

	"github.com/kelaaditya/zomato-weather-union/server/internal/models"
)

type Handler struct {
	TemplateCache map[string]*template.Template
	Logger        *slog.Logger
	Models        *models.Models
}
