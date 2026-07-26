// game.js - Game board logic

const Game = {
    board: [],
    myPiece: null,
    isMyTurn: false,
    gameId: null,
    opponent: null,

    // Initialize the game board
    init() {
        this.board = Array(6).fill(null).map(() => Array(7).fill(0));
        this.renderBoard();
    },

    // Render the game board
    renderBoard() {
        const boardEl = document.getElementById('game-board');
        boardEl.innerHTML = '';

        for (let row = 0; row < 6; row++) {
            for (let col = 0; col < 7; col++) {
                const cell = document.createElement('div');
                cell.className = 'cell';
                cell.dataset.row = row;
                cell.dataset.col = col;

                if (this.board[row][col] === 1) {
                    cell.classList.add('red');
                } else if (this.board[row][col] === 2) {
                    cell.classList.add('yellow');
                }

                cell.addEventListener('click', () => this.handleCellClick(col));
                cell.addEventListener('mouseenter', () => this.previewColumn(col));
                cell.addEventListener('mouseleave', () => this.clearPreview());
                boardEl.appendChild(cell);
            }
        }
    },

    previewColumn(col) {
        if (!this.isMyTurn) return;
        const previewClass = this.myPiece === 1 ? 'preview-red' : 'preview-yellow';
        for (let row = 0; row < 6; row++) {
            if (this.board[row][col] === 0) {
                const idx = row * 7 + col;
                const cell = document.getElementById('game-board').children[idx];
                cell.classList.add(previewClass);
                break;
            }
        }
    },

    clearPreview() {
        for (const cell of document.getElementById('game-board').children) {
            cell.classList.remove('preview-red', 'preview-yellow');
        }
    },

    // Handle cell click
    handleCellClick(col) {
        if (!this.isMyTurn) return;

        // Find the lowest empty row in this column
        for (let row = 5; row >= 0; row--) {
            if (this.board[row][col] === 0) {
                this.isMyTurn = false;
                const indicator = document.getElementById('turn-indicator');
                indicator.textContent = "Waiting for opponent...";
                indicator.style.color = 'var(--text-secondary)';
                WS.makeMove(col);
                return;
            }
        }
    },

    // Update board with a move
    updateBoard(row, col, piece) {
        this.board[row][col] = piece;
        this.renderBoard();
    },

    // Set whose turn it is
    setTurn(isMyTurn) {
        this.isMyTurn = isMyTurn;
        const indicator = document.getElementById('turn-indicator');
        indicator.textContent = isMyTurn ? 'Your Turn' : "Opponent's Turn";
        indicator.style.color = isMyTurn ? 'var(--accent)' : 'var(--text-secondary)';
    },

    // Start a new game
    start(data) {
        this.gameId = data.game_id;
        this.myPiece = data.your_piece;
        this.opponent = data.opponent;
        this.board = Array(6).fill(null).map(() => Array(7).fill(0));

        // Update UI
        document.getElementById('game-opponent').textContent = `vs ${data.opponent}`;

        if (data.your_piece === 1) {
            document.getElementById('player1-name').textContent = App.username || 'You';
            document.getElementById('player2-name').textContent = data.opponent;
        } else {
            document.getElementById('player1-name').textContent = data.opponent;
            document.getElementById('player2-name').textContent = App.username || 'You';
        }

        this.setTurn(data.your_turn);
        this.renderBoard();

        UI.showScreen('game-screen');
    },

    // Handle game over
    gameOver(data) {
        const username = App.username;

        if (data.draw) {
            UI.showNotification('Game ended in a draw!');
            UI.showGameOverModal(false, true);
        } else {
            const won = data.piece === this.myPiece;
            const message = won ? 'You won!' : `${data.opponent || data.winner} wins!`;
            UI.showNotification(message);
            // Glow the winner's pieces
            for (const cell of document.getElementById('game-board').children) {
                if ((data.piece === 1 && cell.classList.contains('red')) ||
                    (data.piece === 2 && cell.classList.contains('yellow'))) {
                    cell.classList.add('winner');
                }
            }
            // Update stats
            if (username) {
                Storage.updateStats(username, won);
            }
            UI.showGameOverModal(won, false);
        }

        // Reset game state
        this.gameId = null;
        this.myPiece = null;
        this.isMyTurn = false;
    },

    // Leave the current game
    leave() {
        WS.leaveGame();
        this.gameId = null;
        this.myPiece = null;
        this.isMyTurn = false;
        UI.showScreen('menu-screen');
    }
};
