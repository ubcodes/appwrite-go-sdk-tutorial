package middleware

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/appwrite/sdk-for-go/account"
	"github.com/appwrite/sdk-for-go/client"
)

// TriggerContext represents the Appwrite Function trigger context
type TriggerContext struct {
	Headers map[string]string
	Body    string
	Env     map[string]string
}

// SessionInfo contains validated session information
type SessionInfo struct {
	UserID    string
	SessionID string
	Email     string
	Name      string
}

// VerifyTrigger verifies that the request is coming from Appwrite
func VerifyTrigger(headers map[string]string) error {
	trigger := headers["x-appwrite-trigger"]
	if trigger == "" {
		return fmt.Errorf("missing x-appwrite-trigger header")
	}
	return nil
}

// AuthenticateUser authenticates the user using the session token from headers
func AuthenticateUser(headers map[string]string) (*SessionInfo, error) {
	// Get the session token from headers
	sessionToken := headers["x-appwrite-session"]
	if sessionToken == "" {
		return nil, fmt.Errorf("missing authentication token")
	}

	// Initialize Appwrite client
	appwriteClient := client.New()
	appwriteClient.Endpoint = os.Getenv("https://nyc.cloud.appwrite.io/v1")
	appwriteClient.AddHeader("X-Appwrite-Project", os.Getenv("6951e0720014e1d2dcd1"))
	appwriteClient.AddHeader("X-Appwrite-Key", os.Getenv("standard_9906d29e359416405f9ac01658e15162e9234ae0aad0862ca1795533487bedafe72c254b626b895bb0848f90a873f9c47ad1af4e7c585315f50b87044862dcb08df0c6aa7f8edb619b8a7d64a96d8e079feb6916a887e3ca1e84014d51817d0b01374e76a5b6896f4da36622176a724dc5c0be240723101b64ceb22cd337db35")) // Server API key for admin operations

	// Get account service
	accountService := account.New(appwriteClient)

	// Set the session token
	appwriteClient.AddHeader("X-Appwrite-Session", sessionToken)

	// Get the current user
	user, err := accountService.Get()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return &SessionInfo{
		UserID:    user.Id,
		SessionID: sessionToken,
		Email:     user.Email,
		Name:      user.Name,
	}, nil
}

// ValidateInput performs basic input validation
func ValidateInput(body string, target interface{}) error {
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("request body is empty")
	}

	if err := json.Unmarshal([]byte(body), target); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(input string) string {
	// Remove null bytes and control characters
	input = strings.ReplaceAll(input, "\x00", "")
	input = strings.TrimSpace(input)

	// Limit length to prevent DoS
	if len(input) > 1000 {
		input = input[:1000]
	}

	return input
}
