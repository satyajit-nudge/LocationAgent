package routes

import (
	"agent-backend/src/handlers"
	"agent-backend/src/middleware"
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo) {
	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
		})
	})

	// Protected routes
	api := e.Group("/api")
	api.Use(middleware.FirebaseAuth)

	// Add your protected routes here
	api.GET("/sharedlocations/:userId", handlers.GetSharedLocations)

	// Public routes can be added directly to 'e' if needed
}
