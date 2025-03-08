package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	firebase "firebase.google.com/go/v4"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/option"
)

var firebaseApp *firebase.App

// InitializeFirebase initializes the Firebase Admin SDK
func InitializeFirebase() {
	ctx := context.Background()

	// Get the current file's directory
	_, currentFile, _, _ := runtime.Caller(0)
	// Get the project root directory (two levels up from current file)
	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))
	// Construct the path to the service account key
	serviceAccountPath := filepath.Join(rootDir, "config", "serviceAccountKey.json")

	// Initialize Firebase with the service account
	opt := option.WithCredentialsFile(serviceAccountPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v\n", err)
	}
	firebaseApp = app
	log.Println("Firebase initialized successfully")
}

// FirebaseAuth middleware to verify Firebase tokens
func FirebaseAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			log.Println("Missing authorization header")
			return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
		}

		token := strings.Replace(authHeader, "Bearer ", "", 1)
		if token == "" {
			log.Println("Invalid token format")
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token format")
		}

		client, err := firebaseApp.Auth(c.Request().Context())
		if err != nil {
			log.Printf("Error getting Auth client: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "error getting Auth client")
		}

		// Try to verify as an ID token first
		verifiedToken, err := client.VerifyIDToken(c.Request().Context(), token)
		if err != nil {
			log.Printf("ID token verification failed: %v", err)
			// If ID token verification fails, try to verify as a custom token
			customToken, err := client.VerifyIDToken(c.Request().Context(), token)
			if err != nil {
				log.Printf("Custom token verification failed: %v", err)
				return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("invalid token: %v", err))
			}
			log.Printf("Custom token verified for user: %s", customToken.UID)
			// Add the verified user ID to the context
			c.Set("userID", customToken.UID)
			return next(c)
		}

		// Add the verified user ID to the context
		c.Set("userID", verifiedToken.UID)
		return next(c)
	}
}
