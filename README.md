# 🎯 Connect Four

A real-time multiplayer Connect Four web app. Play against friends online or a bot.

## ✨ Features

- 🌐 **Multiplayer** — Invite online players to a match via WebSockets
- 🤖 **Bot fallback** — Jump straight into a game against a random-move bot
- 🎭 **Anonymous mode** — Skip the username and play the bot instantly
- 😄 **Emoji reactions** — React to moves with short emoji like "GG", "Nice!", "Hmm..."
- 👥 **Player list** — See who's online in real time
- 🌗 **Light/dark theme** — Toggle from the menu
- 🎬 **Animations** — Piece drops and transitions

## 🧰 Tech Stack

| Layer | Tech |
|-------|------|
| ⚙️ Server | Go, [gorilla/websocket](https://github.com/gorilla/websocket) |
| 🖥️ Frontend | Vanilla HTML, CSS, JavaScript (no frameworks) |
| 🔌 Transport | WebSocket |
| 💾 Storage | In-memory (server), localStorage (client) |

## 🚀 Getting Started

It is on the web! https://logistics-dealers-watches-interventions.trycloudflare.com

## 📁 Project Structure

```
├── main.go                  # Entry point — serves frontend, registers routes
├── server/
│   ├── hub.go               # Central hub — manages clients, games, invites
│   ├── client.go            # WebSocket connection — message dispatch & handlers
│   ├── game.go              # Game logic — board, moves, win detection
│   └── bot.go               # Random-move bot
├── frontend/
│   ├── index.html           # Single-page app shell
│   ├── css/
│   │   ├── style.css        # Layout and components
│   │   ├── themes.css       # Light/dark theme variables
│   │   └── animations.css   # Piece drops and transitions
│   └── js/
│       ├── storage.js       # localStorage wrapper (username, theme, stats)
│       ├── websocket.js     # WebSocket client with auto-reconnect
│       ├── game.js          # Board rendering and game UI
│       ├── ui.js            # Screen transitions, modals, notifications
│       └── app.js           # Wires everything together
├── notify-usernames.sh      # Polls API for new usernames (optional)
├── go.mod
└── go.sum
```

## 🔁 How It Works

1. 👤 Enter a username (or play anonymously) — the server stores it in memory
2. 📋 View the online player list or hit **Find Match** to play a bot
3. ✉️ Click **Invite** on a player to send a match request
4. 🎲 Click columns to drop pieces — standard Connect Four rules (6×7 board, 4 in a row wins)
5. 😄 React to your opponent's moves with emoji
6. ⚠️ All game state lives in server memory — disconnect ends the session

## 🌐 API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/ws` | WebSocket | 🎮 Game and chat transport |
| `/api/usernames` | GET | 👥 Returns a JSON array of online usernames |

## 🛠️ Scripts

| Script | Description |
|--------|-------------|
| `notify-usernames.sh` | 🔔 Polls `/api/usernames` and alerts on new users (default: 5s interval) |

```bash
./notify-usernames.sh       # poll every 5s
./notify-usernames.sh 10    # poll every 10s
```

## 📄 License

[MIT](LICENSE)
