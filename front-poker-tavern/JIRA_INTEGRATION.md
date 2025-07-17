# JIRA Integration Setup

The Planning Poker Tavern now supports importing JIRA issues directly into your estimation sessions!

## Prerequisites

1. **JIRA Account**: You need access to a JIRA instance (Cloud or Server)
2. **API Token**: Generate a personal API token from your JIRA account
3. **Active Session**: You must be in an active Planning Poker session

## Setting Up JIRA API Access

### For JIRA Cloud (Atlassian Cloud)

1. Go to [Atlassian Account Settings](https://id.atlassian.com/manage-profile/security/api-tokens)
2. Click "Create API token"
3. Give it a descriptive name (e.g., "Planning Poker Tavern")
4. Copy the generated token (you won't be able to see it again!)

### For JIRA Server/Data Center

1. Go to your JIRA instance settings
2. Navigate to Personal Settings > Personal Access Tokens
3. Create a new token with appropriate permissions
4. Copy the token for use

## Using the JIRA Import Feature

1. **Create or Join a Session**: Start by creating a new session or joining an existing one
2. **Open JIRA Import**: Look for the "📋 JIRA Integration" section in the left panel
3. **Click "Import JIRA Issues"**: This will open the import form
4. **Fill in the Required Fields**:
   - **JIRA URL**: Your JIRA instance URL (e.g., `https://yourcompany.atlassian.net`)
   - **Username/Email**: Your JIRA account email
   - **API Token**: The token you generated earlier
   - **JQL Query**: A JQL query to filter the issues you want to import

## JQL Query Examples

Here are some common JQL queries you can use:

### Basic Examples
```jql
project = PROJ AND status = "To Do"
```

```jql
project = MYPROJECT AND sprint in openSprints()
```

```jql
project = PROJ AND assignee = currentUser() AND status != Done
```

### Advanced Examples
```jql
project = PROJ AND sprint = "Sprint 1" AND type = Story
```

```jql
project in (PROJ1, PROJ2) AND status = "Ready for Estimation"
```

```jql
project = PROJ AND created >= -7d AND type != Epic
```

## Security Notes

- **API Tokens**: Never share your API tokens. They provide access to your JIRA account
- **Local Storage**: Your API tokens are NOT stored by the application
- **Session Only**: Credentials are only used for the import request and discarded immediately
- **Team Sharing**: Only the imported stories (not credentials) are shared with your team

## Troubleshooting

### Common Issues

1. **"Failed to connect to JIRA"**: 
   - Check your JIRA URL format
   - Ensure your API token is valid
   - Verify your account has access to the JIRA instance

2. **"Failed to fetch JIRA issues"**:
   - Check your JQL query syntax
   - Ensure you have permission to view the issues
   - Try a simpler query first

3. **"No issues found"**:
   - Your JQL query might be too restrictive
   - Try removing some filters
   - Check if the project key is correct

### Testing Your Setup

Start with a simple query like:
```jql
project = YOURPROJECT ORDER BY created DESC
```

This will import the most recently created issues from your project.

## Features

- **Real-time Import**: Issues are imported and immediately visible to all session participants
- **JIRA Key Display**: Each imported story shows its JIRA key for easy reference
- **Collaborative Estimation**: Once imported, stories can be estimated like any other story
- **WebSocket Updates**: All team members see imported stories in real-time

## Limits

- **Maximum Import**: 50 issues per import (for performance)
- **Rate Limiting**: Respects JIRA API rate limits
- **Permissions**: Only issues you can view in JIRA will be imported

## Next Steps

After importing JIRA issues:
1. Select your character class
2. Navigate through the imported stories
3. Use the poker cards to estimate each story
4. Reveal votes when everyone has estimated
5. Discuss and re-estimate if needed

The estimated stories will include the JIRA key so you can easily reference them back in your JIRA system!
