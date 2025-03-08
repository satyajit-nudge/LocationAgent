package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// ValidatePhoneNumber checks if the phone number is in E.164 format
func ValidatePhoneNumber(phoneNumber string) error {
	if !phoneRegex.MatchString(phoneNumber) {
		return fmt.Errorf("invalid phone number format")
	}
	return nil
}

// InitFirebaseApp initializes and returns a Firebase App instance
func InitFirebaseApp() (*firebase.App, error) {
	// Get the current file's directory
	_, currentFile, _, _ := runtime.Caller(0)
	// Get the project root directory
	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))
	// Construct the path to the service account key
	serviceAccountPath := filepath.Join(rootDir, "config", "serviceAccountKey.json")

	log.Printf("Loading service account from: %s", serviceAccountPath)

	// Initialize Firebase Admin SDK
	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}
	return app, nil
}

// exchangeCustomTokenForIDToken exchanges a custom token for an ID token
func exchangeCustomTokenForIDToken(customToken string) (string, error) {
	// Get Firebase Web API Key from environment
	apiKey := os.Getenv("FIREBASE_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("FIREBASE_API_KEY environment variable not set")
	}

	log.Printf("Exchanging custom token for ID token...")

	// Exchange custom token for ID token
	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithCustomToken?key=%s", apiKey)
	reqBody := map[string]interface{}{
		"token":             customToken,
		"returnSecureToken": true,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error exchanging token: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	var result struct {
		IDToken string `json:"idToken"`
		Error   struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error parsing response: %v", err)
	}

	if result.Error.Message != "" {
		return "", fmt.Errorf("error from Firebase: %s", result.Error.Message)
	}

	if result.IDToken == "" {
		return "", fmt.Errorf("no ID token in response")
	}

	log.Printf("Successfully obtained ID token")
	return result.IDToken, nil
}

// GeneratePhoneAuthToken creates a custom token for phone number authentication
func GeneratePhoneAuthToken(phoneNumber string) (string, error) {
	if err := ValidatePhoneNumber(phoneNumber); err != nil {
		return "", err
	}

	app, err := InitFirebaseApp()
	if err != nil {
		return "", err
	}

	// Get Auth client
	client, err := app.Auth(context.Background())
	if err != nil {
		return "", fmt.Errorf("error getting Auth client: %v", err)
	}

	// Create or update user with phone number
	params := (&auth.UserToCreate{}).
		PhoneNumber(phoneNumber)

	user, err := client.CreateUser(context.Background(), params)
	if err != nil {
		log.Printf("User creation failed, trying to get existing user: %v", err)
		// If user already exists, try to get user by phone number
		user, err = client.GetUserByPhoneNumber(context.Background(), phoneNumber)
		if err != nil {
			return "", fmt.Errorf("error getting user by phone: %v", err)
		}
	}

	// Set custom claims for the user
	claims := map[string]interface{}{
		"phone_verified": true,
		"phone_number":   phoneNumber,
	}
	if err := client.SetCustomUserClaims(context.Background(), user.UID, claims); err != nil {
		return "", fmt.Errorf("error setting custom claims: %v", err)
	}

	// Create a custom token
	customToken, err := client.CustomToken(context.Background(), user.UID)
	if err != nil {
		return "", fmt.Errorf("error creating custom token: %v", err)
	}

	// Exchange custom token for ID token
	idToken, err := exchangeCustomTokenForIDToken(customToken)
	if err != nil {
		return "", fmt.Errorf("error exchanging custom token: %v", err)
	}

	log.Printf("Successfully generated ID token")
	return idToken, nil
}

// VerifyPhoneNumber verifies the phone number with OTP
func VerifyPhoneNumber(phoneNumber, code string) (string, error) {
	if err := ValidatePhoneNumber(phoneNumber); err != nil {
		return "", err
	}

	// For testing purposes, we'll accept any non-empty code
	if code == "" {
		return "", fmt.Errorf("verification code is required")
	}

	return GeneratePhoneAuthToken(phoneNumber)
}

// GenerateIDToken creates a Firebase ID token using the Admin SDK (keeping for backward compatibility)
func GenerateIDToken(uid string) (string, error) {
	app, err := InitFirebaseApp()
	if err != nil {
		return "", err
	}

	// Get Auth client
	client, err := app.Auth(context.Background())
	if err != nil {
		return "", fmt.Errorf("error getting Auth client: %v", err)
	}

	// Create or update the test user
	params := (&auth.UserToCreate{}).
		UID(uid).
		Email(fmt.Sprintf("%s@example.com", uid)).
		EmailVerified(true).
		DisplayName("Test User")

	user, err := client.CreateUser(context.Background(), params)
	if err != nil {
		// If user already exists, try to get the user
		user, err = client.GetUser(context.Background(), uid)
		if err != nil {
			return "", fmt.Errorf("error creating/getting user: %v", err)
		}
	}

	// Set custom claims for the user
	claims := map[string]interface{}{
		"admin": true,
		"role":  "tester",
	}
	if err := client.SetCustomUserClaims(context.Background(), user.UID, claims); err != nil {
		return "", fmt.Errorf("error setting custom claims: %v", err)
	}

	// Create a custom token
	token, err := client.CustomToken(context.Background(), user.UID)
	if err != nil {
		return "", fmt.Errorf("error creating custom token: %v", err)
	}

	return token, nil
}
