# Scaling Multi-Tenant SaaS with Go and Appwrite

> **Senior Engineering Architecture**: A production-ready demonstration of Security by Design and Server-Side Orchestration using Appwrite Functions and Go.

## 🎯 Why This Architecture Matters

In a basic Appwrite application, clients communicate directly with the database. While this works for simple apps, **enterprise SaaS applications** require a different approach:

1. **Security by Design**: Critical business logic must run on the server where it cannot be bypassed
2. **Multi-Tenancy**: Data isolation using Appwrite Teams ensures tenant separation
3. **Server-Side Orchestration**: Complex operations that span multiple resources require admin-level permissions
4. **External Integration**: API keys for third-party services (Stripe, SendGrid, etc.) must never touch the frontend

This project demonstrates how to build a **multi-tenant task management system** where workspace creation, team management, and external webhooks are orchestrated securely on the server side.

---

## 🏗️ Architecture Overview

```
┌─────────────┐
│   Frontend  │  (Next.js / React)
│  (Client)   │
└──────┬──────┘
       │ HTTP Request
       │ (with Session Token)
       ▼
┌─────────────────────────────────────┐
│     Appwrite Function (Go)          │
│  ┌───────────────────────────────┐  │
│  │ 1. Verify Trigger              │  │
│  │ 2. Authenticate User          │  │
│  │ 3. Validate Input             │  │
│  │ 4. Create Team (Multi-Tenant) │  │
│  │ 5. Create Workspace Document  │  │
│  │ 6. Initialize Collections     │  │
│  │ 7. Send External Webhook       │  │
│  └───────────────────────────────┘  │
└──────┬──────────────────────────────┘
       │
       │ Uses Server SDK (Admin Permissions)
       ▼
┌─────────────────────────────────────┐
│         Appwrite Backend            │
│  • Authentication                   │
│  • Teams (Multi-Tenancy)            │
│  • Databases                        │
│  • Secrets (API Keys)               │
└─────────────────────────────────────┘
```

---

## 🔐 Security Model: Trust but Verify

### Why Server SDK Over Client SDK?

**Client SDK Limitations:**
- Users can only perform actions they have explicit permissions for
- Complex orchestration (creating teams + documents + collections) requires multiple API calls
- Business logic validation happens on the client (can be bypassed)
- API keys for external services would be exposed

**Server SDK Advantages:**
- **Admin-level permissions** via Server API Key (stored securely in Appwrite Secrets)
- **Atomic operations** that span multiple resources
- **Server-side validation** that cannot be bypassed
- **Secure external API calls** using secrets that never leave the server

### The "Trust but Verify" Principle

While Appwrite's client-side SDK is powerful and secure, a **senior engineer** understands that:

> **Critical business logic (billing, cross-collection writes, tenant creation) must always be protected by a server-side runtime.**

This function demonstrates:
- ✅ Input validation and sanitization
- ✅ Session verification
- ✅ Permission checks
- ✅ Secure external API integration
- ✅ Error handling and rollback logic

---

## 📁 Project Structure

```
appwrite-go-sdk-tutorial/
├── functions/
│   └── create-workspace/
│       ├── main.go              # Main function handler
│       └── go.mod               # Function dependencies
├── shared/
│   ├── models/
│   │   ├── workspace.go        # Workspace data models
│   │   ├── user.go             # User profile models
│   │   └── errors.go           # Error response models
│   ├── middleware/
│   │   ├── auth.go             # Authentication & validation
│   │   └── teams.go            # Team management helpers
│   └── services/
│       └── webhook.go          # External API integration
├── go.mod                      # Root module
└── README.md                   # This file
```

---

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- Appwrite Cloud or Self-Hosted instance
- Appwrite CLI (for deploying functions)

### Setup

1. **Clone and initialize the project:**
   ```bash
   cd appwrite-go-sdk-tutorial
   go mod download
   ```

2. **Configure Appwrite:**
   - Create an Appwrite project
   - Create a database and collection for workspaces
   - Set up Appwrite Teams (enabled by default)

3. **Set Environment Variables in Appwrite:**
   
   In your Appwrite Function settings, configure these secrets:
   ```
   APPWRITE_ENDPOINT=https://cloud.appwrite.io/v1
   APPWRITE_PROJECT_ID=your-project-id
   APPWRITE_API_KEY=your-server-api-key
   APPWRITE_DATABASE_ID=your-database-id
   APPWRITE_WORKSPACES_COLLECTION_ID=your-collection-id
   WEBHOOK_URL=https://your-webhook-service.com/api/webhooks
   WEBHOOK_API_KEY=your-webhook-api-key
   ```

4. **Deploy the Function:**
   ```bash
   appwrite deploy function
   ```

---

## 🔧 How Multi-Tenancy Works

### Appwrite Teams as Tenant Isolation

Each workspace is backed by an **Appwrite Team**. This provides:

1. **Data Isolation**: Collections use team-based permissions
   ```go
   // Only team members can read
   []string{fmt.Sprintf("read(\"team:%s\")", team.ID)}
   
   // Only team owners can write
   []string{fmt.Sprintf("write(\"team:%s[owner]\")", team.ID)}
   ```

2. **Automatic Access Control**: Appwrite enforces team membership before allowing data access

3. **Scalable Architecture**: Teams can have unlimited members, roles, and permissions

### Custom Tenant IDs

For additional flexibility, we also store a `tenantId` field (using the Team ID) that can be used for:
- Custom query filtering
- Analytics and reporting
- Integration with external systems

---

## 📝 Function Flow: Creating a Workspace

When a user calls `create-workspace`, here's what happens:

1. **Trigger Verification**: Ensures the request comes from Appwrite
2. **Authentication**: Validates the user's session token
3. **Input Validation**: Sanitizes and validates the workspace name
4. **Team Creation**: Creates an Appwrite Team (the tenant boundary)
5. **Document Creation**: Creates a workspace document with team-scoped permissions
6. **Collection Initialization**: Sets up private collections for workspace data
7. **External Webhook**: Sends a welcome notification (using secure API keys)
8. **Response**: Returns the created workspace with success status

### Error Handling

The function includes comprehensive error handling:
- **Validation errors**: Return 400 with field-specific messages
- **Authentication errors**: Return 401
- **Creation failures**: Rollback team creation if document creation fails
- **Non-critical errors**: Log warnings but continue (e.g., webhook failures)

---

## 🔌 External API Integration

### Secure Webhook Communication

The `webhook.go` service demonstrates how to securely call external APIs:

```go
// API keys are stored in Appwrite Secrets (never in code)
webhookURL := os.Getenv("WEBHOOK_URL")
apiKey := os.Getenv("WEBHOOK_API_KEY")

// Make authenticated request
req.Header.Set("Authorization", "Bearer "+apiKey)
```

**Why this matters:**
- API keys never appear in client-side code
- Keys are rotated via Appwrite Secrets without code changes
- Requests are made from a trusted server environment

### Example Integrations

- **Stripe**: Create customers and subscriptions
- **SendGrid**: Send transactional emails
- **Slack**: Notify workspace creation
- **Analytics**: Track workspace metrics

---

## 🧪 Testing the Function

### Local Testing

1. **Set up test environment variables:**
   ```bash
   export APPWRITE_ENDPOINT="https://cloud.appwrite.io/v1"
   export APPWRITE_PROJECT_ID="test-project"
   export APPWRITE_API_KEY="test-key"
   ```

2. **Create a test request:**
   ```json
   {
     "headers": {
       "x-appwrite-trigger": "http",
       "x-appwrite-session": "user-session-token"
     },
     "body": "{\"name\": \"My Workspace\", \"plan\": \"free\"}",
     "env": {
       "APPWRITE_ENDPOINT": "https://cloud.appwrite.io/v1",
       "APPWRITE_PROJECT_ID": "your-project-id",
       "APPWRITE_API_KEY": "your-server-key"
     }
   }
   ```

3. **Run the function:**
   ```bash
   cd functions/create-workspace
   go run main.go < test-request.json
   ```

---

## 📚 Key Concepts for Senior Engineers

### 1. Middleware Patterns

The `middleware` package provides reusable components:
- **Authentication**: Verify user sessions
- **Validation**: Sanitize and validate inputs
- **Team Management**: Helper functions for team operations

### 2. Type Safety

Go structs ensure type safety across the application:
- Request/response models
- Error structures
- Workspace and user profiles

### 3. Error Handling

Structured error responses help frontend applications handle failures gracefully:
```go
ErrorResponse{
    Success: false,
    Error: APIError{
        Code:    "VALIDATION_ERROR",
        Message: "Workspace name is required",
        Field:   "name",
    },
}
```

### 4. Security Best Practices

- ✅ Input sanitization
- ✅ Length validation
- ✅ Regex pattern matching
- ✅ Server-side API key management
- ✅ Permission-based access control

---

## 🎓 Why This Demonstrates Senior-Level Skills

### Security by Design
- Server-side validation that cannot be bypassed
- Secure secret management via Appwrite Secrets
- Proper authentication and authorization

### Server-Side Orchestration
- Atomic operations across multiple Appwrite resources
- Rollback logic for failed operations
- Admin-level permissions used appropriately

### Multi-Tenancy
- Data isolation using Appwrite Teams
- Scalable architecture for enterprise SaaS
- Proper permission scoping

### Production Readiness
- Comprehensive error handling
- Logging and monitoring hooks
- Type-safe code structure
- Reusable middleware patterns

---

## 🔄 Next Steps

To extend this architecture:

1. **Add Billing Integration**: Integrate Stripe for subscription management
2. **Implement Workspace Limits**: Enforce plan-based limits (users, storage, etc.)
3. **Add Audit Logging**: Track all workspace operations
4. **Create More Functions**: 
   - `invite-user`: Add users to workspaces
   - `upgrade-plan`: Handle subscription upgrades
   - `delete-workspace`: Safe workspace deletion with cleanup

---

## 📖 Additional Resources

- [Appwrite Go SDK Documentation](https://appwrite.io/docs/sdks/go)
- [Appwrite Functions Guide](https://appwrite.io/docs/functions)
- [Appwrite Teams & Permissions](https://appwrite.io/docs/permissions)
- [Appwrite Secrets Management](https://appwrite.io/docs/functions/secrets)

---

## 🤝 Contributing

This is a demonstration project for senior engineering roles. Feel free to:
- Extend the architecture
- Add more functions
- Improve error handling
- Add comprehensive tests

---

## 📄 License

This project is provided as a demonstration of senior-level Appwrite architecture patterns.

---

**Built with ❤️ for Senior Engineers who understand Security by Design**


