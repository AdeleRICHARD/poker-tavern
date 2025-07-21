package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// Session represents a Planning Poker session.
type Session struct {
	SessionID    string        `json:"sessionId"`
	Name         string        `json:"name"`
	Players      []string      `json:"players"`
	Stories      []Story       `json:"stories"`
	ChatMessages []ChatMessage `json:"chatMessages"`
}

// Story represents an imported story (e.g., from Jira).
type Story struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	JiraKey     string `json:"jiraKey,omitempty"`
}

// ChatMessage represents a chat entry.
type ChatMessage struct {
	Author    string `json:"author"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}

// JiraConnection stores JIRA connection info per session
type JiraConnection struct {
	JiraURL   string
	Username  string
	APIToken  string
	SessionID string
}

// sessions stores the created sessions in memory.
var (
	sessions      = make(map[string]*Session)
	sessionsMutex = &sync.Mutex{}

	// wsHubs manages WebSocket hubs per session for real-time broadcasting.
	wsHubs      = make(map[string]*Hub)
	wsHubsMutex = &sync.Mutex{}

	// JIRA connections per session
	jiraConnections      = make(map[string]*JiraConnection)
	jiraConnectionsMutex = &sync.Mutex{}
)

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.Mutex
}

// newHub creates a new Hub.
func newHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// run starts the hub's event loop in a goroutine.
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mutex.Unlock()
		case message := <-h.broadcast:
			h.mutex.Lock()
			for client := range h.clients {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("WebSocket write error: %v", err)
					delete(h.clients, client)
					client.Close()
				}
			}
			h.mutex.Unlock()
		}
	}
}

// getOrCreateHub retrieves or creates a hub for a session.
func getOrCreateHub(sessionID string) *Hub {
	wsHubsMutex.Lock()
	defer wsHubsMutex.Unlock()
	if hub, ok := wsHubs[sessionID]; ok {
		return hub
	}
	hub := newHub()
	go hub.run()
	wsHubs[sessionID] = hub
	return hub
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Define endpoints.
	http.HandleFunc("/health", healthHandler)                  // Health check endpoint.
	http.HandleFunc("/session", handleSession)                 // POST to create, GET to retrieve session.
	http.HandleFunc("/session/join", joinSessionHandler)       // POST request to join a session.
	http.HandleFunc("/session/import-jira", importJiraHandler) // POST request to import Jira issues.
	http.HandleFunc("/ws", wsHandler)                          // WebSocket endpoint for real-time updates (including chat).
	
	// JIRA integration endpoints
	http.HandleFunc("/api/jira/auth-url", jiraAuthUrlHandler)   // Get JIRA OAuth URL
	http.HandleFunc("/auth/jira/callback", jiraCallbackHandler) // JIRA OAuth callback
	http.HandleFunc("/jira/status", jiraStatusHandler)         // JIRA connection status
	http.HandleFunc("/jira/search", jiraSearchHandler)         // JIRA issue search
	
	// Demo OAuth endpoint for testing popup flow
	http.HandleFunc("/demo/oauth", demoOAuthHandler)

	// Wrap DefaultServeMux with CORS middleware.
	handler := corsMiddleware(http.DefaultServeMux)

	addr := ":8080"
	log.Printf("Starting server on %s\n", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// healthHandler returns a simple JSON status.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{"status": "ok"}
	json.NewEncoder(w).Encode(resp)
}

// handleSession handles session creation and retrieval.
// POST to create; GET to retrieve by ID
func handleSession(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Create session
		var payload struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if payload.Name == "" {
			http.Error(w, "Session name is required", http.StatusBadRequest)
			return
		}

		// Generate a unique session ID.
		sessionID := uuid.New().String()
		session := &Session{
			SessionID:    sessionID,
			Name:         payload.Name,
			Players:      []string{},
			Stories:      []Story{},
			ChatMessages: []ChatMessage{},
		}

		sessionsMutex.Lock()
		sessions[sessionID] = session
		sessionsMutex.Unlock()

		// Initialize WebSocket hub for the new session.
		getOrCreateHub(sessionID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(session)
		return
	}

	if r.Method == http.MethodGet {
		// Get session by ID
		sessionID := r.URL.Query().Get("sessionId")
		if sessionID == "" {
			http.Error(w, "sessionId query param required", http.StatusBadRequest)
			return
		}

		sessionsMutex.Lock()
		session, ok := sessions[sessionID]
		sessionsMutex.Unlock()
		if !ok {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// joinSessionHandler allows a player to join an existing session.
// Expects a POST request with JSON body: {"sessionId": "xxx", "playerName": "name"}
func joinSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		SessionID  string `json:"sessionId"`
		PlayerName string `json:"playerName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if payload.SessionID == "" || payload.PlayerName == "" {
		http.Error(w, "Both sessionId and playerName are required", http.StatusBadRequest)
		return
	}

	sessionsMutex.Lock()
	session, ok := sessions[payload.SessionID]
	if !ok {
		sessionsMutex.Unlock()
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Add player if not already present.
	playerExists := slices.Contains(session.Players, payload.PlayerName)
	if !playerExists {
		session.Players = append(session.Players, payload.PlayerName)
	}
	sessionsMutex.Unlock()

	// Broadcast the updated session to WebSocket clients (in a goroutine for non-blocking).
	go func() {
		hub := getOrCreateHub(payload.SessionID)
		message := map[string]interface{}{
			"type":    "player_joined",
			"session": session,
		}
		msgBytes, _ := json.Marshal(message)
		hub.broadcast <- msgBytes
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// jiraAuthUrlHandler returns the OAuth URL for JIRA authentication (used by popup flow)
func jiraAuthUrlHandler(w http.ResponseWriter, r *http.Request) {
	// Expect a session ID as a query parameter
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	// OAuth 2.0 configuration - get from environment variables
	clientID := os.Getenv("JIRA_CLIENT_ID")
	clientSecret := os.Getenv("JIRA_CLIENT_SECRET")
	demoMode := os.Getenv("DEMO_MODE") == "true"
	
	// Check if demo mode is enabled
	if demoMode {
		// Return demo OAuth URL for testing popup flow
		demoURL := fmt.Sprintf("http://localhost:8080/demo/oauth?sessionId=%s", sessionID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"authUrl": demoURL,
			"demo": "true",
		})
		return
	}
	
	if clientID == "" || clientSecret == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "JIRA OAuth not configured",
			"message": "Please set JIRA_CLIENT_ID and JIRA_CLIENT_SECRET environment variables",
			"setup_instructions": "Run ./setup_oauth.sh to configure OAuth credentials or set DEMO_MODE=true",
		})
		return
	}
	redirectURI := "http://localhost:8080/auth/jira/callback"
	scopes := "offline_access read:jira-work read:jira-user read:account read:me read:issue:jira read:project:jira"

	// Generate state parameter for security (store sessionID in state)
	state := fmt.Sprintf("session_%s", sessionID)

	// Build OAuth authorization URL with proper encoding
	authURL := fmt.Sprintf(
		"https://auth.atlassian.com/authorize?audience=api.atlassian.com&client_id=%s&scope=%s&redirect_uri=%s&state=%s&response_type=code&prompt=consent",
		url.QueryEscape(clientID),
		url.QueryEscape(scopes),
		url.QueryEscape(redirectURI),
		url.QueryEscape(state),
	)

	// Return the OAuth URL as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"authUrl": authURL,
	})
}

// jiraCallbackHandler handles OAuth 2.0 callback from JIRA
func jiraCallbackHandler(w http.ResponseWriter, r *http.Request) {
    // Log the full request for debugging
    log.Printf("OAuth callback received: %s", r.URL.String())
    log.Printf("Query parameters: %v", r.URL.Query())
    
    // Extract code and state from query
    code := r.URL.Query().Get("code")
    state := r.URL.Query().Get("state")
    
    log.Printf("Code: '%s', State: '%s'", code, state)
    
    if code == "" || state == "" {
        // Check if there's an error parameter
        errorParam := r.URL.Query().Get("error")
        errorDesc := r.URL.Query().Get("error_description")
        
        log.Printf("OAuth error - Code: '%s', State: '%s', Error: '%s', Description: '%s'", 
            code, state, errorParam, errorDesc)
        
        if errorParam != "" {
            http.Error(w, fmt.Sprintf("OAuth error: %s - %s", errorParam, errorDesc), http.StatusBadRequest)
        } else {
            http.Error(w, "Code or state parameter missing", http.StatusBadRequest)
        }
        return
    }

    // Parse state to retrieve session ID
    parts := strings.Split(state, "_")
    if len(parts) != 2 {
        http.Error(w, "Invalid state parameter", http.StatusBadRequest)
        return
    }
    sessionID := parts[1]

    // Exchange code for access token
    clientID := os.Getenv("JIRA_CLIENT_ID")
    clientSecret := os.Getenv("JIRA_CLIENT_SECRET")
    if clientID == "" || clientSecret == "" {
        http.Error(w, "JIRA OAuth not configured", http.StatusInternalServerError)
        return
    }
    
    data := url.Values{}
    data.Set("grant_type", "authorization_code")
    data.Set("client_id", clientID)
    data.Set("client_secret", clientSecret)
    data.Set("code", code)
    data.Set("redirect_uri", "http://localhost:8080/auth/jira/callback")

    tokenResp, err := http.Post(
        "https://auth.atlassian.com/oauth/token",
        "application/x-www-form-urlencoded",
        strings.NewReader(data.Encode()))

    if err != nil {
        http.Error(w, "Failed to exchange code: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer tokenResp.Body.Close()

    var tokenData struct {
        AccessToken string `json:"access_token"`
        Scope       string `json:"scope"`
        ExpiresIn   int    `json:"expires_in"`
        TokenType   string `json:"token_type"`
    }
    if err := json.NewDecoder(tokenResp.Body).Decode(&tokenData); err != nil {
        http.Error(w, "Invalid token response", http.StatusInternalServerError)
        return
    }
    
    // Log token details for debugging
    log.Printf("Token received - Type: %s, Expires in: %d seconds", tokenData.TokenType, tokenData.ExpiresIn)
    log.Printf("Granted scopes: %s", tokenData.Scope)

    // Fetch user information from Atlassian API
    userReq, err := http.NewRequest("GET", "https://api.atlassian.com/me", nil)
    if err != nil {
        http.Error(w, "Failed to create user request", http.StatusInternalServerError)
        return
    }
    userReq.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)
    
    userResp, err := http.DefaultClient.Do(userReq)
    if err != nil {
        http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
        return
    }
    defer userResp.Body.Close()
    
    var userData struct {
        AccountID string `json:"account_id"`
        Name      string `json:"name"`
        Email     string `json:"email"`
        Picture   string `json:"picture"`
    }
    if err := json.NewDecoder(userResp.Body).Decode(&userData); err != nil {
        http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
        return
    }

    // Save the token information for the relevant session
    jiraConnectionsMutex.Lock()
    jiraConnections[sessionID] = &JiraConnection{
        SessionID: sessionID,
        APIToken:  tokenData.AccessToken, // Use access token for API requests
        Username:  userData.Email,
        JiraURL:   "your-jira-instance", // Usually dynamically determined
    }
    jiraConnectionsMutex.Unlock()

    // Create user data for postMessage
    userDataJSON, _ := json.Marshal(userData)

    // Send success message to parent window and close popup
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(fmt.Sprintf(`
        <html>
        <body>
        <script>
            window.opener.postMessage({
                type: 'JIRA_AUTH_SUCCESS',
                user: %s
            }, 'http://localhost:5173');
            window.close();
        </script>
        <p>JIRA connection successful! You can close this window.</p>
        </body>
        </html>
    `, userDataJSON)))
}

// jiraStatusHandler checks if a session is connected to JIRA
func jiraStatusHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	jiraConnectionsMutex.Lock()
	_, connExists := jiraConnections[sessionID]
	jiraConnectionsMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"connected": connExists,
	})
}

// jiraSearchHandler enables searching for issues in JIRA
func jiraSearchHandler(w http.ResponseWriter, r *http.Request) {
	// Expect a session ID as a query parameter
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	// Validate that the session is connected to JIRA
	jiraConnectionsMutex.Lock()
	conn, connExists := jiraConnections[sessionID]
	jiraConnectionsMutex.Unlock()
	if !connExists {
		http.Error(w, "JIRA connection not established", http.StatusUnauthorized)
		return
	}

	// Get search query param
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query param required", http.StatusBadRequest)
		return
	}

	// Check if we're in demo mode
	demoMode := os.Getenv("DEMO_MODE") == "true"
	if demoMode {
		// Return demo search results
		mockedIssues := []Story{
			{ID: "DEMO-1", Title: "[DEMO] Fix login issue", Description: "Users cannot log in to the application", JiraKey: "DEMO-1"},
			{ID: "DEMO-2", Title: "[DEMO] Add user profile page", Description: "Create a user profile page with basic info", JiraKey: "DEMO-2"},
			{ID: "DEMO-3", Title: "[DEMO] Performance optimization", Description: "Improve application loading time", JiraKey: "DEMO-3"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"issues": mockedIssues,
		})
		return
	}

	// Real JIRA search implementation
	issues, err := searchJiraIssues(conn, query)
	if err != nil {
		log.Printf("JIRA search error: %v", err)
		http.Error(w, "Failed to search JIRA issues: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"issues": issues,
	})
}

// searchJiraIssues searches for JIRA issues using the Atlassian API with OAuth token
func searchJiraIssues(conn *JiraConnection, query string) ([]Story, error) {
	// Get accessible resources (sites) for the user
	// Log the token being used for debugging
	log.Printf("Using access token for API call: %s", conn.APIToken[:20]+"...")

	// Log the request details
	log.Printf("Making request to accessible-resources endpoint with Authorization header")

	req, err := http.NewRequest("GET", "https://api.atlassian.com/oauth/token/accessible-resources", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+conn.APIToken)
	req.Header.Set("Accept", "application/json")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get accessible resources: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to get accessible resources: %d %s", resp.StatusCode, resp.Status)
		return nil, fmt.Errorf("failed to get accessible resources: %s", resp.Status)
	}
	
	var resources []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
		return nil, fmt.Errorf("failed to decode resources: %v", err)
	}
	
	// Log the accessible resources for debugging
	log.Printf("Found %d accessible resources:", len(resources))
	for i, res := range resources {
		log.Printf("  [%d] ID: %s, Name: %s, URL: %s", i, res.ID, res.Name, res.URL)
	}
	
	if len(resources) == 0 {
		return nil, fmt.Errorf("no accessible JIRA resources found")
	}
	
	// Use the first resource (most common case)
	resource := resources[0]
	log.Printf("Using resource - ID: %s, Name: %s, URL: %s", resource.ID, resource.Name, resource.URL)
	
	// Build JQL query - handle different input types
	jql := query
	
	// Check if the query is a JIRA URL and extract relevant info
	if strings.Contains(query, "atlassian.net") {
		// Extract project key if possible from URL like https://fcms.atlassian.net/jira/software/projects/IMMO/boards/280?selectedIssue=IMMO-5295
		if strings.Contains(query, "selectedIssue=") {
			// Extract the issue key from selectedIssue parameter
			parts := strings.Split(query, "selectedIssue=")
			if len(parts) > 1 {
				issueKey := strings.Split(parts[1], "&")[0] // Remove any additional params
				jql = fmt.Sprintf("key = %s", issueKey)
			}
		} else if strings.Contains(query, "/projects/") {
			// Extract project key from URL path
			parts := strings.Split(query, "/projects/")
			if len(parts) > 1 {
				projectKey := strings.Split(parts[1], "/")[0]
				jql = fmt.Sprintf("project = %s", projectKey)
			}
		} else {
			// Fallback to simple recent issues query
			jql = "ORDER BY updated DESC"
		}
	} else if !strings.Contains(query, "=") && !strings.Contains(query, "~") && !strings.Contains(query, "ORDER BY") {
		// Simple text search - convert to JQL
		jql = fmt.Sprintf("text ~ \"%s\" OR summary ~ \"%s\" OR description ~ \"%s\"", query, query, query)
	}
	
	log.Printf("Original query: %s", query)
	log.Printf("Generated JQL: %s", jql)
	
	// Search for issues using JIRA REST API
	// Try using Atlassian API gateway with cloud ID instead of direct instance URL
	searchURL := fmt.Sprintf("https://api.atlassian.com/ex/jira/%s/rest/api/3/search", resource.ID)
	
	// Prepare search request
	searchData := map[string]interface{}{
		"jql":         jql,
		"maxResults": 20,
		"fields": []string{
			"key",
			"summary",
			"description",
			"issuetype",
			"status",
			"priority",
		},
	}
	
	searchJSON, err := json.Marshal(searchData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search data: %v", err)
	}
	
	searchReq, err := http.NewRequest("POST", searchURL, strings.NewReader(string(searchJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to create search request: %v", err)
	}
	
	searchReq.Header.Set("Authorization", "Bearer "+conn.APIToken)
	searchReq.Header.Set("Accept", "application/json")
	searchReq.Header.Set("Content-Type", "application/json")
	// Add cloud ID context for OAuth
	searchReq.Header.Set("X-Atlassian-Token", "no-check")
	searchReq.Header.Set("X-ExperimentalApi", "opt-in")
	
	// Log the complete request details
	log.Printf("=== JIRA API REQUEST DETAILS ===")
	log.Printf("Method: %s", searchReq.Method)
	log.Printf("URL: %s", searchReq.URL.String())
	log.Printf("Headers:")
	for name, values := range searchReq.Header {
		for _, value := range values {
			if name == "Authorization" {
				log.Printf("  %s: Bearer %s...", name, value[7:27]) // Show first 20 chars of token
			} else {
				log.Printf("  %s: %s", name, value)
			}
		}
	}
	log.Printf("Request Body: %s", string(searchJSON))
	log.Printf("=================================")
	
	searchResp, err := http.DefaultClient.Do(searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to search issues: %v", err)
	}
	defer searchResp.Body.Close()
	
	if searchResp.StatusCode != http.StatusOK {
		// Read the error response body
		errorBody, _ := io.ReadAll(searchResp.Body)
		log.Printf("JIRA search request failed: %d %s, URL: %s", searchResp.StatusCode, searchResp.Status, searchURL)
		log.Printf("JQL query: %s", jql)
		log.Printf("Error response: %s", string(errorBody))
		return nil, fmt.Errorf("search request failed: %s - %s", searchResp.Status, string(errorBody))
	}
	
	// Parse search results
	var searchResults struct {
		Issues []struct {
			Key    string `json:"key"`
			Fields struct {
				Summary     string      `json:"summary"`
				Description interface{} `json:"description"` // Can be string or complex object
				IssueType   struct {
					Name string `json:"name"`
				} `json:"issuetype"`
				Status struct {
					Name string `json:"name"`
				} `json:"status"`
				Priority struct {
					Name string `json:"name"`
				} `json:"priority"`
			} `json:"fields"`
		} `json:"issues"`
		Total int `json:"total"`
	}
	
	if err := json.NewDecoder(searchResp.Body).Decode(&searchResults); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %v", err)
	}
	
	// Convert to Story objects
	var stories []Story
	for _, issue := range searchResults.Issues {
		// Handle description field which can be string or complex object
		var description string
		if issue.Fields.Description != nil {
			switch desc := issue.Fields.Description.(type) {
			case string:
				description = desc
			case map[string]interface{}:
				// Jira rich text format - try to extract plain text
				if descBytes, err := json.Marshal(desc); err == nil {
					description = "Rich text: " + string(descBytes)[:min(200, len(descBytes))] + "..."
				} else {
					description = "[Rich text content]"
				}
			default:
				description = "[Complex content]"
			}
		}
		
		story := Story{
			ID:          uuid.New().String(),
			Title:       issue.Key + ": " + issue.Fields.Summary,
			Description: description,
			JiraKey:     issue.Key,
		}
		stories = append(stories, story)
	}
	
	return stories, nil
}

// importJiraHandler fetches issues from Jira and adds them as stories to the session.
func importJiraHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		SessionID string `json:"sessionId"`
		JiraURL   string `json:"jiraUrl"`  // e.g., "https://your-domain.atlassian.net"
		Username  string `json:"username"` // Jira email
		APIToken  string `json:"apiToken"` // API token
		JQL       string `json:"jql"`      // e.g., "project = PROJ AND status = 'To Do'"
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if payload.JiraURL == "" || payload.Username == "" || payload.APIToken == "" || payload.JQL == "" {
		http.Error(w, "All Jira fields are required", http.StatusBadRequest)
		return
	}

	sessionsMutex.Lock()
	session, ok := sessions[payload.SessionID]
	if !ok {
		sessionsMutex.Unlock()
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	sessionsMutex.Unlock()

	// Connect to Jira
	tp := jira.BasicAuthTransport{
		Username: payload.Username,
		Password: payload.APIToken,
	}
	client, err := jira.NewClient(tp.Client(), payload.JiraURL)
	if err != nil {
		http.Error(w, "Failed to connect to Jira: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch issues via JQL
	issues, _, err := client.Issue.Search(payload.JQL, &jira.SearchOptions{
		MaxResults: 50, // Limit for safety
	})
	if err != nil {
		http.Error(w, "Failed to fetch Jira issues: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Map to Stories
	newStories := []Story{}
	for _, issue := range issues {
		storyID := uuid.New().String()
		story := Story{
			ID:          storyID,
			Title:       issue.Key + ": " + issue.Fields.Summary,
			Description: issue.Fields.Description,
			JiraKey:     issue.Key,
		}
		newStories = append(newStories, story)
	}

	// Add to session
	sessionsMutex.Lock()
	session.Stories = append(session.Stories, newStories...)
	sessionsMutex.Unlock()

	// Broadcast the update
	go func() {
		hub := getOrCreateHub(payload.SessionID)
		message := map[string]interface{}{
			"type":    "stories_imported",
			"stories": newStories,
			"session": session,
		}
		msgBytes, _ := json.Marshal(message)
		hub.broadcast <- msgBytes
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"imported": len(newStories),
		"stories":  newStories,
	})
}

// wsHandler handles WebSocket connections.
// Expects query param: ?sessionId=xxx
func wsHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "sessionId query param required", http.StatusBadRequest)
		return
	}

	// Check if session exists.
	sessionsMutex.Lock()
	session, ok := sessions[sessionID]
	sessionsMutex.Unlock()
	if !ok {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Upgrade to WebSocket.
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Adjust for production security.
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Register the client in the session's hub.
	hub := getOrCreateHub(sessionID)
	hub.register <- conn

	// Goroutine to handle reading messages (now includes chat handling).
	go func() {
		defer func() {
			hub.unregister <- conn
		}()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				break
			}

			var incoming map[string]interface{}
			if err := json.Unmarshal(msg, &incoming); err != nil {
				log.Printf("Invalid message JSON: %v", err)
				continue
			}

			switch incoming["type"] {
			case "chat":
				author, ok := incoming["author"].(string)
				if !ok {
					continue
				}
				text, ok := incoming["text"].(string)
				if !ok {
					continue
				}

				// Append to session's chat history
				sessionsMutex.Lock()
				session.ChatMessages = append(session.ChatMessages, ChatMessage{
					Author:    author,
					Text:      text,
					Timestamp: time.Now().Format(time.RFC3339),
				})
				sessionsMutex.Unlock()

				// Broadcast the new message
				broadcastMsg := map[string]interface{}{
					"type":        "chat_message",
					"chatMessage": session.ChatMessages[len(session.ChatMessages)-1],
				}
				broadcastBytes, _ := json.Marshal(broadcastMsg)
				hub.broadcast <- broadcastBytes
			// TODO: Add other types if needed later (e.g., "vote").
			default:
				log.Printf("Unknown message type: %v", incoming["type"])
			}
		}
	}()
}

// demoOAuthHandler simulates OAuth flow for testing popup mechanism
func demoOAuthHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	// Create mock user data
	mockUserData := map[string]interface{}{
		"account_id": "demo-user-123",
		"name":       "Demo User",
		"email":      "demo@example.com",
		"picture":    "https://avatar.atlassian.com/demo.png",
	}

	// Save mock connection
	jiraConnectionsMutex.Lock()
	jiraConnections[sessionID] = &JiraConnection{
		SessionID: sessionID,
		APIToken:  "demo-token-123",
		Username:  "demo@example.com",
		JiraURL:   "https://demo.atlassian.net",
	}
	jiraConnectionsMutex.Unlock()

	// Create user data JSON
	userDataJSON, _ := json.Marshal(mockUserData)

	// Return HTML page that simulates OAuth success
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Demo OAuth - Planning Poker Tavern</title>
			<style>
				body {
					font-family: Arial, sans-serif;
					background: linear-gradient(135deg, #1a1a2e 0%%, #16213e 100%%);
					color: white;
					padding: 2rem;
					text-align: center;
					min-height: 100vh;
					display: flex;
					flex-direction: column;
					justify-content: center;
					align-items: center;
				}
				.container {
					background: rgba(44, 62, 80, 0.8);
					border: 2px solid #8b4513;
					border-radius: 12px;
					padding: 2rem;
					max-width: 400px;
					margin: 0 auto;
				}
				.success-icon {
					font-size: 3rem;
					margin-bottom: 1rem;
				}
				.button {
					background: linear-gradient(145deg, #27ae60, #229954);
					color: white;
					border: none;
					padding: 1rem 2rem;
					border-radius: 8px;
					cursor: pointer;
					font-size: 1rem;
					margin-top: 1rem;
					transition: all 0.3s ease;
				}
				.button:hover {
					background: linear-gradient(145deg, #229954, #1e8449);
					box-shadow: 0 4px 12px rgba(39, 174, 96, 0.4);
				}
				.countdown {
					color: #f39c12;
					margin-top: 1rem;
					font-size: 0.9rem;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<div class="success-icon">🎉</div>
				<h1>Demo OAuth Success!</h1>
				<p>You have successfully "authenticated" with the demo JIRA integration.</p>
				<p><strong>User:</strong> Demo User (demo@example.com)</p>
				<p>This popup will close automatically and notify the parent window.</p>
				<button class="button" onclick="completeAuth()">✅ Complete Authentication</button>
				<div class="countdown">Auto-closing in <span id="countdown">5</span> seconds...</div>
			</div>

			<script>
				let countdown = 5;
				const countdownElement = document.getElementById('countdown');
				
				function updateCountdown() {
					countdown--;
					if (countdownElement) {
						countdownElement.textContent = countdown;
					}
					if (countdown <= 0) {
						completeAuth();
					}
				}
				
				function completeAuth() {
					try {
						window.opener.postMessage({
							type: 'JIRA_AUTH_SUCCESS',
							user: %s
						}, 'http://localhost:5173');
						window.close();
					} catch (error) {
						console.error('Error sending message:', error);
						window.close();
					}
				}
				
				// Start countdown
				setInterval(updateCountdown, 1000);
			</script>
		</body>
		</html>
	`, userDataJSON)))
}

// corsMiddleware adds CORS headers to every response.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow access from the frontend (both common ports).
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:5173" || origin == "http://localhost:5174" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5174")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
