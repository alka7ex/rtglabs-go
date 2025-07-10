package server

import (
	"net/http"

	web "rtglabs-go/cmd/web"
	page "rtglabs-go/cmd/web/page"
	handlers "rtglabs-go/internal/handlers/auth" // Assuming handlers are now in 'internal/handlers/auth'

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// registerPublicRoutes registers all publicly accessible routes.
func (s *Server) registerPublicRoutes() {
	// Initialize handlers with s.sqlDB instead of s.entClient
	forgotPasswordHandler := handlers.NewForgotPasswordHandler(s.sqlDB, s.emailSender, s.appBaseURL)

	// Public static file server
	fileServer := http.FileServer(http.FS(web.Files))
	s.echo.GET("/assets/*", echo.WrapHandler(fileServer))

	// Web templ examples
	// s.echo.GET("/web", echo.WrapHandler(templ.Handler(web.HelloForm())))
	s.echo.POST("/", echo.WrapHandler(http.HandlerFunc(web.HelloWebHandler)))

	// Health check and Hello World
	s.echo.GET("/", echo.WrapHandler(templ.Handler(page.HomePage())))

	s.echo.GET("/health", s.healthHandler)

	// Auth routes (public for registration/login)
	authHandler := handlers.NewAuthHandler(s.sqlDB) // Pass s.sqlDB instead of s.entClient
	s.echo.POST("/api/register", authHandler.StoreRegister)
	s.echo.POST("/api/login", authHandler.StoreLogin)

	s.echo.POST("/api/forgot-password", forgotPasswordHandler.ForgotPassword)
	s.echo.POST("/api/reset-password", forgotPasswordHandler.ResetPassword)
}
