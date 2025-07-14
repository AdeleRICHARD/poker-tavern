package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Session represents a Planning Poker session.
type Session struct {
	SessionID string   `json:"sessionId"`
	Name      string   `json:"name"`
	Players   []string `json:"players"`
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
	http.HandleFunc("/health", healthHandler)            // Health check endpoint.
	http.HandleFunc("/session", createSessionHandler)    // POST request to create a session.
	http.HandleFunc("/session/join", joinSessionHandler) // POST request to join a session.
	http.HandleFunc("/ws", wsHandler)                    // WebSocket endpoint for real-time updates.

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

// createSessionHandler handles session creation.
// Expects a POST request with JSON body: {"name": "session name"}
func createSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
		SessionID: sessionID,
		Name:      payload.Name,
		Players:   []string{},
	}

	sessionsMutex.Lock()
	sessions[sessionID] = session
	sessionsMutex.Unlock()

	// Initialize WebSocket hub for the new session.
	getOrCreateHub(sessionID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
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
	_, ok := sessions[sessionID]
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

	// Goroutine to handle reading messages (e.g., for future chat/voting).
	go func() {
		defer func() {
			hub.unregister <- conn
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				break
			}
			// TODO: Handle incoming messages (e.g., chat, votes).
		}
	}()
}

// corsMiddleware adds CORS headers to every response.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow access from the frontend.
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
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
