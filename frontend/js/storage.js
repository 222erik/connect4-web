// storage.js - Local storage for match stats

const Storage = {
    STATS_KEY: 'connect4_stats',

    // Get stats for a username
    getStats(username) {
        const stats = JSON.parse(localStorage.getItem(this.STATS_KEY) || '{}');
        return stats[username] || { wins: 0, plays: 0 };
    },

    // Update stats after a game
    updateStats(username, won) {
        if (!username) return;

        const stats = JSON.parse(localStorage.getItem(this.STATS_KEY) || '{}');
        if (!stats[username]) {
            stats[username] = { wins: 0, plays: 0 };
        }

        stats[username].plays++;
        if (won) {
            stats[username].wins++;
        }

        localStorage.setItem(this.STATS_KEY, JSON.stringify(stats));
    },

    // Get theme preference
    getTheme() {
        return localStorage.getItem('connect4_theme') || 'dark';
    },

    // Set theme preference
    setTheme(theme) {
        localStorage.setItem('connect4_theme', theme);
    },

    // Get username
    getUsername() {
        return localStorage.getItem('connect4_username') || '';
    },

    // Set username
    setUsername(username) {
        localStorage.setItem('connect4_username', username);
    }
};
