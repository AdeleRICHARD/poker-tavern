package main

import (
	"encoding/base64"
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
	SessionID       string                       `json:"sessionId"`
	Name            string                       `json:"name"`
	Players         []string                     `json:"players"`
	PlayerCharacters map[string]string            `json:"playerCharacters,omitempty"`
	PlayerEmojis     map[string]string            `json:"playerEmojis,omitempty"`
	Stories         []Story                      `json:"stories"`
	ChatMessages    []ChatMessage                `json:"chatMessages"`
	PersistentVotes map[string]map[string]string `json:"persistentVotes,omitempty"`
}

// Story represents an imported story (e.g., from Jira).
type Story struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	JiraKey         string `json:"jiraKey,omitempty"`
	EstimatedPoints *int   `json:"estimatedPoints,omitempty"`
	// JIRA-specific fields
	Type     string `json:"type,omitempty"`
	Status   string `json:"status,omitempty"`
	Priority string `json:"priority,omitempty"`
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

type upstashClient struct {
	url   string
	token string
	http  *http.Client
}

var (
	sessionStoreUpstash *upstashClient
)

func initSessionStore() {
	// Prefer explicit Upstash REST variables, but also support Vercel KV (Upstash)
	// integration variable names.
	u := strings.TrimSpace(os.Getenv("UPSTASH_REDIS_REST_URL"))
	t := strings.TrimSpace(os.Getenv("UPSTASH_REDIS_REST_TOKEN"))
	if u == "" || t == "" {
		u = strings.TrimSpace(os.Getenv("KV_REST_API_URL"))
		t = strings.TrimSpace(os.Getenv("KV_REST_API_TOKEN"))
	}
	if u == "" || t == "" {
		return
	}
	sessionStoreUpstash = &upstashClient{
		url:   u,
		token: t,
		http:  &http.Client{Timeout: 5 * time.Second},
	}
	if _, err := sessionStoreUpstash.command("PING"); err != nil {
		log.Printf("⚠️ Session store: Upstash REST disabled (PING failed): %v", err)
		sessionStoreUpstash = nil
		return
	}
	log.Printf("✅ Session store: Upstash REST enabled")
}

func (c *upstashClient) command(args ...any) (any, error) {
	// Upstash REST expects a JSON array: ["PING"] / ["SET", "k", "v"]
	body, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", c.url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Result any    `json:"result"`
		Error  string `json:"error"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("upstash invalid json: %w", err)
	}
	if parsed.Error != "" {
		return nil, fmt.Errorf("upstash error: %s", parsed.Error)
	}
	return parsed.Result, nil
}

func sessionKey(sessionID string) string {
	return "session:" + sessionID
}

func loadSession(sessionID string) (*Session, bool) {
	if sessionStoreUpstash == nil {
		sessionsMutex.Lock()
		s, ok := sessions[sessionID]
		sessionsMutex.Unlock()
		return s, ok
	}

	res, err := sessionStoreUpstash.command("GET", sessionKey(sessionID))
	if err != nil || res == nil {
		if err != nil {
			log.Printf("⚠️ Session store GET failed: %v", err)
		}
		return nil, false
	}

	// Upstash returns string for GET
	raw, ok := res.(string)
	if !ok || raw == "" {
		return nil, false
	}

	var s Session
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		log.Printf("⚠️ Session store decode failed: %v", err)
		return nil, false
	}
	return &s, true
}

func saveSession(s *Session) {
	if s == nil {
		return
	}

	if sessionStoreUpstash == nil {
		sessionsMutex.Lock()
		sessions[s.SessionID] = s
		sessionsMutex.Unlock()
		return
	}

	b, err := json.Marshal(s)
	if err != nil {
		return
	}

	// Keep sessions for 24h by default to avoid unbounded growth.
	if _, err := sessionStoreUpstash.command("SET", sessionKey(s.SessionID), string(b), "EX", 60*60*24); err != nil {
		log.Printf("⚠️ Session store SET failed: %v", err)
	}
}

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
			log.Printf("➕ Client registered. Total clients: %d", len(h.clients))
			h.mutex.Unlock()
		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
				log.Printf("➖ Client unregistered. Total clients: %d", len(h.clients))
			}
			h.mutex.Unlock()
		case message := <-h.broadcast:
			h.mutex.Lock()
			log.Printf("📢 Broadcasting message to %d clients", len(h.clients))
			// Log the message content for debugging, truncate if too long
			logMsg := string(message)
			if len(logMsg) > 200 {
				logMsg = logMsg[:200] + "..."
			}
			log.Printf("Broadcasting message content: %s", logMsg)

			for client := range h.clients {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("❌ WebSocket write error: %v", err)
					// On write error, unregister the client
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

// extractTextFromADF recursively extracts formatted text from Atlassian Document Format (ADF)
func extractTextFromADF(node map[string]interface{}) string {
	var text strings.Builder

	// Get the node type
	nodeType, hasType := node["type"]
	typeStr := ""
	if hasType {
		typeStr, _ = nodeType.(string)
	}

	// Handle different node types with formatting
	switch typeStr {
	case "heading":
		// Extract heading level
		level := 3 // default
		if attrs, exists := node["attrs"]; exists {
			if attrsMap, ok := attrs.(map[string]interface{}); ok {
				if levelVal, exists := attrsMap["level"]; exists {
					if levelInt, ok := levelVal.(float64); ok {
						level = int(levelInt)
					}
				}
			}
		}
		tag := fmt.Sprintf("h%d", level)
		text.WriteString(fmt.Sprintf("<%s>", tag))
		if content, exists := node["content"]; exists {
			if contentArray, ok := content.([]interface{}); ok {
				for _, child := range contentArray {
					if childNode, ok := child.(map[string]interface{}); ok {
						text.WriteString(extractTextFromADF(childNode))
					}
				}
			}
		}
		text.WriteString(fmt.Sprintf("</%s>", tag))
		return text.String()

	case "paragraph":
		text.WriteString("<p>")
		if content, exists := node["content"]; exists {
			if contentArray, ok := content.([]interface{}); ok {
				for _, child := range contentArray {
					if childNode, ok := child.(map[string]interface{}); ok {
						text.WriteString(extractTextFromADF(childNode))
					}
				}
			}
		}
		text.WriteString("</p>")
		return text.String()

	case "text":
		// Handle text with potential marks (bold, italic, etc.)
		textContent := ""
		if textVal, exists := node["text"]; exists {
			if str, ok := textVal.(string); ok {
				textContent = str
			}
		}

		// Check for marks (bold, italic, etc.)
		if marks, exists := node["marks"]; exists {
			if marksArray, ok := marks.([]interface{}); ok {
				for _, mark := range marksArray {
					if markMap, ok := mark.(map[string]interface{}); ok {
						if markType, exists := markMap["type"]; exists {
							if markTypeStr, ok := markType.(string); ok {
								switch markTypeStr {
								case "strong":
									textContent = "<strong>" + textContent + "</strong>"
								case "em":
									textContent = "<em>" + textContent + "</em>"
								case "code":
									textContent = "<code>" + textContent + "</code>"
								case "link":
									if attrs, exists := markMap["attrs"]; exists {
										if attrsMap, ok := attrs.(map[string]interface{}); ok {
											if href, exists := attrsMap["href"]; exists {
												if hrefStr, ok := href.(string); ok {
													textContent = fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, hrefStr, textContent)
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		text.WriteString(textContent)
		return text.String()

	case "emoji":
		// Handle emojis
		if attrs, exists := node["attrs"]; exists {
			if attrsMap, ok := attrs.(map[string]interface{}); ok {
				if emojiText, exists := attrsMap["text"]; exists {
					if emojiStr, ok := emojiText.(string); ok {
						text.WriteString(emojiStr)
					}
				}
			}
		}
		return text.String()

	case "hardBreak":
		return "<br>"

	case "rule":
		return "<hr>"

	case "panel":
		// Handle Jira panels (info, warning, etc.)
		panelType := "info"
		if attrs, exists := node["attrs"]; exists {
			if attrsMap, ok := attrs.(map[string]interface{}); ok {
				if pType, exists := attrsMap["panelType"]; exists {
					if pTypeStr, ok := pType.(string); ok {
						panelType = pTypeStr
					}
				}
			}
		}
		text.WriteString(fmt.Sprintf(`<div class="panel panel-%s">`, panelType))
		if content, exists := node["content"]; exists {
			if contentArray, ok := content.([]interface{}); ok {
				for _, child := range contentArray {
					if childNode, ok := child.(map[string]interface{}); ok {
						text.WriteString(extractTextFromADF(childNode))
					}
				}
			}
		}
		text.WriteString("</div>")
		return text.String()

	default:
		// For unknown types or nodes without type, process content recursively
		if content, exists := node["content"]; exists {
			if contentArray, ok := content.([]interface{}); ok {
				for _, child := range contentArray {
					if childNode, ok := child.(map[string]interface{}); ok {
						text.WriteString(extractTextFromADF(childNode))
					}
				}
			}
		}

		// If it's a text node without type
		if textContent, exists := node["text"]; exists {
			if str, ok := textContent.(string); ok {
				text.WriteString(str)
			}
		}
	}

	return text.String()
}

func main() {
	// Load environment variables from .env file (optional in production)
	err := godotenv.Load()
	if err != nil {
		// Only log if it's an actual error, not just a missing file
		if !strings.Contains(err.Error(), "no such file or directory") && !strings.Contains(err.Error(), "cannot find the file") {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}

	// Note: Using in-memory storage for simplicity
	// For production, consider adding database persistence
	initSessionStore()

	// Define endpoints.
	http.HandleFunc("/health", healthHandler)                  // Health check endpoint.
	http.HandleFunc("/session", handleSession)                 // POST to create, GET to retrieve session.
	http.HandleFunc("/session/join", joinSessionHandler)       // POST request to join a session.
	http.HandleFunc("/session/vote", castVoteHandler)          // POST: cast a vote (HTTP fallback for prod).
	http.HandleFunc("/session/character", setCharacterHandler) // POST: persist character selection (HTTP fallback for prod).
	http.HandleFunc("/session/import-jira", importJiraHandler) // POST request to import Jira issues.
	http.HandleFunc("/ws", wsHandler)                          // WebSocket endpoint for real-time updates (including chat).

	// JIRA integration endpoints
	http.HandleFunc("/api/jira/auth-url", jiraAuthUrlHandler)               // Get JIRA OAuth URL
	http.HandleFunc("/auth/jira/callback", jiraCallbackHandler)             // JIRA OAuth callback
	http.HandleFunc("/jira/status", jiraStatusHandler)                      // JIRA connection status
	http.HandleFunc("/jira/search", jiraSearchHandler)                      // JIRA issue search
	http.HandleFunc("/jira/import-issues", importSelectedJiraIssuesHandler) // Import selected JIRA issues
	http.HandleFunc("/jira/story-points", jiraUpdateStoryPointsHandler)     // Update story points on a JIRA issue

	// Demo OAuth endpoint for testing popup flow
	http.HandleFunc("/demo/oauth", demoOAuthHandler)

	// Wrap DefaultServeMux with CORS middleware.
	handler := corsMiddleware(http.DefaultServeMux)

	// Get port from environment variable (for cloud platforms like Render)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default for local development
	}
	addr := "0.0.0.0:" + port
	log.Printf("Starting server on %s\n", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// setCharacterHandler stores a player's character selection in a session.
// This is used as an HTTP fallback when WebSockets are unavailable (e.g. Vercel).
func setCharacterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		SessionID   string `json:"sessionId"`
		PlayerID    string `json:"playerId"`
		PlayerName  string `json:"playerName"`
		Character   string `json:"character"`
		Emoji       string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if payload.SessionID == "" || payload.PlayerID == "" || payload.Character == "" {
		http.Error(w, "sessionId, playerId and character are required", http.StatusBadRequest)
		return
	}

	session, ok := loadSession(payload.SessionID)
	if !ok {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if session.PlayerCharacters == nil {
		session.PlayerCharacters = make(map[string]string)
	}
	session.PlayerCharacters[payload.PlayerID] = payload.Character

	if payload.Emoji != "" {
		if session.PlayerEmojis == nil {
			session.PlayerEmojis = make(map[string]string)
		}
		session.PlayerEmojis[payload.PlayerID] = payload.Emoji
	}

	saveSession(session)

	// Best-effort broadcast for local/dev environments where WS works.
	go func() {
		hub := getOrCreateHub(payload.SessionID)
		msg := map[string]interface{}{
			"type":       "character_changed",
			"playerId":   payload.PlayerID,
			"playerName": payload.PlayerName,
			"character":  payload.Character,
			"emoji":      payload.Emoji,
		}
		if b, err := json.Marshal(msg); err == nil {
			hub.broadcast <- b
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// castVoteHandler stores a vote for a story in a session.
// This is used as an HTTP fallback when WebSockets are unavailable (e.g. Vercel).
func castVoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		SessionID  string `json:"sessionId"`
		StoryID    string `json:"storyId"`
		PlayerID   string `json:"playerId"`
		Vote       string `json:"vote"`
		PlayerName string `json:"playerName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if payload.SessionID == "" || payload.StoryID == "" || payload.PlayerID == "" || payload.Vote == "" {
		http.Error(w, "sessionId, storyId, playerId and vote are required", http.StatusBadRequest)
		return
	}

	session, ok := loadSession(payload.SessionID)
	if !ok {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if session.PersistentVotes == nil {
		session.PersistentVotes = make(map[string]map[string]string)
	}
	if _, ok := session.PersistentVotes[payload.StoryID]; !ok {
		session.PersistentVotes[payload.StoryID] = make(map[string]string)
	}
	session.PersistentVotes[payload.StoryID][payload.PlayerID] = payload.Vote
	saveSession(session)

	// Best-effort broadcast for local/dev environments where WS works.
	go func() {
		hub := getOrCreateHub(payload.SessionID)
		msg := map[string]interface{}{
			"type":        "vote_cast",
			"story_id":    payload.StoryID,
			"player_id":   payload.PlayerID,
			"vote":        payload.Vote,
			"player_name": payload.PlayerName,
		}
		if b, err := json.Marshal(msg); err == nil {
			hub.broadcast <- b
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
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
			PlayerCharacters: map[string]string{},
			PlayerEmojis:     map[string]string{},
			Stories:      []Story{},
			ChatMessages: []ChatMessage{},
		}
		saveSession(session)

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

		session, ok := loadSession(sessionID)
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
		log.Printf("❌ Join session - Invalid JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("🎮 Join session request - SessionID: %s, PlayerName: %s", payload.SessionID, payload.PlayerName)

	if payload.SessionID == "" || payload.PlayerName == "" {
		log.Printf("❌ Join session - Missing required fields")
		http.Error(w, "Both sessionId and playerName are required", http.StatusBadRequest)
		return
	}

	session, ok := loadSession(payload.SessionID)
	if !ok {
		log.Printf("❌ Join session - Session not found: %s", payload.SessionID)
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if session.PlayerCharacters == nil {
		session.PlayerCharacters = make(map[string]string)
	}
	if session.PlayerEmojis == nil {
		session.PlayerEmojis = make(map[string]string)
	}

	log.Printf("✅ Session found: %s (Name: %s, Players: %d)", payload.SessionID, session.Name, len(session.Players))

	// Add player if not already present.
	playerExists := slices.Contains(session.Players, payload.PlayerName)
	if !playerExists {
		session.Players = append(session.Players, payload.PlayerName)
		// Default class for new players.
		if _, ok := session.PlayerCharacters[payload.PlayerName]; !ok {
			session.PlayerCharacters[payload.PlayerName] = "mage"
		}
		if _, ok := session.PlayerEmojis[payload.PlayerName]; !ok {
			session.PlayerEmojis[payload.PlayerName] = "🧙‍♂️"
		}
	}
	saveSession(session)

	// Broadcast the updated session to WebSocket clients (in a goroutine for non-blocking).
	log.Printf("🎭 Player %s joined session %s. Broadcasting to connected clients...", payload.PlayerName, payload.SessionID)
	go func() {
		hub := getOrCreateHub(payload.SessionID)
		message := map[string]interface{}{
			"type":    "player_joined",
			"session": session,
		}
		msgBytes, _ := json.Marshal(message)
		log.Printf("🎭 Broadcasting player_joined message for session %s", payload.SessionID)
		hub.broadcast <- msgBytes
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// getBaseURL determines the base URL from the request
func getBaseURL(r *http.Request) string {
	// Use the Host header to determine the base URL
	scheme := "http"

	// Check for HTTPS in multiple ways (for reverse proxies like Render)
	if r.TLS != nil {
		scheme = "https"
	} else if r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	} else if r.Header.Get("X-Forwarded-Ssl") == "on" {
		scheme = "https"
	} else if strings.Contains(r.Host, "render.com") || strings.Contains(r.Host, "onrender.com") {
		// Render always uses HTTPS for *.onrender.com domains
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

// jiraAuthUrlHandler returns the OAuth URL for JIRA authentication (used by popup flow)
func jiraAuthUrlHandler(w http.ResponseWriter, r *http.Request) {
	// Expect a session ID as a query parameter
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	// Get base URL from request
	baseURL := getBaseURL(r)

	// OAuth 2.0 configuration - get from environment variables
	clientID := os.Getenv("JIRA_CLIENT_ID")
	clientSecret := os.Getenv("JIRA_CLIENT_SECRET")
	demoMode := os.Getenv("DEMO_MODE")
	if demoMode == "" {
		demoMode = "true" // Default to demo mode for easy testing
	}
	demoModeEnabled := demoMode == "true"

	// Check if demo mode is enabled
	if demoModeEnabled {
		// Return demo OAuth URL for testing popup flow
		demoURL := fmt.Sprintf("%s/demo/oauth?sessionId=%s", baseURL, sessionID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"authUrl": demoURL,
			"demo":    "true",
		})
		return
	}

	if clientID == "" || clientSecret == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":              "JIRA OAuth not configured",
			"message":            "Please set JIRA_CLIENT_ID and JIRA_CLIENT_SECRET environment variables",
			"setup_instructions": "Run ./setup_oauth.sh to configure OAuth credentials or set DEMO_MODE=true",
		})
		return
	}
	redirectURI := fmt.Sprintf("%s/auth/jira/callback", baseURL)
	// Include write scopes so we can update Story Points on issues.
	// Atlassian OAuth scopes vary between "classic" (write:jira-work) and "granular" (write:issue:jira).
	// Including both keeps things working across app configurations.
	scopes := "offline_access read:jira-work read:jira-user read:account read:me read:issue:jira read:project:jira write:jira-work write:issue:jira"

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

type jiraAccessibleResource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

func getFirstAccessibleJiraResource(accessToken string) (*jiraAccessibleResource, error) {
	req, err := http.NewRequest("GET", "https://api.atlassian.com/oauth/token/accessible-resources", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get accessible resources: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("accessible resources error (%d): %s", resp.StatusCode, string(body))
	}

	var resources []jiraAccessibleResource
	if err := json.Unmarshal(body, &resources); err != nil {
		return nil, fmt.Errorf("decode accessible resources: %w", err)
	}
	if len(resources) == 0 {
		return nil, fmt.Errorf("no accessible JIRA resources found")
	}
	return &resources[0], nil
}

var (
	jiraStoryPointsFieldCache      = make(map[string]string) // key: cloudID or domain -> customfield id
	jiraStoryPointsFieldCacheMutex = &sync.Mutex{}
)

func resolveStoryPointsFieldIDViaOAuth(accessToken, cloudID string) (string, error) {
	// Allow overriding via env var for deterministic behavior.
	if v := strings.TrimSpace(os.Getenv("JIRA_STORY_POINTS_FIELD_ID")); v != "" {
		return v, nil
	}

	jiraStoryPointsFieldCacheMutex.Lock()
	if cached, ok := jiraStoryPointsFieldCache["oauth:"+cloudID]; ok && cached != "" {
		jiraStoryPointsFieldCacheMutex.Unlock()
		return cached, nil
	}
	jiraStoryPointsFieldCacheMutex.Unlock()

	fieldsURL := fmt.Sprintf("https://api.atlassian.com/ex/jira/%s/rest/api/3/field", cloudID)
	req, err := http.NewRequest("GET", fieldsURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fields request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fields request error (%d): %s", resp.StatusCode, string(body))
	}

	var fields []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &fields); err != nil {
		return "", fmt.Errorf("decode fields: %w", err)
	}

	// Prefer the common Jira Software field label.
	var candidates = []string{"Story point estimate", "Story Points", "Story points"}
	for _, want := range candidates {
		for _, f := range fields {
			if f.ID != "" && strings.EqualFold(strings.TrimSpace(f.Name), want) {
				jiraStoryPointsFieldCacheMutex.Lock()
				jiraStoryPointsFieldCache["oauth:"+cloudID] = f.ID
				jiraStoryPointsFieldCacheMutex.Unlock()
				return f.ID, nil
			}
		}
	}

	// Fallback: pick any customfield containing "story point".
	for _, f := range fields {
		if f.ID != "" && strings.Contains(strings.ToLower(f.Name), "story point") {
			jiraStoryPointsFieldCacheMutex.Lock()
			jiraStoryPointsFieldCache["oauth:"+cloudID] = f.ID
			jiraStoryPointsFieldCacheMutex.Unlock()
			return f.ID, nil
		}
	}

	return "", fmt.Errorf("could not find Story Points field; set JIRA_STORY_POINTS_FIELD_ID env var")
}

func resolveStoryPointsFieldIDViaAPIToken(domain, email, apiToken string) (string, error) {
	if v := strings.TrimSpace(os.Getenv("JIRA_STORY_POINTS_FIELD_ID")); v != "" {
		return v, nil
	}

	cacheKey := "token:" + strings.ToLower(strings.TrimSpace(domain))
	jiraStoryPointsFieldCacheMutex.Lock()
	if cached, ok := jiraStoryPointsFieldCache[cacheKey]; ok && cached != "" {
		jiraStoryPointsFieldCacheMutex.Unlock()
		return cached, nil
	}
	jiraStoryPointsFieldCacheMutex.Unlock()

	fieldsURL := fmt.Sprintf("https://%s/rest/api/3/field", domain)
	req, err := http.NewRequest("GET", fieldsURL, nil)
	if err != nil {
		return "", err
	}
	auth := email + ":" + apiToken
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", "Basic "+encodedAuth)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fields request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fields request error (%d): %s", resp.StatusCode, string(body))
	}

	var fields []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &fields); err != nil {
		return "", fmt.Errorf("decode fields: %w", err)
	}

	var candidates = []string{"Story point estimate", "Story Points", "Story points"}
	for _, want := range candidates {
		for _, f := range fields {
			if f.ID != "" && strings.EqualFold(strings.TrimSpace(f.Name), want) {
				jiraStoryPointsFieldCacheMutex.Lock()
				jiraStoryPointsFieldCache[cacheKey] = f.ID
				jiraStoryPointsFieldCacheMutex.Unlock()
				return f.ID, nil
			}
		}
	}

	for _, f := range fields {
		if f.ID != "" && strings.Contains(strings.ToLower(f.Name), "story point") {
			jiraStoryPointsFieldCacheMutex.Lock()
			jiraStoryPointsFieldCache[cacheKey] = f.ID
			jiraStoryPointsFieldCacheMutex.Unlock()
			return f.ID, nil
		}
	}

	return "", fmt.Errorf("could not find Story Points field; set JIRA_STORY_POINTS_FIELD_ID env var")
}

// jiraUpdateStoryPointsHandler updates Story Points on a JIRA issue.
// Works for both:
// - API token mode (env: JIRA_EMAIL, JIRA_DOMAIN, API_TOKEN_JIRA)
// - OAuth mode (session-scoped access token stored in jiraConnections)
func jiraUpdateStoryPointsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		SessionID   string `json:"sessionId,omitempty"` // required for OAuth mode
		IssueKey    string `json:"issueKey"`
		StoryPoints int    `json:"storyPoints"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	payload.IssueKey = strings.TrimSpace(payload.IssueKey)
	if payload.IssueKey == "" {
		http.Error(w, "issueKey is required", http.StatusBadRequest)
		return
	}
	if payload.StoryPoints < 0 {
		http.Error(w, "storyPoints must be >= 0", http.StatusBadRequest)
		return
	}

	// Safety rail: writing to JIRA is disabled by default to prevent accidental
	// production updates. Explicitly set JIRA_WRITE_ENABLED=true to enable.
	if strings.ToLower(strings.TrimSpace(os.Getenv("JIRA_WRITE_ENABLED"))) != "true" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success":     false,
			"dryRun":      true,
			"message":     "JIRA write operations are disabled. Set JIRA_WRITE_ENABLED=true to enable.",
			"issueKey":    payload.IssueKey,
			"storyPoints": payload.StoryPoints,
		})
		return
	}

	// API token mode (preferred for server-to-server setups)
	if isAPITokenMode() {
		email := strings.TrimSpace(os.Getenv("JIRA_EMAIL"))
		domain := strings.TrimSpace(os.Getenv("JIRA_DOMAIN"))
		token := strings.TrimSpace(os.Getenv("API_TOKEN_JIRA"))

		fieldID, err := resolveStoryPointsFieldIDViaAPIToken(domain, email, token)
		if err != nil {
			http.Error(w, "Failed to resolve Story Points field: "+err.Error(), http.StatusInternalServerError)
			return
		}

		updateURL := fmt.Sprintf("https://%s/rest/api/3/issue/%s", domain, url.PathEscape(payload.IssueKey))
		updateBody := map[string]any{
			"fields": map[string]any{
				fieldID: payload.StoryPoints,
			},
		}
		b, _ := json.Marshal(updateBody)
		req, err := http.NewRequest("PUT", updateURL, strings.NewReader(string(b)))
		if err != nil {
			http.Error(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
			return
		}
		auth := email + ":" + token
		encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
		req.Header.Set("Authorization", "Basic "+encodedAuth)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Failed to update issue: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			http.Error(w, fmt.Sprintf("JIRA API error (%d): %s", resp.StatusCode, string(body)), resp.StatusCode)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success":     true,
			"mode":        "api_token",
			"issueKey":    payload.IssueKey,
			"storyPoints": payload.StoryPoints,
			"fieldId":     fieldID,
		})
		return
	}

	// OAuth mode
	if payload.SessionID == "" {
		http.Error(w, "sessionId is required (OAuth mode)", http.StatusBadRequest)
		return
	}

	jiraConnectionsMutex.Lock()
	conn, ok := jiraConnections[payload.SessionID]
	jiraConnectionsMutex.Unlock()
	if !ok || conn == nil || strings.TrimSpace(conn.APIToken) == "" {
		http.Error(w, "JIRA connection not established", http.StatusUnauthorized)
		return
	}

	resource, err := getFirstAccessibleJiraResource(conn.APIToken)
	if err != nil {
		http.Error(w, "Failed to resolve Jira site: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fieldID, err := resolveStoryPointsFieldIDViaOAuth(conn.APIToken, resource.ID)
	if err != nil {
		http.Error(w, "Failed to resolve Story Points field: "+err.Error(), http.StatusInternalServerError)
		return
	}

	updateURL := fmt.Sprintf("https://api.atlassian.com/ex/jira/%s/rest/api/3/issue/%s", resource.ID, url.PathEscape(payload.IssueKey))
	updateBody := map[string]any{
		"fields": map[string]any{
			fieldID: payload.StoryPoints,
		},
	}
	b, _ := json.Marshal(updateBody)
	req, err := http.NewRequest("PUT", updateURL, strings.NewReader(string(b)))
	if err != nil {
		http.Error(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+conn.APIToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Atlassian-Token", "no-check")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to update issue: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("JIRA API error (%d): %s", resp.StatusCode, string(body)), resp.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success":     true,
		"mode":        "oauth",
		"issueKey":    payload.IssueKey,
		"storyPoints": payload.StoryPoints,
		"fieldId":     fieldID,
		"cloudId":     resource.ID,
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

	// Get base URL from request to use correct host
	baseURL := getBaseURL(r)

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", fmt.Sprintf("%s/auth/jira/callback", baseURL))

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

	// Determine the frontend origin - try multiple approaches
	frontendOrigin := "*" // Use wildcard origin for development

	// First, try to get origin from the referer header
	referer := r.Header.Get("Referer")
	if referer != "" {
		if refererURL, err := url.Parse(referer); err == nil {
			frontendOrigin = fmt.Sprintf("%s://%s", refererURL.Scheme, refererURL.Host)
		}
	}

	// If no referer, try to get from Origin header
	if frontendOrigin == "*" {
		origin := r.Header.Get("Origin")
		if origin != "" {
			frontendOrigin = origin
		}
	}

	// Log for debugging
	log.Printf("OAuth callback - Referer: %s, Origin: %s, Using origin: %s", referer, r.Header.Get("Origin"), frontendOrigin)

	// Send success message to parent window and close popup
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(`
		<html>
		<body>
		<script>
			window.opener.postMessage({
				type: 'JIRA_AUTH_SUCCESS',
				user: %s
			}, '%s');
			window.close();
		</script>
		<p>JIRA connection successful! You can close this window.</p>
		</body>
		</html>
	`, userDataJSON, frontendOrigin)))
}

// jiraStatusHandler checks if a session is connected to JIRA
func jiraStatusHandler(w http.ResponseWriter, r *http.Request) {
	// If API token mode is enabled, always return connected
	if isAPITokenMode() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"connected": true,
			"mode":      "api_token",
			"message":   "Connected via API Token",
		})
		return
	}

	// Legacy OAuth mode
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
		"mode":      "oauth",
	})
}

// isAPITokenMode checks if API token authentication is configured
func isAPITokenMode() bool {
	email := os.Getenv("JIRA_EMAIL")
	domain := os.Getenv("JIRA_DOMAIN")
	token := os.Getenv("API_TOKEN_JIRA")
	return email != "" && domain != "" && token != ""
}

// jiraSearchHandler enables searching for issues in JIRA
func jiraSearchHandler(w http.ResponseWriter, r *http.Request) {
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
			{ID: "DEMO-1", Title: "[DEMO] Fix login issue", Description: "Users cannot log in to the application", JiraKey: "DEMO-1", Type: "Bug", Status: "To Do", Priority: "High"},
			{ID: "DEMO-2", Title: "[DEMO] Add user profile page", Description: "Create a user profile page with basic info", JiraKey: "DEMO-2", Type: "Story", Status: "In Progress", Priority: "Medium"},
			{ID: "DEMO-3", Title: "[DEMO] Performance optimization", Description: "Improve application loading time", JiraKey: "DEMO-3", Type: "Task", Status: "Done", Priority: "Low"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"issues": mockedIssues,
		})
		return
	}

	// Check if API token mode is enabled (no OAuth needed)
	if isAPITokenMode() {
		issues, err := searchJiraIssuesWithAPIToken(query)
		if err != nil {
			log.Printf("JIRA API Token search error: %v", err)
			http.Error(w, "Failed to search JIRA issues: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"issues": issues,
		})
		return
	}

	// Legacy OAuth mode - requires session connection
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required (OAuth mode)", http.StatusBadRequest)
		return
	}

	jiraConnectionsMutex.Lock()
	conn, connExists := jiraConnections[sessionID]
	jiraConnectionsMutex.Unlock()
	if !connExists {
		http.Error(w, "JIRA connection not established", http.StatusUnauthorized)
		return
	}

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

// searchJiraIssuesWithAPIToken searches JIRA using Basic Auth with API token from env vars
func searchJiraIssuesWithAPIToken(query string) ([]Story, error) {
	email := os.Getenv("JIRA_EMAIL")
	domain := os.Getenv("JIRA_DOMAIN")
	token := os.Getenv("API_TOKEN_JIRA")

	log.Printf("Using API Token mode - Email: %s, Domain: %s", email, domain)

	// Build JQL query
	jql := query
	if !strings.Contains(query, "=") && !strings.Contains(query, "~") && !strings.Contains(query, "ORDER BY") {
		// Simple text search - convert to JQL
		jql = fmt.Sprintf("text ~ \"%s\" OR summary ~ \"%s\" ORDER BY updated DESC", query, query)
	}

	log.Printf("JQL Query: %s", jql)

	// Use direct HTTP request to the new JIRA API (v3)
	jiraURL := fmt.Sprintf("https://%s/rest/api/3/search/jql", domain)

	// Prepare search request body
	searchData := map[string]interface{}{
		"jql":        jql,
		"maxResults": 20,
		"fields":     []string{"key", "summary", "description", "issuetype", "status", "priority"},
	}

	searchJSON, err := json.Marshal(searchData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search data: %v", err)
	}

	req, err := http.NewRequest("POST", jiraURL, strings.NewReader(string(searchJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set Basic Auth header
	auth := email + ":" + token
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", "Basic "+encodedAuth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("JIRA API error (%d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var searchResult struct {
		Issues []struct {
			Key    string `json:"key"`
			Fields struct {
				Summary     string      `json:"summary"`
				Description interface{} `json:"description"`
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
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("Found %d issues", len(searchResult.Issues))

	// Convert to Story objects
	var stories []Story
	for _, issue := range searchResult.Issues {
		description := ""
		if issue.Fields.Description != nil {
			switch desc := issue.Fields.Description.(type) {
			case string:
				description = desc
			case map[string]interface{}:
				description = extractTextFromADF(desc)
			}
		}

		story := Story{
			ID:          uuid.New().String(),
			Title:       issue.Key + ": " + issue.Fields.Summary,
			Description: description,
			JiraKey:     issue.Key,
			Type:        issue.Fields.IssueType.Name,
			Status:      issue.Fields.Status.Name,
			Priority:    issue.Fields.Priority.Name,
		}
		stories = append(stories, story)
	}

	return stories, nil
}

// searchJiraIssues searches for JIRA issues using the Atlassian API with OAuth token
func searchJiraIssues(conn *JiraConnection, query string) ([]Story, error) {
	// Log the token being used for debugging
	log.Printf("Using access token for API call: %s", conn.APIToken[:min(20, len(conn.APIToken))]+"...")

	// Step 1: Get accessible resources (sites) for the user
	log.Printf("Getting accessible resources for user")

	req, err := http.NewRequest("GET", "https://api.atlassian.com/oauth/token/accessible-resources", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+conn.APIToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get accessible resources: %v", err)
	}
	defer resp.Body.Close()

	// Read response body for debugging
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	log.Printf("Resources API response status: %d", resp.StatusCode)
	log.Printf("Resources API response body: %s", string(respBody))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get accessible resources: %s - %s", resp.Status, string(respBody))
	}

	var resources []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
	}

	if err := json.Unmarshal(respBody, &resources); err != nil {
		return nil, fmt.Errorf("failed to decode resources: %v", err)
	}

	// Log the accessible resources for debugging
	log.Printf("Found %d accessible resources:", len(resources))
	for i, res := range resources {
		log.Printf("  [%d] ID: %s, Name: %s, URL: %s", i, res.ID, res.Name, res.URL)
	}

	if len(resources) == 0 {
		return nil, fmt.Errorf("no accessible JIRA resources found - user may not have access to any JIRA sites")
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
		"jql":        jql,
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
				log.Printf("  %s: Bearer %s...", name, value[7:min(27, len(value))]) // Show first 20 chars of token
			} else {
				log.Printf("  %s: %s", name, value)
			}
		}
	}
	log.Printf("Request Body: %s", string(searchJSON))
	log.Printf("=================================")

	// Use client with timeout for search request
	searchClient := &http.Client{Timeout: 30 * time.Second}
	searchResp, err := searchClient.Do(searchReq)
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
				// Jira rich text format - extract plain text from ADF
				description = extractTextFromADF(desc)
				log.Printf("Final extracted description: %s", description[:min(300, len(description))])
			default:
				description = "[Complex content]"
			}
		} else {
			log.Printf("Description field is nil for issue: %s", issue.Key)
		}

		story := Story{
			ID:          uuid.New().String(),
			Title:       issue.Key + ": " + issue.Fields.Summary,
			Description: description,
			JiraKey:     issue.Key,
			Type:        issue.Fields.IssueType.Name,
			Status:      issue.Fields.Status.Name,
			Priority:    issue.Fields.Priority.Name,
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

	session, ok := loadSession(payload.SessionID)
	if !ok {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

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
	session.Stories = append(session.Stories, newStories...)
	saveSession(session)

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
	log.Printf("🔌 WebSocket connection attempt for session: %s from %s", sessionID, r.RemoteAddr)

	if sessionID == "" {
		log.Printf("❌ WebSocket rejected: missing sessionId")
		http.Error(w, "sessionId query param required", http.StatusBadRequest)
		return
	}

	// Check if session exists.
	session, ok := loadSession(sessionID)
	if !ok {
		log.Printf("❌ WebSocket rejected: session %s not found", sessionID)
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	log.Printf("✅ Session %s found, has %d players", sessionID, len(session.Players))

	// Upgrade to WebSocket.
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			log.Printf("🔍 WebSocket origin check: %s", r.Header.Get("Origin"))
			return true // Adjust for production security.
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade error: %v", err)
		return
	}

	log.Printf("🚀 WebSocket connection established for session %s", sessionID)

	// Register the client in the session's hub.
	hub := getOrCreateHub(sessionID)
	log.Printf("📡 Registering client with hub for session %s (current clients: %d)", sessionID, len(hub.clients))
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
				session.ChatMessages = append(session.ChatMessages, ChatMessage{
					Author:    author,
					Text:      text,
					Timestamp: time.Now().Format(time.RFC3339),
				})
				saveSession(session)

				// Broadcast the new message
				broadcastMsg := map[string]interface{}{
					"type":        "chat_message",
					"chatMessage": session.ChatMessages[len(session.ChatMessages)-1],
				}
				broadcastBytes, _ := json.Marshal(broadcastMsg)
				hub.broadcast <- broadcastBytes
			case "vote_cast":
				log.Printf("🗳️ Received vote_cast message: %+v", incoming)
				storyID, storyOK := incoming["story_id"].(string)
				playerID, playerOK := incoming["player_id"].(string)
				vote, voteOK := incoming["vote"].(string)
				playerName, playerNameOK := incoming["player_name"].(string)
				log.Printf("🔍 Vote data extracted - storyID:%s(%v), playerID:%s(%v), vote:%s(%v), playerName:%s(%v)",
					storyID, storyOK, playerID, playerOK, vote, voteOK, playerName, playerNameOK)

				if !storyOK || !playerOK || !voteOK || !playerNameOK {
					log.Printf("❌ Invalid vote_cast message: one or more fields are missing or have the wrong type")
					continue
				}

				// Store the vote in the session
				if session.PersistentVotes == nil {
					session.PersistentVotes = make(map[string]map[string]string)
				}
				if _, ok := session.PersistentVotes[storyID]; !ok {
					session.PersistentVotes[storyID] = make(map[string]string)
				}
				session.PersistentVotes[storyID][playerID] = vote
				saveSession(session)

				// Broadcast the vote to all clients
				broadcastMsg := map[string]interface{}{
					"type":        "vote_cast",
					"story_id":    storyID,
					"player_id":   playerID,
					"vote":        vote,
					"player_name": playerName,
				}
				broadcastBytes, _ := json.Marshal(broadcastMsg)
				log.Printf("🗚️ Broadcasting vote_cast message: %s", string(broadcastBytes))
				hub.broadcast <- broadcastBytes
				log.Printf("🗳️ Broadcasted vote_cast for player %s on story %s to %d clients", playerName, storyID, len(hub.clients))
			case "character_changed":
				log.Printf("🎭 Received character_changed message: %+v", incoming)
				playerID, playerOK := incoming["playerId"].(string)
				playerName, playerNameOK := incoming["playerName"].(string)
				character, characterOK := incoming["character"].(string)
				emoji, emojiOK := incoming["emoji"].(string)

				if !playerOK || !characterOK {
					log.Printf("❌ Invalid character_changed message: missing fields")
					continue
				}

				// Persist character on the session so polling/HTTP GET returns it.
				if session.PlayerCharacters == nil {
					session.PlayerCharacters = make(map[string]string)
				}
				session.PlayerCharacters[playerID] = character
				if emojiOK && emoji != "" {
					if session.PlayerEmojis == nil {
						session.PlayerEmojis = make(map[string]string)
					}
					session.PlayerEmojis[playerID] = emoji
				}
				saveSession(session)

				// Broadcast the character change to all clients
				broadcastMsg := map[string]interface{}{
					"type":       "character_changed",
					"playerId":   playerID,
					"playerName": playerName,
					"character":  character,
					"emoji":      emoji,
				}
				broadcastBytes, _ := json.Marshal(broadcastMsg)
				log.Printf("🎭 Broadcasting character_changed: %s", string(broadcastBytes))
				hub.broadcast <- broadcastBytes
				log.Printf("🎭 Broadcasted character change for player %s to %s", playerName, character)
				_ = playerNameOK // Avoid unused variable warning
				_ = emojiOK
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

	// Determine the frontend origin based on the request
	referer := r.Header.Get("Referer")
	frontendOrigin := "http://localhost:5173" // fallback
	if referer != "" {
		if refererURL, err := url.Parse(referer); err == nil {
			frontendOrigin = fmt.Sprintf("%s://%s", refererURL.Scheme, refererURL.Host)
		}
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
						}, '%s');
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
	`, userDataJSON, frontendOrigin)))
}

// importSelectedJiraIssuesHandler allows importing specific JIRA issues as stories
func importSelectedJiraIssuesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		SessionID string  `json:"sessionId"`
		Stories   []Story `json:"stories"` // Array of Story objects to import
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if payload.SessionID == "" || len(payload.Stories) == 0 {
		http.Error(w, "SessionID and stories are required", http.StatusBadRequest)
		return
	}

	// Get the session
	session, ok := loadSession(payload.SessionID)
	if !ok {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Add stories to session
	session.Stories = append(session.Stories, payload.Stories...)
	saveSession(session)

	// Broadcast the update to all clients in the session
	go func() {
		hub := getOrCreateHub(payload.SessionID)
		message := map[string]interface{}{
			"type":    "stories_imported",
			"stories": payload.Stories,
			"session": session,
		}
		msgBytes, _ := json.Marshal(message)
		hub.broadcast <- msgBytes
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"imported": len(payload.Stories),
		"stories":  payload.Stories,
		"success":  true,
	})
}

// corsMiddleware adds CORS headers to every response.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Comma-separated list of allowed origins. Example:
		// ALLOWED_ORIGINS=http://localhost:5173,https://my-frontend.example.com
		allowedOriginsEnv := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS"))
		var allowedOrigins []string
		if allowedOriginsEnv != "" {
			for _, part := range strings.Split(allowedOriginsEnv, ",") {
				if v := strings.TrimSpace(part); v != "" {
					allowedOrigins = append(allowedOrigins, v)
				}
			}
		}

		originAllowed := false
		if origin != "" {
			for _, ao := range allowedOrigins {
				if origin == ao {
					originAllowed = true
					break
				}
			}
		}

		environment := os.Getenv("ENVIRONMENT")

		// Allow any https://*.vercel.app origin (requested behavior).
		// Set ALLOW_VERCEL_APP_ORIGINS=false to disable.
		if !originAllowed && origin != "" && os.Getenv("ALLOW_VERCEL_APP_ORIGINS") != "false" {
			if u, err := url.Parse(origin); err == nil {
				host := strings.ToLower(u.Hostname())
				if u.Scheme == "https" && strings.HasSuffix(host, ".vercel.app") {
					originAllowed = true
				}
			}
		}

		// In local dev it's convenient to allow requests without extra setup.
		if !originAllowed && environment != "production" && origin == "http://localhost:5173" {
			originAllowed = true
		}

		if originAllowed && origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
