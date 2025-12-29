package models

// UserProfile represents extended user information
type UserProfile struct {
	UserID      string   `json:"userId"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	TeamIDs     []string `json:"teamIds"`
	WorkspaceIDs []string `json:"workspaceIds"`
	SubscriptionStatus string `json:"subscriptionStatus"` // active, trial, expired
}

// SessionContext represents the authenticated session context
type SessionContext struct {
	UserID      string `json:"userId"`
	SessionID   string `json:"sessionId"`
	TeamID      string `json:"teamId,omitempty"`
	IsAdmin     bool   `json:"isAdmin"`
}


