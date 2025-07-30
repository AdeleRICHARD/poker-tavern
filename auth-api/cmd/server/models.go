package main

import (
	"time"

	"gorm.io/gorm"
)

// SessionDB represents a Planning Poker session in the database
type SessionDB struct {
	gorm.Model
	SessionID    string            `gorm:"uniqueIndex;not null" json:"sessionId"`
	Name         string            `gorm:"not null" json:"name"`
	IsActive     bool              `gorm:"default:true" json:"isActive"`
	RevealVotes  bool              `gorm:"default:false" json:"revealVotes"`
	Players      []PlayerDB        `gorm:"foreignKey:SessionID;references:SessionID" json:"players"`
	Stories      []StoryDB         `gorm:"foreignKey:SessionID;references:SessionID" json:"stories"`
	ChatMessages []ChatMessageDB   `gorm:"foreignKey:SessionID;references:SessionID" json:"chatMessages"`
	Votes        []VoteDB          `gorm:"foreignKey:SessionID;references:SessionID" json:"votes"`
}

// PlayerDB represents a player in a session
type PlayerDB struct {
	gorm.Model
	PlayerID   string    `gorm:"uniqueIndex;not null" json:"playerId"`
	SessionID  string    `gorm:"not null;index" json:"sessionId"`
	Name       string    `gorm:"not null" json:"name"`
	Character  string    `json:"character"`
	Emoji      string    `json:"emoji"`
	IsReady    bool      `gorm:"default:false" json:"isReady"`
	Position   string    `json:"position"` // JSON string for {x: number, y: number}
	LastSeen   time.Time `json:"lastSeen"`
}

// StoryDB represents an imported story (e.g., from Jira)
type StoryDB struct {
	gorm.Model
	StoryID     string    `gorm:"uniqueIndex;not null" json:"id"`
	SessionID   string    `gorm:"not null;index" json:"sessionId"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `json:"description"`
	JiraKey     string    `json:"jiraKey,omitempty"`
	Type        string    `json:"type,omitempty"`
	Status      string    `json:"status,omitempty"`
	Priority    string    `json:"priority,omitempty"`
	Points      *int      `json:"estimatedPoints,omitempty"`
	Votes       []VoteDB  `gorm:"foreignKey:StoryID;references:StoryID" json:"votes,omitempty"`
}

// ChatMessageDB represents a chat entry
type ChatMessageDB struct {
	gorm.Model
	MessageID string    `gorm:"uniqueIndex;not null" json:"id"`
	SessionID string    `gorm:"not null;index" json:"sessionId"`
	Author    string    `gorm:"not null" json:"author"`
	Text      string    `gorm:"not null" json:"text"`
	Type      string    `gorm:"default:'message'" json:"type"` // "message" or "system"
}

// VoteDB represents a player's vote on a story
type VoteDB struct {
	gorm.Model
	VoteID    string    `gorm:"uniqueIndex;not null" json:"id"`
	SessionID string    `gorm:"not null;index" json:"sessionId"`
	StoryID   string    `gorm:"not null;index" json:"storyId"`
	PlayerID  string    `gorm:"not null;index" json:"playerId"`
	Value     string    `gorm:"not null" json:"value"`
}

// JiraConnectionDB stores JIRA connection info per session
type JiraConnectionDB struct {
	gorm.Model
	SessionID   string     `gorm:"uniqueIndex;not null" json:"sessionId"`
	JiraURL     string     `json:"jiraUrl"`
	Username    string     `json:"username"`
	APIToken    string     `json:"apiToken"`    // Encrypted in production
	AccessToken string     `json:"accessToken"` // OAuth token
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}

// Helper functions to convert between old structs and new DB structs

func (s *SessionDB) ToSession() *Session {
	session := &Session{
		SessionID:    s.SessionID,
		Name:         s.Name,
		Players:      make([]string, len(s.Players)),
		Stories:      make([]Story, len(s.Stories)),
		ChatMessages: make([]ChatMessage, len(s.ChatMessages)),
	}

	// Convert players
	for i, player := range s.Players {
		session.Players[i] = player.Name
	}

	// Convert stories
	for i, story := range s.Stories {
		session.Stories[i] = Story{
			ID:          story.StoryID,
			Title:       story.Title,
			Description: story.Description,
			JiraKey:     story.JiraKey,
			Type:        story.Type,
			Status:      story.Status,
			Priority:    story.Priority,
		}
		if story.Points != nil {
			session.Stories[i].EstimatedPoints = story.Points
		}
	}

	// Convert chat messages
	for i, msg := range s.ChatMessages {
		session.ChatMessages[i] = ChatMessage{
			Author:    msg.Author,
			Text:      msg.Text,
			Timestamp: msg.CreatedAt.Format(time.RFC3339),
		}
	}

	return session
}

func (s *Session) ToSessionDB() *SessionDB {
	sessionDB := &SessionDB{
		SessionID:    s.SessionID,
		Name:         s.Name,
		IsActive:     true,
		RevealVotes:  false,
		Players:      make([]PlayerDB, 0),
		Stories:      make([]StoryDB, 0),
		ChatMessages: make([]ChatMessageDB, 0),
	}

	return sessionDB
}
