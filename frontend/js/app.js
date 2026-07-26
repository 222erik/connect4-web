// app.js - Main application logic

const App = {
    username: null,

    init() {
        // Initialize theme
        UI.initTheme();

        // Check for saved username
        const savedUsername = Storage.getUsername();
        if (savedUsername) {
            this.username = savedUsername;
            document.getElementById('username-input').value = savedUsername;
        }

        // Connect to WebSocket
        WS.connect();

        // Set up event handlers
        this.setupEventListeners();
        this.setupWebSocketHandlers();

        // Start ping interval
        WS.startPingInterval();
    },

    setupEventListeners() {
        // Login
        document.getElementById('login-btn').addEventListener('click', () => this.handleLogin());
        document.getElementById('anonymous-btn').addEventListener('click', () => this.handleAnonymous());
        document.getElementById('username-input').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.handleLogin();
        });

        // Menu
        document.getElementById('find-match-btn').addEventListener('click', async () => {
            const res = await fetch('/api/usernames');
            const players = await res.json();
            const candidates = players.filter(p => p !== this.username && p !== 'Bot');
            if (candidates.length === 0) {
                UI.showNotification('No players to invite');
                return;
            }
            const target = candidates[Math.floor(Math.random() * candidates.length)];
            WS.sendInvite(target);
            UI.showNotification(`Invite sent to ${target}`);
        });
        document.getElementById('bot-match-btn').addEventListener('click', () => {
            WS.findMatch(); // Server will start bot game if no humans available
        });
        document.getElementById('theme-toggle').addEventListener('click', () => UI.toggleTheme());

        // Game
        document.getElementById('leave-game-btn').addEventListener('click', () => Game.leave());
        document.getElementById('emoji-btn').addEventListener('click', () => {
            document.getElementById('emoji-picker').classList.toggle('hidden');
        });
        document.querySelectorAll('.emoji').forEach(btn => {
            btn.addEventListener('click', () => {
                WS.sendEmoji(btn.dataset.emoji);
                document.getElementById('emoji-picker').classList.add('hidden');
            });
        });

        // Invite modal
        document.getElementById('accept-invite-btn').addEventListener('click', () => {
            const from = document.getElementById('invite-modal').dataset.from;
            WS.acceptInvite(from);
            UI.hideInviteModal();
        });
        document.getElementById('decline-invite-btn').addEventListener('click', () => {
            const from = document.getElementById('invite-modal').dataset.from;
            WS.declineInvite(from);
            UI.hideInviteModal();
        });

        // Game over modal
        document.getElementById('play-again-btn').addEventListener('click', () => {
            UI.hideGameOverModal();
            UI.showScreen('menu-screen');
            this.fetchPlayers();
            this.updateMenuStats();
        });
    },

    setupWebSocketHandlers() {
        WS.on('connected', () => {
            console.log('Connected to server');
            // If we have a username, set it
            if (this.username) {
                WS.setUsername(this.username);
            }
        });

        WS.on('disconnected', () => {
            UI.showNotification('Disconnected from server');
        });

        WS.on('user_created', (data) => {
            this.username = data.username;
            Storage.setUsername(data.username);
            document.getElementById('player-name').textContent = data.username;
            document.getElementById('menu-screen').classList.remove('anonymous');
            UI.showScreen('menu-screen');
            this.fetchPlayers();
            this.updateMenuStats();
        });

        WS.on('players_list', (data) => {
            UI.updatePlayersList(data.players);
        });

        WS.on('game_start', (data) => {
            Game.start(data);
        });

        WS.on('game_move', (data) => {
            Game.updateBoard(data.row, data.column, data.piece);
        });

        WS.on('your_turn', () => {
            Game.setTurn(true);
        });

        WS.on('game_over', (data) => {
            Game.gameOver(data);
        });

        WS.on('invite_received', (data) => {
            UI.showInviteModal(data.from);
        });

        WS.on('invite_sent', (data) => {
            UI.showNotification(`Invite sent to ${data.to}`);
        });

        WS.on('invite_declined', (data) => {
            UI.showNotification(`${data.from} declined your invite`);
        });

        WS.on('emoji_reaction', (data) => {
            UI.showNotification(`${data.from}: ${data.emoji}`);
        });

        WS.on('error', (data) => {
            UI.showNotification(`Error: ${data.error}`);
        });
    },

    handleLogin() {
        const input = document.getElementById('username-input');
        const username = input.value.trim();

        if (!username) {
            UI.showNotification('Please enter a username');
            return;
        }

        if (username.length < 2) {
            UI.showNotification('Username must be at least 2 characters');
            return;
        }

        WS.setUsername(username);
    },

    handleAnonymous() {
        this.username = null;
        document.getElementById('player-name').textContent = 'Anonymous';
        document.getElementById('menu-screen').classList.add('anonymous');
        UI.showScreen('menu-screen');
        this.fetchPlayers();
        this.updateMenuStats();
        UI.showNotification('Playing anonymously (can only play against bot)');
    },

    async fetchPlayers() {
        try {
            const res = await fetch('/api/usernames');
            const players = await res.json();
            UI.updatePlayersList(players);
        } catch (e) {
            console.error('Failed to fetch players:', e);
        }
    },

    updateMenuStats() {
        const username = Storage.getUsername();
        if (!username) return;
        const stats = Storage.getStats(username);
        document.getElementById('menu-stat-plays').textContent = `${stats.plays} plays`;
        document.getElementById('menu-stat-wins').textContent = `${stats.wins} wins`;
    }
};

// Start the app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => App.init());
