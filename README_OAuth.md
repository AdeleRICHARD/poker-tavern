# Clean OAuth Pattern Implementation

This implementation provides a professional OAuth integration pattern for JIRA authentication, inspired by PlanItPoker's approach. The pattern uses:

- **Backend**: Clean API endpoints with proper session management
- **Frontend**: Popup-based OAuth flow with promise-based authentication

## Backend Implementation

### Key Endpoints

1. **`/api/jira/auth-url`** - Returns the OAuth authorization URL
2. **`/auth/jira/callback`** - Handles OAuth callback and exchanges code for token
3. **`/jira/status`** - Check authentication status
4. **`/jira/search`** - Search for JIRA issues (protected)

### Features

- Server-side session management (in-memory map)
- User information fetching from Atlassian API
- Proper error handling and security
- PostMessage communication with frontend

## Frontend Implementation

### SimpleJiraAuth Class

```typescript
import SimpleJiraAuth from '@/auth/SimpleJiraAuth';

const jiraAuth = new SimpleJiraAuth();

// Authenticate user
try {
  const userData = await jiraAuth.authenticate(sessionId);
  console.log('User authenticated:', userData);
} catch (error) {
  console.error('Authentication failed:', error);
}

// Check authentication status
const isAuthenticated = await jiraAuth.isAuthenticated(sessionId);

// Search issues
const issues = await jiraAuth.searchIssues(sessionId, 'project = PROJ');
```

### Vue Component Integration

The `JiraImport.vue` component demonstrates how to:

1. Check authentication status on mount
2. Handle OAuth flow with popup
3. Search and import issues
4. Handle disconnection

## Setup Instructions

### 1. Backend Setup

**Environment Variables:**
```bash
export JIRA_CLIENT_ID="your-atlassian-oauth-client-id"
export JIRA_CLIENT_SECRET="your-atlassian-oauth-client-secret"
```

**Start the server:**
```bash
cd auth-api
go run ./cmd/server/main.go
```

### 2. Frontend Setup

**Start the development server:**
```bash
cd front-poker-tavern
npm run dev
```

### 3. Atlassian App Configuration

Configure your Atlassian app with:
- **Callback URL**: `http://localhost:8080/auth/jira/callback`
- **Scopes**: `read:jira-work read:jira-user`

## Usage Flow

1. User clicks "Connect to JIRA"
2. Frontend fetches OAuth URL from `/api/jira/auth-url`
3. Popup opens with Atlassian OAuth page
4. User authorizes the application
5. Callback exchanges code for token and fetches user info
6. PostMessage sent to parent window with user data
7. Frontend resolves promise and updates UI
8. User can now search and import JIRA issues

## Security Features

- State parameter validation
- Origin verification for PostMessage
- Server-side session management
- Popup timeout handling (5 minutes)

## Error Handling

- Network errors
- OAuth denial/cancellation
- Popup blocking
- Invalid state parameters
- API rate limiting

## Benefits of This Pattern

1. **Professional**: Follows industry best practices
2. **Secure**: Proper token handling and validation
3. **User-friendly**: Popup-based flow doesn't redirect away
4. **Maintainable**: Clean separation of concerns
5. **Scalable**: Easy to extend with additional OAuth providers

This pattern provides a much cleaner and more professional implementation compared to direct OAuth URL redirection patterns.
