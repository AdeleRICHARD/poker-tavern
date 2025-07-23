#!/bin/bash

# JIRA OAuth Setup Script
# This script helps you set up the required environment variables for JIRA OAuth

echo "🔧 JIRA OAuth Setup"
echo "===================="
echo ""

# Check if .env file exists
if [ -f ".env" ]; then
    echo "📄 Found existing .env file"
    source .env
else
    echo "📄 Creating new .env file"
    touch .env
fi

# Function to read input with default value
read_with_default() {
    local prompt="$1"
    local default="$2"
    local varname="$3"
    
    if [ -n "${!varname}" ]; then
        default="${!varname}"
    fi
    
    echo -n "$prompt"
    if [ -n "$default" ]; then
        echo -n " [current: $default]"
    fi
    echo -n ": "
    
    read input
    if [ -z "$input" ] && [ -n "$default" ]; then
        input="$default"
    fi
    
    export $varname="$input"
}

echo "To set up JIRA OAuth, you need to:"
echo "1. Go to https://developer.atlassian.com/console/myapps"
echo "2. Create a new app or use an existing one"
echo "3. Enable OAuth 2.0 (3LO) authorization"
echo "4. Add callback URLs:"
echo "   - http://localhost:8080/auth/jira/callback (for local development)"
echo "   - http://192.168.35.176:8080/auth/jira/callback (for network access)"
echo "5. Add scopes: read:jira-work read:jira-user"
echo ""

read_with_default "Enter your JIRA OAuth Client ID" "" "JIRA_CLIENT_ID"
read_with_default "Enter your JIRA OAuth Client Secret" "" "JIRA_CLIENT_SECRET"

# Save to .env file
echo "# JIRA OAuth Configuration" > .env
echo "export JIRA_CLIENT_ID=\"$JIRA_CLIENT_ID\"" >> .env
echo "export JIRA_CLIENT_SECRET=\"$JIRA_CLIENT_SECRET\"" >> .env

echo ""
echo "✅ Configuration saved to .env file"
echo ""
echo "To use these variables, run:"
echo "  source .env"
echo "  go run ./cmd/server/main.go"
echo ""
echo "Or in one command:"
echo "  source .env && go run ./cmd/server/main.go"
