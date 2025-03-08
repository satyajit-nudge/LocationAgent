package test

import (
	"agent-backend/src/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

type Location struct {
	UserID    string    `json:"user_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
}

type APIResponse struct {
	Message string     `json:"message,omitempty"`
	Error   string     `json:"error,omitempty"`
	Data    []Location `json:"data,omitempty"`
}

func TestPhoneAuthAPIAccess(t *testing.T) {
	// Test setup
	phoneNumber := "+17206453833"
	testCode := "123456"

	// Step 1: Generate token using phone authentication
	token, err := utils.VerifyPhoneNumber(phoneNumber, testCode)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// Wait for server to process token
	time.Sleep(1 * time.Second)

	// Step 2: Test protected routes
	tests := []struct {
		name          string
		endpoint      string
		method        string
		expectedCode  int
		checkResponse func(*testing.T, *http.Response)
	}{
		{
			name:         "Get Shared Locations",
			endpoint:     fmt.Sprintf("/api/sharedlocations/%s", phoneNumber),
			method:       "GET",
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, resp *http.Response) {
				// Log response headers for debugging

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("Failed to read response body: %v", err)
					return
				}

				// Try to decode as array first
				var locations []Location
				err = json.Unmarshal(body, &locations)
				if err == nil {
					// Successfully decoded as array
					if len(locations) == 0 {
						t.Log("No locations returned (this might be expected)")
						return
					}
					return
				}

				// If array decode failed, try as APIResponse
				var apiResp APIResponse
				if err := json.Unmarshal(body, &apiResp); err != nil {
					t.Errorf("Failed to decode response as either array or APIResponse: %v", err)
					return
				}

				// Check for error in APIResponse
				if apiResp.Error != "" {
					t.Errorf("Expected no error, got: %s", apiResp.Error)
					return
				}

				// Verify locations data from APIResponse
				if len(apiResp.Data) == 0 {
					t.Log("No locations returned (this might be expected)")
					return
				}
			},
		},
		{
			name:         "Access Without Token",
			endpoint:     fmt.Sprintf("/api/sharedlocations/%s", phoneNumber),
			method:       "GET",
			expectedCode: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, resp *http.Response) {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("Failed to read response body: %v", err)
					return
				}

				// Try to decode as APIResponse first
				var apiResp APIResponse
				if err := json.Unmarshal(body, &apiResp); err != nil {
					// If that fails, try as a simple map
					var errorResp map[string]interface{}
					if err := json.Unmarshal(body, &errorResp); err != nil {
						t.Errorf("Failed to decode error response: %v", err)
						return
					}
					if _, hasMessage := errorResp["message"]; !hasMessage {
						t.Error("Expected error message, got none")
					}
					return
				}

				if apiResp.Error == "" && apiResp.Message == "" {
					t.Error("Expected error message, got none")
				}
			},
		},
	}

	baseURL := "http://localhost:8080"
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, baseURL+tt.endpoint, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Add authorization header for protected routes
			if tt.expectedCode != http.StatusUnauthorized {
				authHeader := fmt.Sprintf("Bearer %s", token)
				req.Header.Set("Authorization", authHeader)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedCode {
				// Read and log the response body for debugging
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, resp.StatusCode)
				return
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, resp)
			}
		})
	}
}

func TestServerAvailability(t *testing.T) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("http://localhost:8080/health")
	if err != nil {
		t.Skip("Server is not running. Start the server to run integration tests.")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
