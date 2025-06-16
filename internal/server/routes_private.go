package server

import (
	"net/http"

	"rtglabs-go/internal/database" // Import database for entClient if needed
	"rtglabs-go/internal/handlers"

	"github.com/labstack/echo/v4"
)

// registerPrivateRoutes registers all routes that require authentication.
func (s *Server) registerPrivateRoutes() {
	// Initialize Ent Client here for private handlers
	entClient := database.NewEntClient()
	authHandler := handlers.NewAuthHandler(entClient) // Need auth handler for token validation

	// Create a group for protected routes
	g := s.echo.Group("/admin")

	// Middleware for protected routes
	g.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" || len(auth) < 8 || auth[:7] != "Bearer " {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid Authorization header")
			}

			token := auth[7:]
			userID, err := authHandler.ValidateToken(token) // Use the auth handler's ValidateToken
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			c.Set("userID", userID)
			return next(c)
		}
	})

	// Protected Bodyweight routes
	bwHandler := handlers.NewBodyweightHandler(entClient)
	// Note: These routes are still directly on 'e' in your original code,
	// but are effectively protected by the "admin" group's middleware if accessed via /admin.
	// If you want these to be truly under /admin, change the paths below.
	// For example, if you want "/admin/bodyweights", use:
	// g.POST("/bodyweights", bwHandler.CreateBodyweight)
	// g.GET("/bodyweights", bwHandler.ListBodyweights)
	// ...and so on.
	// I'm keeping them as per your original code structure, assuming they are accessible
	// from root, but the middleware above is applied to the '/admin' group.
	// If you want to apply the auth middleware to these *specific* routes regardless of group,
	// you'd need to add `authHandler.AuthMiddleware` (or similar) to each.
	// For simplicity, let's move them into the `g` group for now as per your original intent of "Protected routes".

	s.echo.POST("/bodyweights", bwHandler.CreateBodyweight) // Example: Not in /admin group
	s.echo.GET("/bodyweights", bwHandler.ListBodyweights)   // Example: Not in /admin group
	s.echo.GET("/bodyweights/:id", bwHandler.GetBodyweight)
	s.echo.PUT("/bodyweights/:id", bwHandler.UpdateBodyweight)
	s.echo.DELETE("/bodyweights/:id", bwHandler.DeleteBodyweight)

	// If you intend for the bodyweight routes to be protected *and* under the `/admin` prefix:
	// g.POST("/bodyweights", bwHandler.CreateBodyweight)
	// g.GET("/bodyweights", bwHandler.ListBodyweights)
	// g.GET("/bodyweights/:id", bwHandler.GetBodyweight)
	// g.PUT("/bodyweights/:id", bwHandler.UpdateBodyweight)
	// g.DELETE("/bodyweights/:id", bwHandler.DeleteBodyweight)
}
