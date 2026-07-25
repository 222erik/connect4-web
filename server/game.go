package server

import (
	"encoding/json"
	"log"
	"time"
)

const (
	Rows = 6
	Cols = 7
)

// Player represents a player in a game
type Player struct {
	Client *Client
	Piece  int // 1 for red, 2 for yellow
}

// Game represents a Connect Four game
type Game struct {
	ID         string
	Board      [Rows][Cols]int
	Players    [2]*Player
	Turn       int // Index into Players array (0 or 1)
	Winner     *Player
	IsFinished bool
	MoveCount  int

	*Bot
}

// NewGame creates a new game between two clients
func NewGame(id string, player1, player2 *Client) *Game {
	return &Game{
		ID: id,
		Players: [2]*Player{
			{Client: player1, Piece: 1}, // Red goes first
			{Client: player2, Piece: 2}, // Yellow
		},
		Turn: 0,
	}
}

// Start begins the game
func (g *Game) Start() {
	// Notify both players
	for i, player := range g.Players {
		player.Client.sendMessage(map[string]any{
			"type":       "game_start",
			"game_id":    g.ID,
			"your_piece": player.Piece,
			"opponent":   g.Players[1-i].Client.Username,
			"your_turn":  i == 0, // Red goes first
		})
	}

	// If second player is a bot, start bot's turn logic
	if g.Players[1].Client.IsBot {
		g.Bot = NewBot()
	}
}

// MakeMove handles a player making a move
func (g *Game) MakeMove(client *Client, column int) {
	// Check if it's the player's turn
	if g.IsFinished {
		return
	}

	currentPlayer := g.Players[g.Turn]
	if currentPlayer.Client != client {
		client.sendError("Not your turn")
		return
	}

	// Find the lowest empty row in the column
	row := -1
	for r := Rows - 1; r >= 0; r-- {
		if g.Board[r][column] == 0 {
			row = r
			break
		}
	}

	if row == -1 {
		client.sendError("Column is full")
		return
	}

	// Make the move
	g.Board[row][column] = currentPlayer.Piece
	g.MoveCount++

	// Broadcast move to both players
	g.BroadcastToPlayers(map[string]any{
		"type":   "game_move",
		"row":    row,
		"column": column,
		"piece":  currentPlayer.Piece,
		"turn":   g.Turn,
	})

	// Check for win
	if g.checkWin(row, column, currentPlayer.Piece) {
		g.Winner = currentPlayer
		g.IsFinished = true
		g.BroadcastToPlayers(map[string]any{
			"type":   "game_over",
			"winner": currentPlayer.Client.Username,
			"piece":  currentPlayer.Piece,
		})
		return
	}

	// Check for draw
	if g.MoveCount >= Rows*Cols {
		g.IsFinished = true
		g.BroadcastToPlayers(map[string]any{
			"type": "game_over",
			"draw": true,
		})
		return
	}

	// Switch turns
	g.Turn = 1 - g.Turn

	// If it's bot's turn, make bot move
	if g.Players[g.Turn].Client.IsBot {
		go g.makeBotMove()
	} else {
		// Notify the next player it's their turn
		g.Players[g.Turn].Client.sendMessage(map[string]any{
			"type": "your_turn",
		})
	}
}

func (g *Game) makeBotMove() {
	if g.Bot == nil {
		return
	}

	// Bot delay for realism
	time.Sleep(500 * time.Millisecond)

	column := g.Bot.GetMove(g.Board)

	// Find the lowest empty row in the column
	row := -1
	for r := Rows - 1; r >= 0; r-- {
		if g.Board[r][column] == 0 {
			row = r
			break
		}
	}

	if row == -1 {
		// Column full, try another
		for c := range Cols {
			for r := Rows - 1; r >= 0; r-- {
				if g.Board[r][c] == 0 {
					row = r
					column = c
					break
				}
			}
			if row != -1 {
				break
			}
		}
	}

	if row == -1 {
		// Board is full
		g.IsFinished = true
		g.BroadcastToPlayers(map[string]any{
			"type": "game_over",
			"draw": true,
		})
		return
	}

	botPlayer := g.Players[g.Turn]
	g.Board[row][column] = botPlayer.Piece
	g.MoveCount++

	// Broadcast move
	g.BroadcastToPlayers(map[string]any{
		"type":   "game_move",
		"row":    row,
		"column": column,
		"piece":  botPlayer.Piece,
		"turn":   g.Turn,
	})

	// Check for win
	if g.checkWin(row, column, botPlayer.Piece) {
		g.Winner = botPlayer
		g.IsFinished = true
		g.BroadcastToPlayers(map[string]any{
			"type":   "game_over",
			"winner": "Bot",
			"piece":  botPlayer.Piece,
		})
		return
	}

	// Check for draw
	if g.MoveCount >= Rows*Cols {
		g.IsFinished = true
		g.BroadcastToPlayers(map[string]any{
			"type": "game_over",
			"draw": true,
		})
		return
	}

	// Switch back to human player's turn
	g.Turn = 1 - g.Turn
	g.Players[g.Turn].Client.sendMessage(map[string]any{
		"type": "your_turn",
	})
}

// checkWin checks if the last move won the game
func (g *Game) checkWin(row, col, piece int) bool {
	// Check horizontal
	count := 0
	for c := max(0, col-3); c <= min(Cols-1, col+3); c++ {
		if g.Board[row][c] == piece {
			count++
			if count >= 4 {
				return true
			}
		} else {
			count = 0
		}
	}

	// Check vertical
	count = 0
	for r := max(0, row-3); r <= min(Rows-1, row+3); r++ {
		if g.Board[r][col] == piece {
			count++
			if count >= 4 {
				return true
			}
		} else {
			count = 0
		}
	}

	// Check diagonal (top-left to bottom-right)
	count = 0
	for i := -3; i <= 3; i++ {
		r, c := row+i, col+i
		if r >= 0 && r < Rows && c >= 0 && c < Cols {
			if g.Board[r][c] == piece {
				count++
				if count >= 4 {
					return true
				}
			} else {
				count = 0
			}
		}
	}

	// Check diagonal (top-right to bottom-left)
	count = 0
	for i := -3; i <= 3; i++ {
		r, c := row+i, col-i
		if r >= 0 && r < Rows && c >= 0 && c < Cols {
			if g.Board[r][c] == piece {
				count++
				if count >= 4 {
					return true
				}
			} else {
				count = 0
			}
		}
	}

	return false
}

// HandleDisconnect handles a player disconnecting from the game
func (g *Game) HandleDisconnect(client *Client) {
	if g.IsFinished {
		return
	}

	g.IsFinished = true

	// Find the other player
	for _, player := range g.Players {
		if player.Client != client {
			// Notify the remaining player
			player.Client.sendMessage(map[string]any{
				"type":   "game_over",
				"winner": player.Client.Username,
				"piece":  player.Piece,
				"reason": "opponent_disconnected",
			})
			player.Client.GameID = ""
			break
		}
	}

	client.GameID = ""
}

// HandleLeave handles a player voluntarily leaving the game
func (g *Game) HandleLeave(client *Client) {
	if g.IsFinished {
		return
	}

	g.IsFinished = true

	// Find the other player
	for _, player := range g.Players {
		if player.Client != client {
			// Notify the remaining player
			player.Client.sendMessage(map[string]any{
				"type":   "game_over",
				"winner": player.Client.Username,
				"piece":  player.Piece,
				"reason": "opponent_left",
			})
			player.Client.GameID = ""
			break
		}
	}

	client.GameID = ""
}

// BroadcastToPlayers sends a message to both players in the game
func (g *Game) BroadcastToPlayers(msg map[string]any) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for _, player := range g.Players {
		if !player.Client.IsBot {
			select {
			case player.Client.Send <- data:
			default:
				// Buffer full
			}
		}
	}
}
