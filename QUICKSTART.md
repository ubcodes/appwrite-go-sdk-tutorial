# Quick Start Guide

Get your multi-tenant SaaS backend running in 5 minutes.

## Prerequisites

- Appwrite account (cloud or self-hosted)
- Go 1.21+
- Appwrite CLI (optional, for deployment)

## 1. Clone and Setup

```bash
cd appwrite-go-sdk-tutorial
go mod download
cd functions/create-workspace
go mod download
```

## 2. Configure Appwrite

### Create Database and Collection

1. In Appwrite Dashboard → **Databases** → **Create Database**
2. Create a collection named `workspaces`
3. Add these attributes:
   - `name` (String, required, 255 chars)
   - `slug` (String, required, unique, 255 chars)
   - `teamId` (String, required)
   - `ownerId` (String, required)
   - `status` (String, enum: `active`, `suspended`, `archived`)
   - `plan` (String, enum: `free`, `pro`, `enterprise`)
   - `tenantId` (String, required)
   - `createdAt` (String, required)
   - `description` (String, optional)

4. Set collection permissions:
   - **Read**: `team` (any team member)
   - **Write**: `team[owner]` (team owners only)

### Get Your API Keys

1. **Server API Key**: Dashboard → **Settings** → **API Keys** → **Create API Key** → Select **Server** scope
2. **Project ID**: Dashboard → **Settings** → **General** → Copy Project ID
3. **Database ID**: Databases → Your database → Copy ID
4. **Collection ID**: Databases → Your collection → Copy ID

## 3. Deploy Function

### Option A: Using Appwrite CLI

```bash
# Login
appwrite login

# Deploy
appwrite deploy function create-workspace
```

### Option B: Manual Deployment

1. Build the function:
   ```bash
   cd functions/create-workspace
   go build -o main .
   ```

2. In Appwrite Dashboard:
   - **Functions** → **Create Function**
   - Name: `create-workspace`
   - Runtime: `go-1.21`
   - Upload the `main` binary

3. Add environment variables (Secrets):
   - `APPWRITE_ENDPOINT` = `https://cloud.appwrite.io/v1` (or your self-hosted URL)
   - `APPWRITE_PROJECT_ID` = Your project ID
   - `APPWRITE_API_KEY` = Your **Server API Key** (not Client SDK key!)
   - `APPWRITE_DATABASE_ID` = Your database ID
   - `APPWRITE_WORKSPACES_COLLECTION_ID` = Your collection ID
   - `WEBHOOK_URL` = (Optional) Your webhook endpoint
   - `WEBHOOK_API_KEY` = (Optional) Webhook API key

## 4. Test the Function

### From Appwrite Console

1. Go to **Functions** → `create-workspace` → **Execute**
2. Enter test data:
   ```json
   {
     "name": "My First Workspace",
     "plan": "free"
   }
   ```
3. Click **Execute**
4. Check the response - you should see a workspace object with `teamId`

### Verify Results

1. **Check Teams**: Dashboard → **Teams** → You should see a new team
2. **Check Database**: Dashboard → **Databases** → `workspaces` → You should see a new document

## 5. Integrate with Frontend

See `examples/frontend-integration.ts` for a complete Next.js/React example.

Basic usage:
```typescript
import { Client, Functions } from 'appwrite';

const client = new Client()
  .setEndpoint('https://cloud.appwrite.io/v1')
  .setProject('YOUR_PROJECT_ID');

const functions = new Functions(client);

const response = await functions.createExecution(
  'create-workspace',
  JSON.stringify({ name: 'My Workspace', plan: 'free' }),
  false
);
```

## Troubleshooting

### "Authentication Failed"
- Ensure you're passing a valid user session token
- Check that the user is logged in

### "Team Creation Failed"
- Verify you're using the **Server API Key**, not Client SDK key
- Check that Teams feature is enabled in your project

### "Collection Not Found"
- Double-check your `APPWRITE_DATABASE_ID` and `APPWRITE_WORKSPACES_COLLECTION_ID`
- Ensure the collection exists and has the correct attributes

### Build Errors
- Run `go mod download` in both root and function directories
- Ensure Go 1.21+ is installed

## Next Steps

- Read [README.md](./README.md) for architecture overview
- Read [ARCHITECTURE.md](./ARCHITECTURE.md) for deep dive
- Read [DEPLOYMENT.md](./DEPLOYMENT.md) for production deployment
- Extend the function with additional features (billing, user invites, etc.)

## Support

- [Appwrite Documentation](https://appwrite.io/docs)
- [Appwrite Discord](https://discord.gg/appwrite)
- [Go SDK Reference](https://appwrite.io/docs/sdks/go)

---

**Ready to build?** Start with the [README](./README.md) for the full architecture explanation.


