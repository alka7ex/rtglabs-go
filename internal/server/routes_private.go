package server

import (
	"net/http"

	"rtglabs-go/config/database"
	auth_handlers "rtglabs-go/internal/handlers/auth"      // <-- Explicit alias
	bw_handlers "rtglabs-go/internal/handlers/bodyweights" // <-- Explicit alias
	exercise_handler "rtglabs-go/internal/handlers/exercise"
	workout_handler "rtglabs-go/internal/handlers/workout"
	workout_log_handler "rtglabs-go/internal/handlers/workout_log"

	"github.com/labstack/echo/v4"
)

// registerPrivateRoutes registers all routes that require authentication.
func (s *Server) registerPrivateRoutes() {
	// Initialize Ent Client here for private handlers
	entClient := database.NewEntClient()

	// Create the auth handler instance using the correct package alias
	authHandler := auth_handlers.NewAuthHandler(entClient)

	// Create the bodyweight handler instance using the correct package alias
	bwHandler := bw_handlers.NewBodyweightHandler(entClient)

	exerciseHandler := exercise_handler.NewExerciseHandler(entClient)

	workoutHandler := workout_handler.NewWorkoutHandler(entClient)

	workoutLogHandler := workout_log_handler.NewWorkoutLogHandler(entClient)

	// FIX 1: Create the group from the server's Echo instance.
	g := s.echo.Group("/api")

	// Middleware for protected routes
	g.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid Authorization header")
			}

			token := authHeader[7:]
			// Call ValidateToken from the auth handler instance.
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
	g.POST("/bodyweights", bwHandler.StoreBodyweight)
	g.GET("/bodyweights", bwHandler.IndexBodyweight)
	g.GET("/bodyweights/:id", bwHandler.GetBodyweight)
	g.PUT("/bodyweights/:id", bwHandler.UpdateBodyweight)
	g.DELETE("/bodyweights/:id", bwHandler.DestroyBodyweight)

	// Protected Exercise routes
	g.GET("/exercise", exerciseHandler.IndexExercise)

	// Protected Workout routes
	g.POST("/workouts", workoutHandler.StoreWorkout)  // Create
	g.GET("/workouts", workoutHandler.IndexWorkout)   // List
	g.GET("/workouts/:id", workoutHandler.GetWorkout) // Get (Show)
	g.PUT("/workouts/:id", workoutHandler.UpdateWorkout)
	g.DELETE("/workouts/:id", workoutHandler.DestroyWorkout) // Delete (Soft Delete)

	g.GET("/workout-logs", workoutLogHandler.IndexWorkoutLog)
	g.POST("/workout-logs", workoutLogHandler.StoreWorkoutLog)
	g.GET("/workout-logs/:id", workoutLogHandler.ShowWorkoutLog)
	g.DELETE("/workout-logs/:id", workoutLogHandler.DestroyWorkoutLog)
}
