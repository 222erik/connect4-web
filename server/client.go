package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Client represents a connected WebSocket client
type Client struct {
	hub *Hub

	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	Send chan []byte

	// Player's username (empty if anonymous)
	Username string

	// Player's ID
	ID string

	// Current game ID (empty if not in a game)
	GameID string

	// Is this client a bot
	IsBot bool

	// Last ping time for activity tracking
	LastPing time.Time
}

// HandleWebSocket handles WebSocket requests from the peer
func HandleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:      hub,
		conn:     conn,
		Send:     make(chan []byte, 256),
		LastPing: time.Now(),
	}

	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.UnregisterClient(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.LastPing = time.Now()
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from the client
func (c *Client) handleMessage(data []byte) {
	var msg map[string]any
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Invalid message format: %v", err)
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		log.Printf("Message missing type field")
		return
	}

	// Update last ping time on any message
	c.LastPing = time.Now()

	switch msgType {
	case "new_user":
		c.handleNewUser(msg)
	case "ping":
		c.handlePing()
	case "find_match":
		c.startBotGame()
	case "game_move":
		c.handleGameMove(msg)
	case "send_invite":
		c.handleSendInvite(msg)
	case "accept_invite":
		c.handleAcceptInvite(msg)
	case "decline_invite":
		c.handleDeclineInvite(msg)
	case "leave_game":
		c.handleLeaveGame()
	case "emoji_reaction":
		c.handleEmojiReaction(msg)
	default:
		log.Printf("Unknown message type: %s", msgType)
	}
}

func (c *Client) handleNewUser(msg map[string]any) {
	username, ok := msg["username"].(string)
	if !ok {
		c.sendError("Invalid username")
		return
	}

	if user := c.hub.ClientFromUsername(username); user != nil {
		c.sendError("Username already taken")
		return
	}

	c.Username = username
	c.hub.RegisterClient(c)
	c.sendMessage(map[string]any{
		"type":     "user_created",
		"username": username,
	})
	// Broadcast updated player list
	c.hub.BroadcastOnlinePlayers()
}

func (c *Client) handlePing() {
	c.sendMessage(map[string]any{
		"type": "pong",
	})
}

func (c *Client) handleGameMove(msg map[string]any) {
	if c.GameID == "" {
		c.sendError("Not in a game")
		return
	}

	column, ok := msg["column"].(float64)
	if !ok || column < 0 || column > 6 {
		c.sendError("Invalid column")
		return
	}

	game := c.hub.GetGame(c.GameID)
	if game == nil {
		c.sendError("Game not found")
		return
	}

	game.MakeMove(c, int(column))
}

func (c *Client) handleSendInvite(msg map[string]any) {
	if c.Username == "" {
		c.sendError("Must be logged in to send invites")
		return
	}

	to, ok := msg["to"].(string)
	if !ok || to == "" {
		c.sendError("Invalid invite target")
		return
	}

	if !c.hub.SendInvite(c, to) {
		c.sendError("Failed to send invite")
		return
	}

	c.sendMessage(map[string]any{
		"type": "invite_sent",
		"to":   to,
	})
}

func (c *Client) handleAcceptInvite(msg map[string]any) {
	from, ok := msg["from"].(string)
	if !ok || from == "" {
		c.sendError("Invalid invite")
		return
	}

	inviter := c.hub.ClientFromUsername(from)
	if inviter == nil {
		c.sendError("Inviter not found")
		return
	}

	// Create game
	game := c.hub.CreateGame(inviter, c)
	game.Start()

	// Remove invite
	c.hub.CancelInvite(from)
}

func (c *Client) handleDeclineInvite(msg map[string]any) {
	from, ok := msg["from"].(string)
	if !ok {
		return
	}

	inviter := c.hub.ClientFromUsername(from)
	if inviter != nil {
		inviter.sendMessage(map[string]any{
			"type": "invite_declined",
			"from": c.Username,
		})
	}

	c.hub.CancelInvite(from)
}

func (c *Client) handleLeaveGame() {
	if c.GameID == "" {
		return
	}

	game := c.hub.GetGame(c.GameID)
	if game != nil {
		game.HandleLeave(c)
	}
}

func (c *Client) handleEmojiReaction(msg map[string]any) {
	if c.GameID == "" {
		return
	}

	game := c.hub.GetGame(c.GameID)
	if game == nil {
		return
	}

	emoji, ok := msg["emoji"].(string)
	if !ok {
		return
	}

	game.BroadcastToPlayers(map[string]any{
		"type":  "emoji_reaction",
		"from":  c.Username,
		"emoji": emoji,
	})
}

func (c *Client) startBotGame() {
	game := c.hub.CreateGame(c, &Client{
		hub:      c.hub,
		Send:     make(chan []byte, 256),
		Username: "Bot",
		IsBot:    true,
	})
	game.Start()
}

func (c *Client) sendMessage(msg map[string]any) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	select {
	case c.Send <- data:
	default:
		// Buffer full
	}
}

func (c *Client) sendError(errorMsg string) {
	c.sendMessage(map[string]any{
		"type":  "error",
		"error": errorMsg,
	})
}
