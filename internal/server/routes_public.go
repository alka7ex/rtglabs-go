package server

import (
	"net/http"

	"rtglabs-go/cmd/web"
	"rtglabs-go/internal/database" // Import database for entClient if needed
	"rtglabs-go/internal/handlers/auth"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// registerPublicRoutes registers all publicly accessible routes.
func (s *Server) registerPublicRoutes() {
	// Initialize Ent Client here if needed for handlers
	entClient := database.NewEntClient() // Assuming you need this for auth handler

	// Public static file server
	fileServer := http.FileServer(http.FS(web.Files))
	s.echo.GET("/assets/*", echo.WrapHandler(fileServer))

	// Web templ examples
	s.echo.GET("/web", echo.WrapHandler(templ.Handler(web.HelloForm())))
	s.echo.POST("/hello", echo.WrapHandler(http.HandlerFunc(web.HelloWebHandler)))

	// Health check and Hello World
	s.echo.GET("/", s.HelloWorldHandler)
	s.echo.GET("/health", s.healthHandler)

	// Auth routes (public for registration/login)
	authHandler := handlers.NewAuthHandler(entClient)
	s.echo.POST("/auth/register", authHandler.StoreRegister) // Use the new name
	s.echo.POST("/auth/login", authHandler.StoreLogin)       // Use the new name
}
