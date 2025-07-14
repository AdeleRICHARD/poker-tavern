package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"sync"

	"github.com/google/uuid"
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
)

func main() {
	// Define endpoints.
	http.HandleFunc("/health", healthHandler)            // Health check endpoint.
	http.HandleFunc("/session", createSessionHandler)    // POST request to create a session.
	http.HandleFunc("/session/join", joinSessionHandler) // POST request to join a session.

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

// joinSessionHandler allows a player to join an existing session.
// Expects a POST request with JSON body: {"sessionId": "xxx", "playerName": "name"}
func joinSessionHandler(w http.ResponseWriter, r *http.Request) {
	// possible go routine to enhance
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
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
