package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Hub manages WebSocket connections
type Hub struct {
	// Registered clients
	clients map[*Client]bool // This bool doesn't do anything

	// Active games
	games map[string]*Game

	// Pending invites: inviter -> invited
	invites map[string]string

	// Mutex for thread safety
	mu sync.RWMutex

	// Channel for registering clients
	register chan *Client

	// Channel for unregistering clients
	unregister chan *Client
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		games:      make(map[string]*Game),
		invites:    make(map[string]string),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			delete(h.clients, client)
			// Clean up any invites from this client
			delete(h.invites, client.Username)
			// Check if client was in a game
			if client.GameID != "" {
				if game, ok := h.games[client.GameID]; ok {
					// Handle game cleanup
					game.HandleDisconnect(client)
					if game.IsFinished {
						delete(h.games, client.GameID)
					}
				}
			}
			close(client.Send)
			h.mu.Unlock()
		}
	}
}

// BroadcastOnlinePlayers sends the current player list to all clients
func (h *Hub) BroadcastOnlinePlayers() {
	h.mu.RLock()
	players := make([]string, 0, len(h.clients))
	for c := range h.clients {
		if c.Username == "" {
			continue
		}
		players = append(players, c.Username)
	}
	h.mu.RUnlock()

	// Create message
	msg := map[string]any{
		"type":    "players_list",
		"players": players,
	}

	data, _ := json.Marshal(msg)

	h.mu.RLock()
	for client := range h.clients {
		select {
		case client.Send <- data:
		default:
			// Buffer full, skip
		}
	}
	h.mu.RUnlock()
}

// ClientFromUsername returns the client with the given username
func (h *Hub) ClientFromUsername(username string) *Client {
	if username == "" {
		return nil
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		if c.Username == username {
			return c
		}
	}
	return nil
}

// RegisterClient registers a new client
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// Usernames returns all online usernames as JSON
func (h *Hub) Usernames(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	usernames := make([]string, 0, len(h.clients))
	for c := range h.clients {
		if c.Username != "" {
			usernames = append(usernames, c.Username)
		}
	}
	h.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usernames)
}

// UnregisterClient unregisters a client
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// SendInvite sends a game invite from one player to another
func (h *Hub) SendInvite(inviter *Client, inviteeUsername string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Get invitee client (inline to avoid deadlock with ClientFromUsername's RLock)
	var invitee *Client
	for client := range h.clients {
		if client.Username == inviteeUsername {
			invitee = client
			break
		}
	}
	if invitee == nil {
		return false
	}

	// Check if invitee is in a game
	if invitee.GameID != "" {
		return false
	}

	// Store the invite
	h.invites[inviter.Username] = inviteeUsername

	// Send invite to invitee
	inviteMsg := map[string]any{
		"type": "invite_received",
		"from": inviter.Username,
	}
	data, _ := json.Marshal(inviteMsg)

	select {
	case invitee.Send <- data:
		return true
	default:
		return false
	}
}

// CancelInvite cancels a pending invite
func (h *Hub) CancelInvite(inviterUsername string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.invites, inviterUsername)
}

// CreateGame creates a new game between two players
func (h *Hub) CreateGame(player1, player2 *Client) *Game {
	h.mu.Lock()
	defer h.mu.Unlock()

	gameID := h.generateGameID()
	game := NewGame(gameID, player1, player2)
	h.games[gameID] = game

	player1.GameID = gameID
	player2.GameID = gameID

	return game
}

// GetGame returns a game by ID
func (h *Hub) GetGame(gameID string) *Game {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.games[gameID]
}

// RemoveGame removes a game from the hub
func (h *Hub) RemoveGame(gameID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.games, gameID)
}

func (h *Hub) generateGameID() string {
	// Simple game ID generator - could use UUID in production
	return fmt.Sprintf("game_%d", len(h.games)+1)
}
