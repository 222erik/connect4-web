package server

import "math"

// Bot represents an AI opponent using minimax with alpha-beta pruning.
type Bot struct{}

// NewBot creates a new Bot
func NewBot() *Bot {
	return &Bot{}
}

// GetMove determines the bot's next move using minimax.
func (b *Bot) GetMove(board [Rows][Cols]int) int {
	validCols := getValidColumns(board)
	if len(validCols) == 0 {
		return 3 // fallback — should never happen
	}

	// Order columns center-first for better pruning
	orderColumns(validCols)

	bestScore := math.MinInt
	bestCol := validCols[0]

	for _, col := range validCols {
		row := dropPiece(board, col, 2)
		if row == -1 {
			continue
		}
		score := minimax(board, 5, math.MinInt, math.MaxInt, false)
		undoPiece(board, row, col)
		if score > bestScore {
			bestScore = score
			bestCol = col
		}
	}

	return bestCol
}

const (
	winScore     = 1000
	threeOpen    = 5
	twoOpen      = 2
	centerWeight = 3
)

// minimax evaluates the board recursively with alpha-beta pruning.
func minimax(board [Rows][Cols]int, depth int, alpha, beta int, isMaximizing bool) int {
	// Terminal checks
	if hasWon(board, 2) {
		return winScore + depth // prefer faster wins
	}
	if hasWon(board, 1) {
		return -(winScore + depth) // prefer slower losses
	}
	if depth == 0 {
		return evaluate(board)
	}

	validCols := getValidColumns(board)
	if len(validCols) == 0 {
		return 0 // draw
	}

	orderColumns(validCols)

	if isMaximizing {
		best := math.MinInt
		for _, col := range validCols {
			row := dropPiece(board, col, 2)
			if row == -1 {
				continue
			}
			score := minimax(board, depth-1, alpha, beta, false)
			undoPiece(board, row, col)
			if score > best {
				best = score
			}
			if best > alpha {
				alpha = best
			}
			if alpha >= beta {
				break
			}
		}
		return best
	}

	best := math.MaxInt
	for _, col := range validCols {
		row := dropPiece(board, col, 1)
		if row == -1 {
			continue
		}
		score := minimax(board, depth-1, alpha, beta, true)
		undoPiece(board, row, col)
		if score < best {
			best = score
		}
		if best < beta {
			beta = best
		}
		if alpha >= beta {
			break
		}
	}
	return best
}

// evaluate scores the board from the bot's (piece 2) perspective.
func evaluate(board [Rows][Cols]int) int {
	score := 0

	// Center column preference
	for r := 0; r < Rows; r++ {
		if board[r][3] == 2 {
			score += centerWeight
		} else if board[r][3] == 1 {
			score -= centerWeight
		}
	}

	// Evaluate all windows of 4
	// Horizontal
	for r := 0; r < Rows; r++ {
		for c := 0; c <= Cols-4; c++ {
			score += scoreWindow(board[r][c], board[r][c+1], board[r][c+2], board[r][c+3])
		}
	}

	// Vertical
	for c := 0; c < Cols; c++ {
		for r := 0; r <= Rows-4; r++ {
			score += scoreWindow(board[r][c], board[r+1][c], board[r+2][c], board[r+3][c])
		}
	}

	// Diagonal (top-left to bottom-right)
	for r := 0; r <= Rows-4; r++ {
		for c := 0; c <= Cols-4; c++ {
			score += scoreWindow(board[r][c], board[r+1][c+1], board[r+2][c+2], board[r+3][c+3])
		}
	}

	// Diagonal (top-right to bottom-left)
	for r := 0; r <= Rows-4; r++ {
		for c := 3; c < Cols; c++ {
			score += scoreWindow(board[r][c], board[r+1][c-1], board[r+2][c-2], board[r+3][c-3])
		}
	}

	return score
}

// scoreWindow evaluates a window of 4 cells.
func scoreWindow(a, b, c, d int) int {
	counts := [3]int{} // [empty, piece1, piece2]
	for _, v := range [4]int{a, b, c, d} {
		counts[v]++
	}

	bot := counts[2]
	human := counts[1]
	empty := counts[0]

	// Bot wins
	if bot == 4 {
		return winScore
	}
	// Human wins (shouldn't happen if minimax works, but score defensively)
	if human == 4 {
		return -winScore
	}
	// Bot has 3 + empty — strong threat
	if bot == 3 && empty == 1 {
		return threeOpen
	}
	// Human has 3 + empty — must block
	if human == 3 && empty == 1 {
		return -threeOpen
	}
	// Bot has 2 + 2 empty — building
	if bot == 2 && empty == 2 {
		return twoOpen
	}
	// Human has 2 + 2 empty
	if human == 2 && empty == 2 {
		return -twoOpen
	}

	return 0
}

// hasWon checks if the given piece has 4 in a row anywhere on the board.
func hasWon(board [Rows][Cols]int, piece int) bool {
	// Horizontal
	for r := 0; r < Rows; r++ {
		for c := 0; c <= Cols-4; c++ {
			if board[r][c] == piece && board[r][c+1] == piece &&
				board[r][c+2] == piece && board[r][c+3] == piece {
				return true
			}
		}
	}
	// Vertical
	for c := 0; c < Cols; c++ {
		for r := 0; r <= Rows-4; r++ {
			if board[r][c] == piece && board[r+1][c] == piece &&
				board[r+2][c] == piece && board[r+3][c] == piece {
				return true
			}
		}
	}
	// Diagonal (top-left to bottom-right)
	for r := 0; r <= Rows-4; r++ {
		for c := 0; c <= Cols-4; c++ {
			if board[r][c] == piece && board[r+1][c+1] == piece &&
				board[r+2][c+2] == piece && board[r+3][c+3] == piece {
				return true
			}
		}
	}
	// Diagonal (top-right to bottom-left)
	for r := 0; r <= Rows-4; r++ {
		for c := 3; c < Cols; c++ {
			if board[r][c] == piece && board[r+1][c-1] == piece &&
				board[r+2][c-2] == piece && board[r+3][c-3] == piece {
				return true
			}
		}
	}
	return false
}

// dropPiece places a piece in the given column and returns the row it landed on.
func dropPiece(board [Rows][Cols]int, col, piece int) int {
	for r := Rows - 1; r >= 0; r-- {
		if board[r][col] == 0 {
			board[r][col] = piece
			return r
		}
	}
	return -1
}

// undoPiece removes a piece from the board.
func undoPiece(board [Rows][Cols]int, row, col int) {
	board[row][col] = 0
}

// getValidColumns returns columns that aren't full.
func getValidColumns(board [Rows][Cols]int) []int {
	cols := make([]int, 0, Cols)
	for c := 0; c < Cols; c++ {
		if board[0][c] == 0 {
			cols = append(cols, c)
		}
	}
	return cols
}

// orderColumns sorts columns center-first for better alpha-beta pruning.
func orderColumns(cols []int) {
	center := Cols / 2 // 3
	for i := 0; i < len(cols)-1; i++ {
		for j := i + 1; j < len(cols); j++ {
			if abs(cols[i]-center) > abs(cols[j]-center) {
				cols[i], cols[j] = cols[j], cols[i]
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
