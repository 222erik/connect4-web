// storage.js - Local storage for match stats

const Storage = {
    STATS_KEY: 'connect4_stats',

    // Get stats for a username (returns a copy to avoid shared-reference mutations)
    getStats(username) {
        try {
            const stats = JSON.parse(localStorage.getItem(this.STATS_KEY) || '{}');
            if (stats[username]) {
                return { wins: stats[username].wins, plays: stats[username].plays };
            }
        } catch (e) {
            console.error('Failed to parse stats:', e);
        }
        return { wins: 0, plays: 0 };
    },

    // Update stats after a game
    updateStats(username, won) {
        if (!username) return;

        let stats = {};
        try {
            stats = JSON.parse(localStorage.getItem(this.STATS_KEY) || '{}');
        } catch (e) {
            console.error('Failed to parse stats, resetting:', e);
        }

        if (!stats[username]) {
            stats[username] = { wins: 0, plays: 0 };
        }

        stats[username].plays++;
        if (won) {
            stats[username].wins++;
        }

        try {
            localStorage.setItem(this.STATS_KEY, JSON.stringify(stats));
        } catch (e) {
            console.error('Failed to save stats:', e);
        }
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
