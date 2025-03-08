package main

import (
	"agent-backend/src/middleware"
	"agent-backend/src/routes"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	// Initialize Firebase
	middleware.InitializeFirebase()

	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())

	// Add CORS middleware if needed
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	routes.SetupRoutes(e)

	e.Logger.Fatal(e.Start(":8080"))
}
