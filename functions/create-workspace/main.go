package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"appwrite-go-sdk-tutorial/shared/middleware"
	"appwrite-go-sdk-tutorial/shared/models"

	"github.com/appwrite/sdk-for-go/client"
	"github.com/appwrite/sdk-for-go/databases"
	"github.com/appwrite/sdk-for-go/teams"
)

// AppwriteFunctionRequest represents the incoming request structure
type AppwriteFunctionRequest struct {
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	Env     map[string]string `json:"env"`
}

// AppwriteFunctionResponse represents the response structure
type AppwriteFunctionResponse struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

func main() {
	// This is the entry point for Appwrite Functions
	// The function receives JSON input via stdin
	var req AppwriteFunctionRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		respondError(500, "INVALID_REQUEST", "Failed to parse request", "")
		return
	}

	// Execute the handler
	response := handleRequest(req)

	// Output the response
	output, _ := json.Marshal(response)
	fmt.Print(string(output))
}

func handleRequest(req AppwriteFunctionRequest) AppwriteFunctionResponse {
	// Step 1: Verify the trigger
	if err := middleware.VerifyTrigger(req.Headers); err != nil {
		return respondError(401, "UNAUTHORIZED", err.Error(), "")
	}

	// Step 2: Authenticate the user
	sessionInfo, err := middleware.AuthenticateUser(req.Headers)
	if err != nil {
		return respondError(401, "AUTHENTICATION_FAILED", err.Error(), "")
	}

	// Step 3: Parse and validate input
	var createReq models.WorkspaceCreateRequest
	if err := middleware.ValidateInput(req.Body, &createReq); err != nil {
		return respondError(400, "INVALID_INPUT", err.Error(), "")
	}

	// Step 4: Sanitize and validate workspace name
	workspaceName := middleware.SanitizeString(createReq.Name)
	if workspaceName == "" {
		return respondError(400, "VALIDATION_ERROR", "Workspace name is required", "name")
	}

	// Validate name length
	if len(workspaceName) < 3 || len(workspaceName) > 50 {
		return respondError(400, "VALIDATION_ERROR", "Workspace name must be between 3 and 50 characters", "name")
	}

	// Validate name format (alphanumeric, spaces, hyphens, underscores)
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\s\-_]+$`, workspaceName)
	if !matched {
		return respondError(400, "VALIDATION_ERROR", "Workspace name contains invalid characters", "name")
	}

	// Step 5: Check subscription status (example - in production, query your billing system)
	plan := createReq.Plan
	if plan == "" {
		plan = "free"
	}

	// Step 6: Initialize Appwrite Server SDK with Admin permissions
	appwriteClient := client.New()
	appwriteClient.Endpoint = req.Env["APPWRITE_ENDPOINT"]
	appwriteClient.AddHeader("X-Appwrite-Project", req.Env["APPWRITE_PROJECT_ID"])
	appwriteClient.AddHeader("X-Appwrite-Key", req.Env["APPWRITE_API_KEY"]) // Server API key for admin operations

	// Step 7: Create Appwrite Team for multi-tenancy
	teamsService := teams.New(appwriteClient)
	teamName := fmt.Sprintf("%s Workspace", workspaceName)
	team, err := teamsService.Create("unique()", teamName)
	if err != nil {
		log.Printf("Error creating team: %v", err)
		return respondError(500, "TEAM_CREATION_FAILED", "Failed to create workspace team", "")
	}

	// Step 8: Create workspace collection document
	databasesService := databases.New(appwriteClient)
	databaseID := req.Env["6952f51d00043d5bb1d9"]
	collectionID := req.Env["workspaces"]

	// Generate workspace slug
	slug := generateSlug(workspaceName)

	// Create workspace document
	workspaceDoc := map[string]interface{}{
		"name":        workspaceName,
		"slug":        slug,
		"teamId":      team.Id,
		"ownerId":     sessionInfo.UserID,
		"status":      "active",
		"plan":        plan,
		"tenantId":    team.Id, // Use team ID as tenant ID for data isolation
		"createdAt":   time.Now().Format(time.RFC3339),
		"description": createReq.Description,
	}

	// Set permissions: read for team members, write for team owners
	permissions := []string{
		fmt.Sprintf("read(\"team:%s\")", team.Id),
		fmt.Sprintf("write(\"team:%s[owner]\")", team.Id),
	}

	doc, err := databasesService.CreateDocument(
		databaseID,
		collectionID,
		"unique()", // Auto-generate document ID
		workspaceDoc,
		databases.CreateDocumentOption(func(opts *databases.CreateDocumentOptions) {
			opts.Permissions = permissions
		}),
	)
	if err != nil {
		log.Printf("Error creating workspace document: %v", err)
		// Cleanup: Delete the team if document creation fails
		teamsService.Delete(team.Id)
		return respondError(500, "WORKSPACE_CREATION_FAILED", "Failed to create workspace document", "")
	}

	// Step 9: Initialize private collection for workspace data
	// This demonstrates creating a collection that only this workspace can access
	_, err = initializeWorkspaceCollection(
		databasesService,
		databaseID,
		workspaceName,
		team.Id,
	)
	if err != nil {
		log.Printf("Warning: Failed to initialize workspace collection: %v", err)
		// Non-critical error, continue
	}

	// Step 10: Send welcome webhook to external service
	if err := sendWelcomeWebhook(req.Env, workspaceName, sessionInfo.Email, team.Id); err != nil {
		log.Printf("Warning: Failed to send welcome webhook: %v", err)
		// Non-critical error, continue
	}

	// Step 11: Return success response
	workspace := models.Workspace{
		ID:        doc.Id,
		Name:      workspaceName,
		Slug:      slug,
		TeamID:    team.Id,
		OwnerID:   sessionInfo.UserID,
		CreatedAt: workspaceDoc["createdAt"].(string),
		Status:    "active",
		Plan:      plan,
		TenantID:  team.Id,
	}

	response := models.WorkspaceResponse{
		Success:   true,
		Workspace: workspace,
		Message:   "Workspace created successfully",
	}

	responseBody, _ := json.Marshal(response)
	return AppwriteFunctionResponse{
		StatusCode: 200,
		Body:       string(responseBody),
	}
}

// generateSlug creates a URL-friendly slug from a workspace name
func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

// initializeWorkspaceCollection creates a private collection for workspace-specific data
func initializeWorkspaceCollection(
	dbService *databases.Databases,
	databaseID string,
	workspaceName string,
	teamID string,
) (string, error) {
	// In a real implementation, you would create a collection here
	// For this PoC, we'll return a placeholder
	// The collection would have permissions like:
	// - Read: team:teamID
	// - Write: team:teamID[owner]

	// Note: Collection creation via SDK might require additional setup
	// This is a conceptual example of what you'd do

	return "", nil // Placeholder - actual implementation would create collection
}

// sendWelcomeWebhook demonstrates secure external API integration using Appwrite Secrets
func sendWelcomeWebhook(env map[string]string, workspaceName, userEmail, teamID string) error {
	// Get webhook URL from Appwrite Secrets (never hardcoded)
	webhookURL := env["WEBHOOK_URL"]
	if webhookURL == "" {
		return fmt.Errorf("webhook URL not configured")
	}

	// Get API key from Appwrite Secrets (secure, never exposed to client)
	apiKey := env["WEBHOOK_API_KEY"]
	if apiKey == "" {
		return fmt.Errorf("webhook API key not configured")
	}

	// In a real implementation, you would make an HTTP request here
	// Example:
	// client := &http.Client{Timeout: 10 * time.Second}
	// req, _ := http.NewRequest("POST", webhookURL, payload)
	// req.Header.Set("Authorization", "Bearer "+apiKey)
	// resp, err := client.Do(req)

	log.Printf("Would send webhook to %s for workspace %s (team: %s)", webhookURL, workspaceName, teamID)

	return nil
}

// respondError creates a standardized error response
func respondError(statusCode int, code, message, field string) AppwriteFunctionResponse {
	errorResp := models.NewErrorResponse(code, message, field)
	body, _ := json.Marshal(errorResp)

	return AppwriteFunctionResponse{
		StatusCode: statusCode,
		Body:       string(body),
	}
}
