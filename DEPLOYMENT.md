# Deployment Guide

## Prerequisites

1. **Appwrite Instance**: Cloud or self-hosted
2. **Appwrite CLI**: Install from [appwrite.io/docs/command-line](https://appwrite.io/docs/command-line)
3. **Go 1.21+**: For local development and testing

## Step 1: Set Up Appwrite Project

1. Create a new Appwrite project
2. Enable **Teams** feature (for multi-tenancy)
3. Create a **Database**
4. Create a **Collection** named `workspaces` with the following attributes:
   - `name` (String, required)
   - `slug` (String, required, unique)
   - `teamId` (String, required)
   - `ownerId` (String, required)
   - `status` (String, enum: active, suspended, archived)
   - `plan` (String, enum: free, pro, enterprise)
   - `tenantId` (String, required)
   - `createdAt` (String, required)
   - `description` (String, optional)

5. Set collection permissions:
   - **Read**: `team` (team members can read)
   - **Write**: `team[owner]` (only owners can write)

## Step 2: Configure Appwrite Secrets

In your Appwrite project dashboard:

1. Go to **Functions** → **Settings**
2. Add the following **Secrets**:
   - `APPWRITE_ENDPOINT`: Your Appwrite endpoint (e.g., `https://cloud.appwrite.io/v1`)
   - `APPWRITE_PROJECT_ID`: Your project ID
   - `APPWRITE_API_KEY`: Your **Server API Key** (not the Client SDK key!)
   - `APPWRITE_DATABASE_ID`: Your database ID
   - `APPWRITE_WORKSPACES_COLLECTION_ID`: Your collection ID
   - `WEBHOOK_URL`: (Optional) External webhook URL
   - `WEBHOOK_API_KEY`: (Optional) Webhook API key

**⚠️ Important**: Use the **Server API Key**, not the Client SDK key. The Server API Key has admin permissions needed for creating teams and documents.

## Step 3: Install Dependencies

```bash
# Install root dependencies
go mod download

# Install function dependencies
cd functions/create-workspace
go mod download
cd ../..
```

## Step 4: Deploy Function

### Using Appwrite CLI

1. **Login to Appwrite:**
   ```bash
   appwrite login
   ```

2. **Initialize project (if not already done):**
   ```bash
   appwrite init project
   ```

3. **Deploy the function:**
   ```bash
   appwrite deploy function create-workspace
   ```

### Manual Deployment

1. **Build the function:**
   ```bash
   cd functions/create-workspace
   go build -o main .
   ```

2. **Create function in Appwrite Dashboard:**
   - Go to **Functions** → **Create Function**
   - Name: `create-workspace`
   - Runtime: `go-1.21`
   - Upload the `main` binary

3. **Set environment variables** (as configured in Step 2)

## Step 5: Test the Function

### Using Appwrite Console

1. Go to **Functions** → `create-workspace` → **Execute**
2. Use the test interface with:
   ```json
   {
     "name": "Test Workspace",
     "plan": "free"
   }
   ```

### Using cURL

```bash
curl -X POST \
  'https://cloud.appwrite.io/v1/functions/YOUR_FUNCTION_ID/executions' \
  -H 'X-Appwrite-Project: YOUR_PROJECT_ID' \
  -H 'X-Appwrite-Key: YOUR_SERVER_KEY' \
  -H 'Content-Type: application/json' \
  -d '{
    "data": "{\"name\": \"Test Workspace\", \"plan\": \"free\"}"
  }'
```

### From Frontend (Next.js Example)

```typescript
import { Client, Functions } from 'appwrite';

const client = new Client()
  .setEndpoint('https://cloud.appwrite.io/v1')
  .setProject('YOUR_PROJECT_ID');

const functions = new Functions(client);

// Execute function with user session
const response = await functions.createExecution(
  'create-workspace',
  JSON.stringify({
    name: 'My Workspace',
    plan: 'free'
  }),
  false, // async
  '/path/to/your/x-appwrite-session-header' // session token
);
```

## Step 6: Verify Deployment

1. **Check function logs:**
   - Go to **Functions** → `create-workspace` → **Logs`
   - Look for successful executions

2. **Verify team creation:**
   - Go to **Teams** in Appwrite dashboard
   - You should see a new team created for each workspace

3. **Verify document creation:**
   - Go to **Databases** → `workspaces` collection
   - You should see workspace documents with proper permissions

## Troubleshooting

### Function Fails with "Authentication Failed"

- **Issue**: Missing or invalid session token
- **Solution**: Ensure the frontend passes `x-appwrite-session` header

### Function Fails with "Team Creation Failed"

- **Issue**: Insufficient permissions
- **Solution**: Verify you're using the **Server API Key**, not Client SDK key

### Function Fails with "Collection Not Found"

- **Issue**: Wrong collection ID or database ID
- **Solution**: Double-check environment variables in Appwrite Secrets

### Build Errors

- **Issue**: Missing dependencies
- **Solution**: Run `go mod download` in both root and function directories

## Production Considerations

1. **Enable Logging**: Keep function logs enabled for debugging
2. **Set Timeout**: Adjust function timeout based on external API response times
3. **Monitor Usage**: Track function executions and errors
4. **Rate Limiting**: Implement rate limiting for workspace creation
5. **Backup Strategy**: Regular backups of database and teams

## Security Checklist

- ✅ Server API Key stored in Appwrite Secrets (never in code)
- ✅ Input validation and sanitization
- ✅ Session verification on every request
- ✅ Team-based permissions for data isolation
- ✅ Error messages don't leak sensitive information
- ✅ External API keys stored securely

---

**Need Help?** Check the [Appwrite Documentation](https://appwrite.io/docs) or open an issue.


