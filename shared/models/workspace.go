package models

// Workspace represents a tenant workspace in the multi-tenant system
type Workspace struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	TeamID      string `json:"teamId"`
	OwnerID     string `json:"ownerId"`
	CreatedAt   string `json:"createdAt"`
	Status      string `json:"status"` // active, suspended, archived
	Plan        string `json:"plan"`   // free, pro, enterprise
	TenantID    string `json:"tenantId"` // For custom multi-tenancy
}

// WorkspaceCreateRequest represents the input for creating a new workspace
type WorkspaceCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Plan        string `json:"plan,omitempty"` // defaults to "free"
}

// WorkspaceResponse represents the API response for workspace operations
type WorkspaceResponse struct {
	Success   bool      `json:"success"`
	Workspace Workspace `json:"workspace,omitempty"`
	Message   string    `json:"message,omitempty"`
	Error     string    `json:"error,omitempty"`
}


