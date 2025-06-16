package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"rtglabs-go/cmd/web"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewPrettyLogger() *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // Colorized level
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

func colorizeStatus(status int) string {
	switch {
	case status >= 200 && status < 300:
		return fmt.Sprintf("\x1b[32m%d\x1b[0m", status) // Green
	case status >= 400 && status < 500:
		return fmt.Sprintf("\x1b[33m%d\x1b[0m", status) // Yellow
	case status >= 500:
		return fmt.Sprintf("\x1b[31m%d\x1b[0m", status) // Red
	default:
		return fmt.Sprintf("%d", status) // No color
	}
}

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	logger := NewPrettyLogger()

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogHost:   true,
		LogMethod: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			coloredStatus := colorizeStatus(v.Status)
			logger.Info(fmt.Sprintf("Incoming request | %-6s | %s | %s | status: %s | %s",
				v.Method,
				v.URI,
				v.Host,
				coloredStatus,
				time.Now().Format(time.RFC3339),
			))
			return nil
		},
	}))
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	fileServer := http.FileServer(http.FS(web.Files))

	// PUBLIC ROUTE
	e.GET("/assets/*", echo.WrapHandler(fileServer))
	e.GET("/web", echo.WrapHandler(templ.Handler(web.HelloForm())))
	e.POST("/hello", echo.WrapHandler(http.HandlerFunc(web.HelloWebHandler)))
	e.GET("/", s.HelloWorldHandler)
	e.GET("/health", s.healthHandler)

	// PRIVATE ROUTE
	g := e.Group("/admin")
	g.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return key == "valid-key", nil
	}))

	g.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract the credentials from HTTP request header and perform a security
			// check

			// For invalid credentials
			return echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")

			// For valid credentials call next
			// return next(c)
		}
	})

	return e
}

func (s *Server) HelloWorldHandler(c echo.Context) error {
	resp := map[string]string{
		"message": "Hello World",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}
