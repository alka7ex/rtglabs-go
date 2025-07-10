package server

import (
	"net/http"

	auth_handlers "rtglabs-go/internal/handlers/auth"      // <-- Explicit alias
	bw_handlers "rtglabs-go/internal/handlers/bodyweights" // <-- Explicit alias
	exercise_handler "rtglabs-go/internal/handlers/exercise"
	workout_handler "rtglabs-go/internal/handlers/workout"
	workout_log_handler "rtglabs-go/internal/handlers/workout_log"

	"github.com/labstack/echo/v4"
)

// registerPrivateRoutes registers all routes that require authentication.
func (s *Server) registerPrivateRoutes() {
	// Instead of initializing Ent Client here, we use s.sqlDB which is
	// already initialized in NewServer().
	// entClient := database.NewEntClient() // REMOVE THIS LINE

	// Create the auth handler instance, passing s.sqlDB
	authHandler := auth_handlers.NewAuthHandler(s.sqlDB) // Change to s.sqlDB

	// Create the bodyweight handler instance, passing s.sqlDB
	bwHandler := bw_handlers.NewBodyweightHandler(s.sqlDB) // Change to s.sqlDB

	exerciseHandler := exercise_handler.NewExerciseHandler(s.sqlDB) // Change to s.sqlDB
	//
	workoutHandler := workout_handler.NewWorkoutHandler(s.sqlDB) // Change to s.sqlDB
	//
	workoutLogHandler := workout_log_handler.NewWorkoutLogHandler(s.sqlDB) // Change to s.sqlDB

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
			// Ensure ValidateToken in authHandler uses s.sqlDB for its token validation logic.
			userID, err := authHandler.ValidateToken(token)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
			}

			// FIX 2: Set the correct key "user_id" for the handlers to retrieve.
			c.Set("user_id", userID)
			return next(c)
		}
	})

	g.POST("/logout", authHandler.DestroySession) // Or DELETE /auth/session/:token if you prefer
	// Protected Profile routes
	g.GET("/user/profile", authHandler.GetProfile)
	g.PUT("/user/profile", authHandler.UpdateProfile)

	// Protected Bodyweight routes
	g.POST("/bodyweights", bwHandler.StoreBodyweight)
	g.GET("/bodyweights", bwHandler.IndexBodyweight)
	g.GET("/bodyweights/:id", bwHandler.GetBodyweight)
	g.PUT("/bodyweights/:id", bwHandler.UpdateBodyweight)
	g.DELETE("/bodyweights/:id", bwHandler.DestroyBodyweight)

	// Protected Exercise routes
	g.GET("/exercise", exerciseHandler.IndexExercise)
	g.POST("/exercise", exerciseHandler.StoreExercise)
	// // Protected Workout routes
	g.POST("/workouts", workoutHandler.StoreWorkout)
	g.GET("/workouts", workoutHandler.IndexWorkout)
	g.GET("/workouts/:id", workoutHandler.GetWorkout)
	g.PUT("/workouts/:id", workoutHandler.UpdateWorkout)
	g.DELETE("/workouts/:id", workoutHandler.DestroyWorkout)
	//
	g.GET("/workout-logs", workoutLogHandler.IndexWorkoutLog)
	g.POST("/workout-logs", workoutLogHandler.StoreWorkoutLog)
	g.GET("/workout-logs/:id", workoutLogHandler.ShowWorkoutLog)
	g.PUT("/workout-logs/:id", workoutLogHandler.UpdateWorkoutLog)
	g.DELETE("/workout-logs/:id", workoutLogHandler.DestroyWorkoutLog)
}
