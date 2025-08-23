package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/priyankishorems/transmyaction/api/handlers"
)

func SetupRoutes(h *handlers.Handlers) *echo.Echo {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		// AllowOrigins:     []string{"http://localhost:5173"},
		AllowCredentials: true,
	}))
	// e.Use(IPRateLimit(h))
	e.Use(middleware.RemoveTrailingSlash())
	e.Static("/static", "static")

	e.HideBanner = true
	e.Renderer = handlers.InitTemplates()

	e.GET("/", h.HomeFunc)

	// api := e.Group("/api")
	// {

	// }

	return e
}
