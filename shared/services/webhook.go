package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// WebhookPayload represents the data sent to external webhooks
type WebhookPayload struct {
	Event     string                 `json:"event"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// WebhookService handles secure external API communication
type WebhookService struct {
	client *http.Client
}

// NewWebhookService creates a new webhook service
func NewWebhookService() *WebhookService {
	return &WebhookService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendWelcomeWebhook sends a welcome notification to an external service
// This demonstrates how to securely use API keys stored in Appwrite Secrets
func (ws *WebhookService) SendWelcomeWebhook(
	workspaceName string,
	userEmail string,
	teamID string,
) error {
	webhookURL := os.Getenv("WEBHOOK_URL")
	apiKey := os.Getenv("WEBHOOK_API_KEY")

	if webhookURL == "" || apiKey == "" {
		return fmt.Errorf("webhook configuration missing")
	}

	payload := WebhookPayload{
		Event:     "workspace.created",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: map[string]interface{}{
			"workspaceName": workspaceName,
			"userEmail":     userEmail,
			"teamId":        teamID,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("X-Webhook-Source", "appwrite-function")

	resp, err := ws.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// SendBillingWebhook demonstrates integration with billing services like Stripe
func (ws *WebhookService) SendBillingWebhook(
	event string,
	billingData map[string]interface{},
) error {
	// Example: Integration with Stripe webhook
	// In production, you would:
	// 1. Get Stripe API key from Appwrite Secrets
	// 2. Create Stripe customer/subscription
	// 3. Handle webhook responses securely
	
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		return fmt.Errorf("stripe key not configured")
	}

	// Implementation would go here
	// This is a placeholder to show the pattern
	
	return nil
}


