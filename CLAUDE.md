# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

Fight the Landlord (斗地主) — a fair, open-source 3-player card game with networked multiplayer, implemented in Go. Client-server architecture via WebSocket with protobuf serialization. The client is a TUI built on Bubble Tea. The server uses Redis for persistence (rooms, leaderboard, reconnect tokens). Bots can be driven by either a heuristic engine or the DouZero deep-reinforcement-learning AI (separate Python service).

## Build and development

```bash
# Server
go run ./cmd/server              # starts WebSocket server on :1780
go build -o server ./cmd/server

# Client (TUI)
go run ./cmd/client              # connects to localhost:1780 by default
go run ./cmd/client -server wss://example.com/ws

# Run all tests with race detector
go test -race ./...
# Run a single package's tests
go test -race ./internal/game/rule/...
# Run a single test
go test -race ./internal/game/rule/... -run TestCanBeat

# Lint (golangci-lint v2, config at .golangci.yml)
golangci-lint run

# Generate coverage report
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html

# Regenerate protobuf code and message-type mapping
make proto
```

Makefile targets: `make lint`, `make test`, `make coverage`, `make proto`, `make release`.

## Local dev setup

1. Start Redis: `redis-server`
2. (Optional) Start DouZero AI: `cd douzero && uv sync && uv run python server.py`
3. Start server: `go run ./cmd/server`
4. Start client: `go run ./cmd/client`

Or use `docker compose up -d` which brings up redis, douzero, and the game server together.

## High-level architecture

**Server** (`cmd/server/main.go` → `internal/server/`):
- `Server` struct holds all subsystems and implements `types.ServerInterface` (to break circular deps).
- `connection.go` handles WebSocket upgrade, read/write pumps, and message dispatch through `handler.Handler`.
- `handler/` maps `protocol.MessageType` → handler functions. `game.go` routes in-game actions to `GameSession` methods. `room.go` handles room CRUD. `connection.go` handles connect/disconnect/reconnect.
- `session/` contains the full game state machine: `bid.go` (bidding/grabbing landlord), `play.go` (card play validation + broadcasting), `game.go` (deal, start, end, scoring with spring/anti-spring), `player.go` (player tracking), `lifecycle.go` (offline/timeout handling), `timer.go`.
- `storage/` — `RedisStore` persists room data; `LeaderboardManager` manages Redis sorted-set leaderboards (total/daily/weekly).
- `security.go` — rate limiting, origin checking, message limits, IP filtering.

**Client** (`cmd/client/main.go` → `internal/`):
- `transport/` — WebSocket client with read/write pumps, send/receive channels, automatic reconnect with exponential backoff, heartbeat pings.
- `ui/` — Bubble Tea TUI model. `ui.go` is the root model that switches between lobby, room, and game screens.
- `client/` — `game_state.go` tracks the current game from the client's perspective; `card_counter.go` tracks which cards have been played (the "card memorizer" feature, toggled with `C`).
- `update/` — self-update mechanism. Client checks `/version` against its own version and downloads new releases.
- `sound/` — audio playback for game events (beep library, with noop fallback).
- `logger/` — file-based logging for debugging client crashes.

**Game engine** (`internal/game/`):
- `card/` — `Card` struct (Suit + Rank), `Deck` (shuffle/deal), and `Hand` utilities.
- `rule/` — hand-type classification (15 types from Single to Rocket), `CanBeat` comparison, `FindSmallestBeatingCards` for AI/hints, `FindAllPlayable` for generating all legal moves. Entry point is `ParseHand(cards)` which returns a `ParsedHand` with Type/KeyRank/Length.
- `room/` — `Room` (3-player, code-based), `RoomManager` (create/join/leave, cleanup goroutine for idle rooms), Redis serialization via `RoomData`.
- `match/` — `Matcher` with queue-based matchmaking. When the queue has < 3 players after a configurable timeout, bots fill remaining seats. Also handles `PracticeMatch` (1 human + 2 bots).

**Bots** (`internal/bot/`):
- `BotClient` implements `types.ClientInterface` — the server treats bots identically to human players. Bots receive the same game messages and call `SessionInterface` methods (HandleBid, HandlePlayCards, HandlePass) after a randomized think delay.
- `DecisionEngine` interface with `DecidePlay` and `DecideBid`. Two implementations:
  - `HeuristicEngine` — rule-based, finds smallest beating hand, scores hands for bidding (bombs + high cards ≥ threshold).
  - `DouZeroEngine` — calls a Python HTTP service (`douzero/server.py`) running the DouZero neural network.

**Protocol** (`internal/protocol/`):
- `message.go` defines all `MessageType` constants (client→server and server→client).
- `payloads.go` defines Go structs for every message payload.
- `codec/` — protobuf encode/decode with `proto.Unmarshal`. Uses object pools (`pool.go`) for Message and PBMessage to reduce GC pressure.
- `convert/` — card conversion (`CardInfo` ↔ `Card`), payload encoding (`payloadconv.EncodePayload`/`DecodePayload`), message-type mapping.
- `pb/` — generated protobuf Go code. Regenerated via `make proto`.

**Configuration** (`config.yaml` + env vars):
- YAML config loaded by `internal/config/config.go`. Environment variables (loaded from `.env.local` for local dev) override YAML values. `.env` is Docker-only and not loaded locally.
- Key settings: server host/port, Redis addr, game timeouts, security rate limits, bot/DouZero configuration.

## Key patterns

- **Circular dependency breaking**: `internal/types/interfaces.go` defines `ServerInterface` and `ClientInterface` consumed by `handler`, `room`, `match`, and `bot` packages.
- **Testing**: `testutil/` provides mock implementations. `miniredis` is used for Redis tests. Room tests use `testing.go` helpers in each package. Session tests use real `GameSession` instances.
- **Protobuf serialization**: Messages on the wire are protobuf; the `codec` package abstracts this so handler code works with Go structs via `codec.ParsePayload[T](msg)`.
