package http

import (
	"embed"
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

// templatesFS holds the server-rendered client pages.
//
//go:embed templates/*.html
var templatesFS embed.FS

// templateRenderer adapts html/template to echo.Renderer.
type templateRenderer struct {
	templates *template.Template
}

func newTemplateRenderer() *templateRenderer {
	return &templateRenderer{
		templates: template.Must(template.ParseFS(templatesFS, "templates/*.html")),
	}
}

func (t *templateRenderer) Render(w io.Writer, name string, data any, _ echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// loginPage serves the sign-in page.
func (h *Handler) loginPage(c echo.Context) error {
	return c.Render(http.StatusOK, "login.html", map[string]any{
		"LoginURL": "/api/users/login",
		"AppPath":  "/app",
	})
}

// appPage serves the authenticated dashboard. The JWT lives in the browser's
// localStorage, so this handler only serves the shell; the page's JS drives the
// API using the stored token.
func (h *Handler) appPage(c echo.Context) error {
	return c.Render(http.StatusOK, "app.html", map[string]any{
		"APIBase":   "/api",
		"LoginPath": "/",
	})
}
