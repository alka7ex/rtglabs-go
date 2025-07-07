package server

import (
	"database/sql" // <--- NEW: Import for standard SQL DB
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload" // Automatically loads .env file

	"rtglabs-go/config"          // KEEP this import if 'appConfig' field is intended, though it won't be initialized here.
	"rtglabs-go/config/database" // Updated: This package will now provide *sql.DB
	mail "rtglabs-go/provider"   // THIS IS THE CORRECT IMPORT for EmailSender

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// Server holds the server configuration and dependencies.
type Server struct {
	port        int
	db          database.Service // This service should now wrap *sql.DB or use it directly
	echo        *echo.Echo
	logger      *zap.Logger
	emailSender mail.EmailSender
	appBaseURL  string
	appConfig   *config.AppConfig // This field remains, but will be nil as NewServer() doesn't take config.
	sqlDB       *sql.DB           // <--- NEW: Your standard SQL database client
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
	log.Printf("DATABASE_URL: %s", os.Getenv("DATABASE_URL")) // <--- NEW: Log DB URL

	// --- ADDITION: Initialize sql.DB, EmailSender, and AppBaseURL ---

	// 1. Initialize SQL DB Client using your updated database package
	// Assumes DATABASE_URL environment variable holds the connection string.
	// Example for SQLite: "file:./data.db?_foreign_keys=on"
	// Example for MySQL: "user:password@tcp(127.0.0.1:3306)/database?parseTime=true"
	// Example for PostgreSQL: "postgres://user:password@host:port/database?sslmode=disable"
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// You will need to change this to the appropriate driver name
	// e.g., "mysql", "sqlite3", "postgres"
	driverName := os.Getenv("DB_DRIVER") // <--- NEW: Get DB driver from env
	if driverName == "" {
		log.Fatal("DB_DRIVER environment variable is not set (e.g., mysql, sqlite3, postgres)")
	}

	sqlDB, err := database.NewSQLClient(driverName, dbURL) // <--- CALL YOUR NEW SQL CLIENT HERE
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Ping the database to ensure the connection is open
	if err = sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection established successfully!")

	// 2. Initialize EmailSender
	emailSender := mail.NewSMTPEmailSender(
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
		db:          database.New(), // Initialize your custom DB service (which should now use *sql.DB internally)
		echo:        echo.New(),
		logger:      NewPrettyLogger(), // Initialize the pretty logger
		emailSender: emailSender,       // ASSIGN TO STRUCT FIELD
		appBaseURL:  appBaseURL,        // ASSIGN TO STRUCT FIELD
		sqlDB:       sqlDB,             // <--- ASSIGN THE INITIALIZED SQL DB HERE
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
	s.logger.Info("  Registered Routes:")
	s.logger.Info("--------------------------------------------------")
	for _, route := range s.echo.Routes() {
		// You can customize the format here
		s.logger.Info(fmt.Sprintf("%-10s %-30s -> %s", route.Method, route.Path, route.Name))
	}
	s.logger.Info("--------------------------------------------------")
}

// RegisterRoutes registers all public and private routes.
// NOTE: When you initialize your handlers in the "separate page" you mentioned,
// they will now properly receive s.sqlDB, s.emailSender and s.appBaseURL
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
	// For health check with *sql.DB, you can ping the database
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

// You will also need to adjust your 'registerPublicRoutes' and 'registerPrivateRoutes'
// and all associated handlers to use `s.sqlDB` and `squirrel` for database operations,
// instead of `s.entClient`.
// This means you'll construct SQL queries using squirrel.StatementBuilder,
// and then use s.sqlDB.QueryRow, s.sqlDB.Exec, etc. to execute them.
