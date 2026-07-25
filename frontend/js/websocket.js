// websocket.js - WebSocket connection handling

const WS = {
    ws: null,
    reconnectAttempts: 0,
    maxReconnectAttempts: 5,
    reconnectDelay: 1000,
    handlers: {},

    // Connect to WebSocket server
    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const host = window.location.host;
        const url = `${protocol}//${host}/ws`;

        this.ws = new WebSocket(url);

        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.reconnectAttempts = 0;
            this.emit('connected');
        };

        this.ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                this.handleMessage(data);
            } catch (e) {
                console.error('Failed to parse message:', e);
            }
        };

        this.ws.onclose = () => {
            console.log('WebSocket disconnected');
            this.emit('disconnected');
            this.attemptReconnect();
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    },

    // Attempt to reconnect
    attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            setTimeout(() => {
                console.log(`Reconnecting (attempt ${this.reconnectAttempts})...`);
                this.connect();
            }, this.reconnectDelay * this.reconnectAttempts);
        }
    },

    // Send a message to the server
    send(data) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(data));
        }
    },

    // Handle incoming messages
    handleMessage(data) {
        switch (data.type) {
            case 'user_created':
                this.emit('user_created', data);
                break;
            case 'pong':
                this.emit('pong', data);
                break;
            case 'players_list':
                this.emit('players_list', data);
                break;
            case 'game_start':
                this.emit('game_start', data);
                break;
            case 'game_move':
                this.emit('game_move', data);
                break;
            case 'your_turn':
                this.emit('your_turn', data);
                break;
            case 'game_over':
                this.emit('game_over', data);
                break;
            case 'invite_received':
                this.emit('invite_received', data);
                break;
            case 'invite_sent':
                this.emit('invite_sent', data);
                break;
            case 'invite_declined':
                this.emit('invite_declined', data);
                break;
            case 'emoji_reaction':
                this.emit('emoji_reaction', data);
                break;
            case 'error':
                this.emit('error', data);
                break;
            default:
                console.log('Unknown message type:', data.type);
        }
    },

    // Event handler registration
    on(event, handler) {
        if (!this.handlers[event]) {
            this.handlers[event] = [];
        }
        this.handlers[event].push(handler);
    },

    // Emit event to handlers
    emit(event, data) {
        if (this.handlers[event]) {
            this.handlers[event].forEach(handler => handler(data));
        }
    },

    // Start ping interval to keep connection alive
    startPingInterval() {
        setInterval(() => {
            this.send({ type: 'ping' });
        }, 60000); // Every 60 seconds
    },

    // Set username
    setUsername(username) {
        this.send({ type: 'new_user', username });
    },

    // Find a match
    findMatch() {
        this.send({ type: 'find_match' });
    },

    // Cancel search
    cancelSearch() {
        this.send({ type: 'cancel_search' });
    },

    // Make a game move
    makeMove(column) {
        this.send({ type: 'game_move', column });
    },

    // Send invite to a player
    sendInvite(to) {
        this.send({ type: 'send_invite', to });
    },

    // Accept an invite
    acceptInvite(from) {
        this.send({ type: 'accept_invite', from });
    },

    // Decline an invite
    declineInvite(from) {
        this.send({ type: 'decline_invite', from });
    },

    // Leave current game
    leaveGame() {
        this.send({ type: 'leave_game' });
    },

    // Send emoji reaction
    sendEmoji(emoji) {
        this.send({ type: 'emoji_reaction', emoji });
    }
};
