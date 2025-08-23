package handlers

import (
	"io"
	"net/http"

	"html/template"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/priyankishorems/transmyaction/utils"
)

type Cake map[string]interface{}
type Handlers struct {
	Config   utils.Config
	Validate validator.Validate
	Utils    utils.Utilities
	// Data     data.Models
}

type Template struct {
	Templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.Templates.ExecuteTemplate(w, name, data)
}

func InitTemplates() *Template {
	tmpl := template.Must(template.New("").ParseGlob("templates/*.html"))
	return &Template{Templates: tmpl}
}

func (h *Handlers) HomeFunc(c echo.Context) error {
	data := map[string]interface{}{
		"title": "TransMyAction",
	}

	return c.Render(http.StatusOK, "layout.html", data)
}
