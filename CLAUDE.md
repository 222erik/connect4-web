# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Real-time multiplayer Connect Four web app. Users play against each other (or a bot fallback) via WebSocket.

## Commands

- **Build:** `go build -o connect4 .`
- **Run:** `go run main.go` (listens on port 8080 by default, overridable via `PORT` env var)
- **Clean build:** `go build -o connect4 . && ./connect4`
- **Add dependency:** `go get <module>` then `go mod tidy`

## Architecture

### Go Server (`server/`)

All server code is in the `server` package with `main.go` as the entry point.

**Files:**

- `main.go` — Serves static files from `./frontend/` at `/`, registers the `/ws` WebSocket endpoint, and starts the hub's Run loop.
- `server/hub.go` — Central Hub type that holds all state in-memory: connected clients (map[*Client]bool), active games (map[string]*Game), and pending invites (map[string]string). Thread safety via `sync.RWMutex`. Communication with clients is done through two channels (`register`, `unregister`) processed in a main select loop.
- `server/client.go` — Each WebSocket connection is a `Client`. Contains the WebSocket read/write pumps, message dispatch (`handleMessage`), and handlers for all message types (`handleNewUser`, `handleGameMove`, `handleSendInvite`, etc.). Client is registered in the hub on connect. Bot clients are synthetic `Client` structs with `IsBot: true`.
- `server/game.go` — Game struct holds a 6x7 board, two players, turn state, and win detection. `MakeMove` validates the player, drops the piece, checks for win/draw, switches turns, and triggers bot moves. Win check scans horizontal, vertical, and both diagonals from the last move.
- `server/bot.go` — Simple random-move bot. `GetMove` loops until it finds an empty column.

**Key patterns:**

- `map[string]any` is used for all JSON message construction.
- JSON numbers decode as `float64` — always type-assert as `float64` and convert to `int`.
- `ClientFromUsername` acquires `RLock` — never call it while holding a `Lock` (deadlock risk). Inline the look-up instead when under `Lock`.
- Bot games are created by `startBotGame()` which builds a synthetic `Client{IsBot: true, Username: "Bot"}`. The `find_match` message type triggers this. There is no real matchmaking queue — it goes straight to a bot game.
- No database — everything lives in server memory. Disconnect = data loss.

### Frontend (`frontend/`)

Vanilla JS, no frameworks. All JS is in global singletons loaded via `<script>` tags in order.

- `js/storage.js` — `Storage` singleton wrapping `localStorage` for username, theme, and match stats.
- `js/websocket.js` — `WS` singleton managing the WebSocket connection, auto-reconnect (up to 5 attempts), message dispatch, and outgoing message helpers. Uses an event system (`on`/`emit`) to decouple from the rest of the app.
- `js/game.js` — `Game` singleton handling board rendering (6x7 grid of divs), piece placement, turn indication, and game-over logic.
- `js/ui.js` — `UI` singleton for screen transitions, modals, notifications, player list rendering, and theme toggling.
- `js/app.js` — `App` singleton that wires everything together: sets up event listeners, registers WebSocket handlers, handles login flow.

**Key patterns:**

- `Storage.getUsername()` is used to identify the current player.
- Login sends `set_username` → server responds `username_set`.
- Anonymous play skips username and goes straight to the menu (can't invite or be invited).
- The `emoji-btn` toggles a simple emoji picker — emoji are short strings like "GG", "Nice!", "Hmm...".

### WebSocket Protocol

**Client → Server:**

| Type | Fields | Description |
|---|---|---|
| `new_user` | `username` | Register/set a username |
| `ping` | — | Keepalive (client sends every 60s) |
| `find_match` | — | Start a bot game |
| `game_move` | `column` (0-6) | Drop a piece |
| `send_invite` | `to` | Invite a player |
| `accept_invite` | `from` | Accept an invite from player |
| `decline_invite` | `from` | Decline an invite |
| `leave_game` | — | Forfeit the current game |
| `emoji_reaction` | `emoji` | Send emoji reaction |

**Server → Client:**

| Type | Fields | Description |
|---|---|---|
| `user_created` | `username` | Username was registered |
| `pong` | — | Keepalive response |
| `players_list` | `players` (string array) | Active player names |
| `game_start` | `game_id`, `your_piece`, `opponent`, `your_turn` | Game begins |
| `game_move` | `row`, `column`, `piece`, `turn` | A piece was placed |
| `your_turn` | — | It's now this client's turn |
| `game_over` | `winner`, `piece`, `reason`, `draw` | Game ended |
| `invite_received` | `from` | Someone invited you |
| `invite_sent` | `to` | Your invite was sent |
| `invite_declined` | `from` | Your invite was declined |
| `emoji_reaction` | `from`, `emoji` | Reaction from opponent |
| `error` | `error` | Error message |

### State Flow

1. User enters username → `send({type: "new_user", username})` → server responds `user_created` → player list broadcasts to all clients.
2. "Find Match" button → `send({type: "find_match"})` → server creates a bot game.
3. During game: clicks on column → `send({type: "game_move", column})` → server validates, drops piece, broadcasts `game_move`, checks win, sends `game_over` or `your_turn`.
4. Invite flow: click "Invite" on a player → `send({type: "send_invite", to})` → server forwards as `invite_received` to target → target accepts/declines.
5. Disconnect: client's readPump exits → unregister channel → game cleanup, player removed from list.

## Common Pitfalls

- **JSON number types:** `json.Unmarshal` into `map[string]any` always produces `float64` for numbers. Asserting as `int` silently fails. Always use `.(float64)` then `int()`.
- **Mutex reentrancy:** `sync.RWMutex` is not reentrant. A method holding `Lock` cannot call a method that acquires `RLock` on the same mutex — this deadlocks. Inline the logic instead.
- **Client registration:** `HandleWebSocket` registers the client in the hub. The username handlers should NOT re-register — just set `c.Username` and call `AddClient`/broadcast.
- **Username not set on handlers:** `handleNewUser` must set `c.Username` or the client will be invisible to `BroadcastOnlinePlayers` (which filters by `Username != ""`).
