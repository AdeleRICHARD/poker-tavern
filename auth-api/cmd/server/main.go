package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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
}

// ChatMessage represents a chat entry.
type ChatMessage struct {
	Author    string `json:"author"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}

// sessions stores the created sessions in memory.
var (
	sessions      = make(map[string]*Session)
	sessionsMutex = &sync.Mutex{}

	// wsHubs manages WebSocket hubs per session for real-time broadcasting.
	wsHubs      = make(map[string]*Hub)
	wsHubsMutex = &sync.Mutex{}
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

func main() {
	// Define endpoints.
	http.HandleFunc("/health", healthHandler)                  // Health check endpoint.
	http.HandleFunc("/session", handleSession)                 // POST to create, GET to retrieve session.
	http.HandleFunc("/session/join", joinSessionHandler)       // POST request to join a session.
	http.HandleFunc("/session/import-jira", importJiraHandler) // POST request to import Jira issues.
	http.HandleFunc("/ws", wsHandler)                          // WebSocket endpoint for real-time updates (including chat).

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
