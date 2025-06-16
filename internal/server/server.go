package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload" // Automatically loads .env file

	"rtglabs-go/internal/database"
	"rtglabs-go/internal/validators"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// Server holds the server configuration and dependencies.
type Server struct {
	port   int
	db     database.Service
	echo   *echo.Echo
	logger *zap.Logger
}

// NewServer initializes and returns a new HTTP server.
func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	s := &Server{
		port:   port,
		db:     database.New(), // Initialize your custom DB service
		echo:   echo.New(),
		logger: NewPrettyLogger(), // Initialize the pretty logger
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
	s.echo.Validator = validators.NewValidator() // Set custom validator

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
	return c.JSON(http.StatusOK, s.db.Health())
}

