package server

import (
	"net/http"

	"rtglabs-go/cmd/web"
	handlers "rtglabs-go/internal/handlers/auth"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// registerPublicRoutes registers all publicly accessible routes.
func (s *Server) registerPublicRoutes() {
	// Initialize Ent Client here if needed for handlers

	forgotPasswordHandler := handlers.NewForgotPasswordHandler(s.entClient, s.emailSender, s.appBaseURL)

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
	authHandler := handlers.NewAuthHandler(s.entClient)
	s.echo.POST("/api/register", authHandler.StoreRegister) // Use the new name
	s.echo.POST("/api/login", authHandler.StoreLogin)       // Use the new name

	s.echo.POST("/api/forgot-password", forgotPasswordHandler.ForgotPassword)
	s.echo.POST("/api/reset-password", forgotPasswordHandler.ResetPassword)
}
