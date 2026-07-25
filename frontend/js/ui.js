// ui.js - UI interactions and updates

const UI = {
    currentScreen: 'login-screen',

    // Show a specific screen
    showScreen(screenId) {
        document.querySelectorAll('.screen').forEach(screen => {
            screen.classList.remove('active');
        });
        document.getElementById(screenId).classList.add('active');
        this.currentScreen = screenId;
    },

    // Show a notification
    showNotification(message, duration = 3000) {
        const notif = document.getElementById('notification');
        notif.textContent = message;
        notif.classList.remove('hidden');

        setTimeout(() => {
            notif.classList.add('hidden');
        }, duration);
    },

    // Show the invite modal
    showInviteModal(from) {
        const modal = document.getElementById('invite-modal');
        document.getElementById('invite-message').textContent = `${from} wants to play!`;
        modal.classList.remove('hidden');
        modal.dataset.from = from;
    },

    // Hide the invite modal
    hideInviteModal() {
        document.getElementById('invite-modal').classList.add('hidden');
    },

    // Show game over modal
    showGameOverModal(won, draw) {
        const modal = document.getElementById('gameover-modal');
        const title = document.getElementById('gameover-title');
        const message = document.getElementById('gameover-message');

        if (draw) {
            title.textContent = 'Draw!';
            message.textContent = 'The game ended in a draw.';
        } else if (won) {
            title.textContent = 'You Won!';
            message.textContent = 'Congratulations!';
        } else {
            title.textContent = 'You Lost';
            message.textContent = 'Better luck next time!';
        }

        // Update stats display
        const username = Storage.getUsername();
        if (username) {
            const stats = Storage.getStats(username);
            document.getElementById('stat-wins').textContent = stats.wins;
            document.getElementById('stat-plays').textContent = stats.plays;
        }

        modal.classList.remove('hidden');
    },

    // Hide game over modal
    hideGameOverModal() {
        document.getElementById('gameover-modal').classList.add('hidden');
    },

    // Update players list
    updatePlayersList(players) {
        const list = document.getElementById('players-list');
        const count = document.getElementById('player-count');
        const currentUsername = Storage.getUsername();

        count.textContent = `(${players.length})`;
        list.innerHTML = '';

        players.forEach(username => {
            if (username === currentUsername || username === 'Bot') return;

            const li = document.createElement('li');
            li.innerHTML = `
                <span>${username}</span>
                <button class="invite-btn" data-username="${username}">Invite</button>
            `;
            list.appendChild(li);
        });

        // Add invite button handlers
        list.querySelectorAll('.invite-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                const to = btn.dataset.username;
                WS.sendInvite(to);
                this.showNotification(`Invite sent to ${to}`);
            });
        });
    },

    // Filter players based on search
    filterPlayers(query) {
        const list = document.getElementById('players-list');
        const items = list.querySelectorAll('li');

        items.forEach(item => {
            const username = item.querySelector('span').textContent.toLowerCase();
            item.style.display = username.includes(query.toLowerCase()) ? 'flex' : 'none';
        });
    },

    // Toggle theme
    toggleTheme() {
        const html = document.documentElement;
        const currentTheme = html.getAttribute('data-theme');
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        html.setAttribute('data-theme', newTheme);
        Storage.setTheme(newTheme);
    },

    // Initialize theme from storage
    initTheme() {
        const theme = Storage.getTheme();
        document.documentElement.setAttribute('data-theme', theme);
    }
};
