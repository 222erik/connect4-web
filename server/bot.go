package server

import (
	"math/rand"
)

// Bot represents a simple AI opponent
type Bot struct {
	// Could add difficulty levels later
}

// NewBot creates a new Bot
func NewBot() *Bot {
	return &Bot{}
}

// GetMove determines the bot's next move
func (b *Bot) GetMove(board [Rows][Cols]int) int {
	// The bots stategy is just to make random moves
	for {
		col := rand.Intn(Cols)
		if board[0][col] == 0 {
			return col
		}
	}
}
