package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload" // Automatically loads .env file

	"rtglabs-go/config"          // <--- KEEP this import if 'appConfig' field is intended, though it won't be initialized here. Otherwise, remove if not used.
	"rtglabs-go/config/database" // <--- ADD THIS IMPORT for database.NewEntClient()
	"rtglabs-go/ent"             // <--- Ensure this import is present
	mail "rtglabs-go/provider"   // <--- THIS IS THE CORRECT IMPORT for EmailSender

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	// No need for explicit _ "github.com/go-sql-driver/mysql" if using SQLite and database/ent.go handles driver import.
	// _ "github.com/mattn/go-sqlite3" // You might need this here if database/ent.go doesn't exclusively handle it, but it's usually better to keep driver imports with ent.Open.
)

// Server holds the server configuration and dependencies.
type Server struct {
	port        int
	db          database.Service
	echo        *echo.Echo
	logger      *zap.Logger
	emailSender mail.EmailSender // <--- Corrected type: use 'provider.EmailSender'
	appBaseURL  string
	appConfig   *config.AppConfig // This field remains, but will be nil as NewServer() doesn't take config.
	entClient   *ent.Client       // Your Ent database client
}

// NewServer initializes and returns a new HTTP server.
func NewServer() *http.Server {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		log.Fatal("PORT environment variable is not set")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid PORT value: %s", portStr)
	}

	log.Printf("APP ENV: %s", os.Getenv("APP_ENV"))
	log.Printf("PORT: %s", os.Getenv("PORT"))
	log.Printf("APP_BASE_URL: %s", os.Getenv("APP_BASE_URL"))

	// --- ADDITION: Initialize EntClient, EmailSender, and AppBaseURL ---

	// 1. Initialize Ent Client using your database.NewEntClient()
	entClient := database.NewEntClient() // <--- CALL YOUR NEW ENT CLIENT HERE

	// 2. Initialize EmailSender
	emailSender := mail.NewSMTPEmailSender( // <--- Corrected package: use 'provider.NewSMTPEmailSender'
		os.Getenv("SMTP_HOST"),
		os.Getenv("SMTP_PORT"),
		os.Getenv("SMTP_USERNAME"),
		os.Getenv("SMTP_PASSWORD"),
		os.Getenv("SMTP_FROM_EMAIL"),
	)
	// 3. Get AppBaseURL
	appBaseURL := os.Getenv("APP_BASE_URL")
	// --- END ADDITION ---

	s := &Server{
		port:        port,
		db:          database.New(), // Initialize your custom DB service
		echo:        echo.New(),
		logger:      NewPrettyLogger(), // Initialize the pretty logger
		emailSender: emailSender,       // <--- ASSIGN TO STRUCT FIELD
		appBaseURL:  appBaseURL,        // <--- ASSIGN TO STRUCT FIELD
		entClient:   entClient,         // <--- ASSIGN THE INITIALIZED ENT CLIENT HERE
		// appConfig field will remain nil as NewServer() doesn't receive a config object.
	}

	s.setupMiddleware()
	s.RegisterRoutes() // This will call public and private route registration

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.echo, // Use the echo instance as the handler
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

// setupMiddleware configures all common middlewares for the Echo instance.
func (s *Server) setupMiddleware() {
	s.echo.Validator = config.NewValidator() // Set custom validator

	// Remove trailing slash
	s.echo.Pre(middleware.RemoveTrailingSlash())

	// Request Logger Middleware
	s.echo.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogHost:   true,
		LogMethod: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			coloredStatus := colorizeStatus(v.Status) // Using the colorizeStatus from logger.go
			s.logger.Info(fmt.Sprintf("| %-6s | %s | %s | status: %s",
				v.Method,
				v.URI,
				v.Host,
				coloredStatus,
			))
			return nil
		},
	}))

	// Recovery Middleware
	s.echo.Use(middleware.Recover())

	// CORS Middleware
	s.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

func (s *Server) PrintRoutes() {
	s.logger.Info("--------------------------------------------------")
	s.logger.Info("  Registered Routes:")
	s.logger.Info("--------------------------------------------------")
	for _, route := range s.echo.Routes() {
		// You can customize the format here
		s.logger.Info(fmt.Sprintf("%-10s %-30s -> %s", route.Method, route.Path, route.Name))
	}
	s.logger.Info("--------------------------------------------------")
}

// RegisterRoutes registers all public and private routes.
// NOTE: When you initialize your handlers in the "separate page" you mentioned,
// they will now properly receive s.entClient, s.emailSender and s.appBaseURL
// from this Server instance.
func (s *Server) RegisterRoutes() {
	s.registerPublicRoutes()
	s.registerPrivateRoutes()
}

// HelloWorldHandler is a simple handler for the "/" route.
func (s *Server) HelloWorldHandler(c echo.Context) error {
	resp := map[string]string{
		"message": "Hello World",
	}
	return c.JSON(http.StatusOK, resp)
}

// healthHandler is a simple handler for the "/health" route.
func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}
