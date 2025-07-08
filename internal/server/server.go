package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"rtglabs-go/config"
	"rtglabs-go/config/database"
	mail "rtglabs-go/provider"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// Server holds the server configuration and dependencies.
type Server struct {
	port        int
	db          database.Service
	echo        *echo.Echo
	logger      *zap.Logger
	emailSender mail.EmailSender
	appBaseURL  string
	appConfig   *config.AppConfig
	sqlDB       *sql.DB
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
	log.Printf("DATABASE_URL: %s", os.Getenv("DATABASE_URL"))

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	driverName := os.Getenv("DB_DRIVER")
	if driverName == "" {
		log.Fatal("DB_DRIVER environment variable is not set (e.g., mysql, sqlite3, postgres)")
	}

	sqlDB, err := database.NewSQLClient(driverName, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err = sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection established successfully!")

	emailSender := mail.NewSMTPEmailSender(
		os.Getenv("SMTP_HOST"),
		os.Getenv("SMTP_PORT"),
		os.Getenv("SMTP_USERNAME"),
		os.Getenv("SMTP_PASSWORD"),
		os.Getenv("SMTP_FROM_EMAIL"),
	)

	appBaseURL := os.Getenv("APP_BASE_URL")

	s := &Server{
		port:        port,
		db:          database.New(),
		echo:        echo.New(),
		logger:      NewPrettyLogger(),
		emailSender: emailSender,
		appBaseURL:  appBaseURL,
		sqlDB:       sqlDB,
	}

	s.setupMiddleware()
	s.RegisterRoutes()

	// --- NEW: Set custom HTTPErrorHandler for 404s ---
	s.echo.HTTPErrorHandler = s.customHTTPErrorHandler

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.echo,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

// customHTTPErrorHandler is a custom HTTP error handler that provides more informative
// messages for 404 Not Found errors.
func (s *Server) customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Internal Server Error"
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = fmt.Sprintf("%v", he.Message) // Ensure message is a string
	}

	// For 404 Not Found errors, provide a more specific message
	if code == http.StatusNotFound {
		requestedPath := c.Request().URL.Path
		requestedMethod := c.Request().Method
		message = fmt.Sprintf("The endpoint '%s' with method '%s' is not available.", requestedPath, requestedMethod)
		s.logger.Warn(fmt.Sprintf("404 Not Found: %s %s", requestedMethod, requestedPath)) // Log the 404
	}

	// Always log errors, except for 404 which we specifically log as Warn
	if code != http.StatusNotFound {
		s.logger.Error("HTTP Error",
			zap.Int("status", code),
			zap.String("method", c.Request().Method),
			zap.String("path", c.Request().URL.Path),
			zap.Error(err),
		)
	}

	// Respond with JSON error
	if !c.Response().Committed {
		if err := c.JSON(code, map[string]string{"error": message}); err != nil {
			s.logger.Error("Failed to send error response", zap.Error(err))
		}
	}
}

// setupMiddleware configures all common middlewares for the Echo instance.
func (s *Server) setupMiddleware() {
	s.echo.Validator = config.NewValidator()

	s.echo.Pre(middleware.RemoveTrailingSlash())

	s.echo.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogHost:   true,
		LogMethod: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			coloredStatus := colorizeStatus(v.Status)
			s.logger.Info(fmt.Sprintf("| %-6s | %s | %s | status: %s",
				v.Method,
				v.URI,
				v.Host,
				coloredStatus,
			))
			return nil
		},
	}))

	s.echo.Use(middleware.Recover())

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
	s.logger.Info("  Registered Routes:")
	s.logger.Info("--------------------------------------------------")
	for _, route := range s.echo.Routes() {
		s.logger.Info(fmt.Sprintf("%-10s %-30s -> %s", route.Method, route.Path, route.Name))
	}
	s.logger.Info("--------------------------------------------------")
}

// RegisterRoutes registers all public and private routes.
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
	if err := s.sqlDB.Ping(); err != nil {
		s.logger.Error("Database health check failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"status": "unhealthy",
			"error":  "database connection failed",
		})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// You need to define `colorizeStatus` somewhere, likely in `logger.go` as you mentioned.
// For demonstration, I'll include a placeholder if it's not provided elsewhere.
// Assuming colorizeStatus is in logger.go, no need to duplicate it here.
// If it's not, you'd need something like:
/*
import (
	"github.com/fatih/color"
)

func colorizeStatus(status int) string {
	switch {
	case status >= 200 && status < 300:
		return color.GreenString("%d", status)
	case status >= 300 && status < 400:
		return color.YellowString("%d", status)
	case status >= 400 && status < 500:
		return color.RedString("%d", status)
	case status >= 500:
		return color.MagentaString("%d", status)
	default:
		return fmt.Sprintf("%d", status)
	}
}
*/
