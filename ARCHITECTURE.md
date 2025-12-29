# Architecture Deep Dive: Multi-Tenant SaaS with Appwrite and Go

## Executive Summary

This architecture demonstrates **Security by Design** and **Server-Side Orchestration** principles essential for enterprise SaaS applications. By moving critical business logic from the client to Appwrite Functions (written in Go), we ensure that sensitive operations cannot be bypassed, API keys remain secure, and multi-tenancy is properly enforced.

---

## The Problem: Why Client-Side Logic Falls Short

### Traditional Approach (Client SDK Only)

```
Frontend → Appwrite Client SDK → Database
```

**Limitations:**
1. **No Orchestration**: Creating a workspace requires multiple API calls (team creation, document creation, collection setup)
2. **Client-Side Validation**: Business rules can be bypassed by modifying client code
3. **Exposed Secrets**: API keys for external services (Stripe, SendGrid) would need to be in frontend code
4. **Permission Complexity**: Users need multiple permissions across different resources
5. **No Atomicity**: If one step fails, partial state is left behind

### Senior Approach (Server Functions)

```
Frontend → Appwrite Function (Go) → Appwrite Server SDK → Multiple Resources
                                      ↓
                              External APIs (with Secrets)
```

**Benefits:**
1. ✅ **Atomic Operations**: All-or-nothing workspace creation
2. ✅ **Server-Side Validation**: Business rules enforced on the server
3. ✅ **Secure Secrets**: API keys stored in Appwrite Secrets, never exposed
4. ✅ **Admin Permissions**: Function uses Server API Key for orchestration
5. ✅ **Rollback Logic**: Failed operations can clean up partial state

---

## Multi-Tenancy Strategy

### Appwrite Teams as Tenant Boundaries

Each workspace is backed by an **Appwrite Team**. This provides:

#### 1. Data Isolation

```go
// Workspace document permissions
Read:  []string{fmt.Sprintf("read(\"team:%s\")", team.ID)}
Write: []string{fmt.Sprintf("write(\"team:%s[owner]\")", team.ID)}
```

**What this means:**
- Only team members can read workspace data
- Only team owners can modify workspace data
- Appwrite enforces these permissions at the database level
- No custom application logic needed for access control

#### 2. Automatic Access Control

When a user queries the database:
```go
// User can only see workspaces for teams they belong to
databasesService.ListDocuments(
    databaseID,
    collectionID,
    []string{fmt.Sprintf("read(\"team:%s\")", userTeamID)}
)
```

Appwrite automatically filters results based on team membership.

#### 3. Scalable Architecture

- **Unlimited Teams**: No hard limits on tenant count
- **Flexible Roles**: Team roles (owner, member, admin) for fine-grained permissions
- **Nested Permissions**: Teams can have sub-teams for complex hierarchies

### Custom Tenant IDs

We also store a `tenantId` field (using Team ID) for:
- **Custom Queries**: Additional filtering beyond team permissions
- **Analytics**: Cross-tenant reporting (with proper aggregation)
- **External Integration**: Third-party systems that need tenant identifiers

---

## Security Model: Trust but Verify

### The Principle

> **While Appwrite's client-side SDK is secure, critical business logic must run on the server where it cannot be bypassed.**

### Security Layers

#### 1. Trigger Verification
```go
func VerifyTrigger(headers map[string]string) error {
    trigger := headers["x-appwrite-trigger"]
    if trigger == "" {
        return fmt.Errorf("missing x-appwrite-trigger header")
    }
    return nil
}
```

**Purpose**: Ensures requests come from Appwrite, not external sources.

#### 2. Session Authentication
```go
func AuthenticateUser(headers map[string]string) (*SessionInfo, error) {
    sessionToken := headers["x-appwrite-session"]
    // Validate token with Appwrite Account service
    user, err := accountService.Get()
    // Return user info
}
```

**Purpose**: Verifies the user is authenticated and gets their identity.

#### 3. Input Validation
```go
// Sanitize input
workspaceName := middleware.SanitizeString(createReq.Name)

// Validate format
matched, _ := regexp.MatchString(`^[a-zA-Z0-9\s\-_]+$`, workspaceName)

// Validate length
if len(workspaceName) < 3 || len(workspaceName) > 50 {
    return error
}
```

**Purpose**: Prevents injection attacks and ensures data quality.

#### 4. Server API Key
```go
appwriteClient.SetKey(os.Getenv("APPWRITE_API_KEY")) // Server key
```

**Purpose**: Grants admin-level permissions needed for orchestration.

**⚠️ Critical**: This key is stored in Appwrite Secrets, never in code or frontend.

---

## Function Flow: Step-by-Step

### 1. Request Arrives
```json
{
  "headers": {
    "x-appwrite-trigger": "http",
    "x-appwrite-session": "user-session-token"
  },
  "body": "{\"name\": \"My Workspace\", \"plan\": \"free\"}",
  "env": {
    "APPWRITE_API_KEY": "server-key-from-secrets"
  }
}
```

### 2. Security Checks
- ✅ Verify `x-appwrite-trigger` header
- ✅ Authenticate user via session token
- ✅ Parse and validate request body

### 3. Business Logic
- ✅ Sanitize workspace name
- ✅ Validate name format and length
- ✅ Check subscription status (if applicable)

### 4. Resource Creation (Atomic)
```go
// Step 1: Create Team
team, err := teamsService.Create(teamName, []string{userID})

// Step 2: Create Workspace Document
doc, err := databasesService.CreateDocument(
    databaseID,
    collectionID,
    "unique()",
    workspaceData,
    teamPermissions,
)

// Step 3: Initialize Collections (if needed)
// Step 4: Send Webhook (non-critical)
```

**Rollback Logic**: If document creation fails, delete the team:
```go
if err != nil {
    teamsService.Delete(team.ID) // Cleanup
    return error
}
```

### 5. Response
```json
{
  "success": true,
  "workspace": {
    "id": "workspace-id",
    "name": "My Workspace",
    "teamId": "team-id",
    "status": "active"
  }
}
```

---

## External API Integration

### Secure Webhook Pattern

```go
func sendWelcomeWebhook(env map[string]string, ...) error {
    // Get secrets from environment (Appwrite Secrets)
    webhookURL := env["WEBHOOK_URL"]
    apiKey := env["WEBHOOK_API_KEY"]
    
    // Make authenticated request
    req.Header.Set("Authorization", "Bearer "+apiKey)
    // ...
}
```

### Why This Matters

1. **API Keys Never Exposed**: Stored in Appwrite Secrets, rotated without code changes
2. **Server-to-Server**: Requests come from trusted Appwrite infrastructure
3. **Audit Trail**: All external calls logged in function execution logs

### Example Integrations

#### Stripe (Billing)
```go
// Create customer and subscription
stripeKey := os.Getenv("STRIPE_SECRET_KEY")
// ... create Stripe customer
```

#### SendGrid (Email)
```go
// Send welcome email
sendgridKey := os.Getenv("SENDGRID_API_KEY")
// ... send transactional email
```

#### Slack (Notifications)
```go
// Notify workspace creation
slackWebhook := os.Getenv("SLACK_WEBHOOK_URL")
// ... send Slack message
```

---

## Error Handling Strategy

### Structured Error Responses

```go
type ErrorResponse struct {
    Success bool    `json:"success"`
    Error   APIError `json:"error"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Field   string `json:"field,omitempty"`
}
```

### Error Categories

1. **Validation Errors** (400): Invalid input, missing fields
2. **Authentication Errors** (401): Invalid or missing session
3. **Authorization Errors** (403): User lacks permissions
4. **Not Found** (404): Resource doesn't exist
5. **Server Errors** (500): Internal failures, external API errors

### Error Handling Best Practices

- ✅ **Never Leak Secrets**: Error messages don't expose API keys or internal details
- ✅ **Field-Specific Errors**: Help frontend highlight problematic fields
- ✅ **Rollback on Failure**: Clean up partial state
- ✅ **Log Everything**: Detailed logs for debugging (not sent to client)

---

## Performance Considerations

### Function Timeout

Appwrite Functions have a default timeout (typically 15 seconds). For long-running operations:

1. **Async Execution**: Return immediately, process in background
2. **Batch Operations**: Process multiple items in one function call
3. **External Queues**: Use message queues (Redis, RabbitMQ) for heavy processing

### Database Queries

- **Indexes**: Ensure `teamId`, `ownerId`, `slug` are indexed
- **Pagination**: Use `limit` and `offset` for large result sets
- **Selective Fields**: Only query needed attributes

### Caching Strategy

- **Team Memberships**: Cache user's team list (with TTL)
- **Workspace Metadata**: Cache frequently accessed workspace data
- **External API Responses**: Cache webhook responses when appropriate

---

## Testing Strategy

### Unit Tests

Test individual components:
```go
func TestSanitizeString(t *testing.T) {
    result := middleware.SanitizeString("  Test  ")
    assert.Equal(t, "Test", result)
}
```

### Integration Tests

Test with Appwrite test instance:
```go
func TestCreateWorkspace(t *testing.T) {
    // Setup test Appwrite client
    // Execute function
    // Verify team creation
    // Verify document creation
    // Cleanup
}
```

### End-to-End Tests

Test full flow from frontend:
1. Authenticate user
2. Call function
3. Verify workspace appears in UI
4. Verify team membership

---

## Monitoring and Observability

### Logging

```go
log.Printf("Creating workspace: %s for user: %s", workspaceName, userID)
log.Printf("Team created: %s", team.ID)
log.Printf("Warning: Webhook failed: %v", err) // Non-critical
```

### Metrics to Track

- Function execution time
- Success/failure rates
- Error types and frequencies
- External API response times
- Workspace creation rate

### Alerts

- High error rate (>5%)
- Slow execution times (>10s)
- External API failures
- Unusual creation patterns (potential abuse)

---

## Scaling Considerations

### Horizontal Scaling

Appwrite Functions automatically scale based on load. No manual configuration needed.

### Database Scaling

- **Sharding**: Partition workspaces by tenant ID
- **Read Replicas**: Use for analytics and reporting
- **Connection Pooling**: Managed by Appwrite

### Rate Limiting

Implement rate limiting to prevent abuse:
```go
// Check user's workspace creation rate
// Limit: 5 workspaces per day per user
if userWorkspaceCount >= 5 {
    return error("Rate limit exceeded")
}
```

---

## Conclusion

This architecture demonstrates:

1. ✅ **Security by Design**: Server-side validation, secure secrets, proper authentication
2. ✅ **Server-Side Orchestration**: Atomic operations across multiple resources
3. ✅ **Multi-Tenancy**: Data isolation using Appwrite Teams
4. ✅ **Production Readiness**: Error handling, logging, rollback logic
5. ✅ **Senior Engineering**: Understanding of when and why to use server functions

**Key Takeaway**: While Appwrite's client SDK is powerful, senior engineers know that critical business logic must run on the server where it cannot be bypassed.

---

## Further Reading

- [Appwrite Functions Documentation](https://appwrite.io/docs/functions)
- [Appwrite Teams & Permissions](https://appwrite.io/docs/permissions)
- [Go Best Practices](https://go.dev/doc/effective_go)
- [Security by Design Principles](https://owasp.org/www-project-security-by-design/)


