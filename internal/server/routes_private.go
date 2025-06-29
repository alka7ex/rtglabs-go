package server

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"rtglabs-go/internal/database"
	"rtglabs-go/internal/handlers"
)

// registerPrivateRoutes registers all routes that require authentication.
func (s *Server) registerPrivateRoutes() {
	// Initialize Ent Client here for private handlers
	entClient := database.NewEntClient()
	authHandler := handlers.NewAuthHandler(entClient)

	// FIX 1: Create the group from the server's Echo instance.
	g := s.echo.Group("/admin")

	// Middleware for protected routes
	g.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" || len(auth) < 8 || auth[:7] != "Bearer " {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid Authorization header")
			}

			token := auth[7:]
			// Call ValidateToken from the auth handler.
			userID, err := authHandler.ValidateToken(token)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			// FIX 2: Set the correct key "user_id" for the handlers to retrieve.
			c.Set("user_id", userID)
			return next(c)
		}
	})

	// Protected Profile routes
	// These routes are now prefixed with "/admin".
	g.GET("/profile", authHandler.GetProfile)
	g.PUT("/profile", authHandler.UpdateProfile)

	// Protected Bodyweight routes
	// These routes are also now prefixed with "/admin".
	bwHandler := handlers.NewBodyweightHandler(entClient)
	g.POST("/bodyweights", bwHandler.CreateBodyweight)
	g.GET("/bodyweights", bwHandler.ListBodyweights)
	g.GET("/bodyweights/:id", bwHandler.GetBodyweight)
	g.PUT("/bodyweights/:id", bwHandler.UpdateBodyweight)
	g.DELETE("/bodyweights/:id", bwHandler.DeleteBodyweight)
}

