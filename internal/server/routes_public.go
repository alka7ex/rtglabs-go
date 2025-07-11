package server

import (
	"log"           // Add log import
	"os"            // Add os import
	"path/filepath" // Add path/filepath import
	"strings"

	page "rtglabs-go/cmd/web/page"
	page_auth "rtglabs-go/cmd/web/page/auth"
	handlers "rtglabs-go/internal/handlers/auth"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// registerPublicRoutes registers all publicly accessible routes.
func (s *Server) registerPublicRoutes() {
	// ... (your existing handlers initialization)
	forgotPasswordHandler := handlers.NewForgotPasswordHandler(s.sqlDB, s.emailSender, s.appBaseURL)
	authHandler := handlers.NewAuthHandler(s.sqlDB)

	// --- MODIFIED STATIC FILE SERVER FOR DEVELOPMENT ---
	// Determine the path to your 'assets' directory relative to the executable.
	assetsDir := filepath.Join("cmd", "web", "assets") // Default for running from project root
	if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
		// Fallback for compiled binaries
		ex, err := os.Executable()
		if err != nil {
			log.Fatalf("Failed to get executable path: %v", err)
		}
		exPath := filepath.Dir(ex)
		assetsDir = filepath.Join(exPath, "assets") // Assuming assets are copied next to the binary
		if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
			log.Printf("Warning: Static assets directory not found at default path nor relative to executable: %s", assetsDir)
		}
	}
	log.Printf("Serving static files from disk: %s at URL path /assets", assetsDir)
	s.echo.Static("/assets", assetsDir) // Use echo.Static directly

	// --- Original Static File Server (Comment out or remove) ---
	// fileServer := http.FileServer(http.FS(web.Files))
	// s.echo.GET("/assets/*", echo.WrapHandler(fileServer))

	// Health check and Hello World
	s.echo.GET("/", echo.WrapHandler(templ.Handler(page.HomePage())))

	s.echo.GET("/health", s.healthHandler)

	// Auth routes (public for registration/login)
	s.echo.POST("/api/register", authHandler.StoreRegister)
	s.echo.POST("/api/login", authHandler.StoreLogin)

	s.echo.POST("/api/forgot-password", forgotPasswordHandler.ForgotPassword)
	s.echo.POST("/api/reset-password", forgotPasswordHandler.ResetPassword)
	s.echo.GET("/reset-password", func(c echo.Context) error {
		token := strings.TrimSpace(c.QueryParam("token"))
		scheme := os.Getenv("APP_SCHEME") // example: rtglabsdev
		if scheme == "" {
			scheme = "rtglabs" // fallback default
		}
		return page_auth.ResetPasswordRedirect(scheme, token).Render(c.Request().Context(), c.Response().Writer)
	})
}
