# Duverge

[English](README.md) | [中文](README_ch.md)

Duverge is an experimental multiplayer-game backend. It is split into a Gateway, Room, and UserData service and currently implements user registration/login, session management, room membership, ready state, game state transitions, TCP heartbeats, and room/private chat.

> The repository is under active development. The API and storage contracts below describe the current implementation, not a stable production API.

## Architecture

```text
Client
  ├── HTTP :26526 ───────┐
  └── TCP  :26527         │
                         ▼
                    Gateway (Go)
                  ┌──────┴──────┐
                  │             │
          gRPC :26529     gRPC :26528
                  │             │
              UserData       Room (Go)
                  │             │
                MySQL ◄──── Redis :6379
```

| Service | Stack | Container port | Responsibility |
|---|---|---:|---|
| `gateway` | Go 1.26 | `26526` HTTP, `26527` TCP | External entry point, sessions, TCP auth, heartbeat, chat, room HTTP API |
| `userdata` | Java 25 / Spring Boot | `26529` gRPC | User registration, BCrypt password verification, login-session generation |
| `room` | Go 1.26 | `26528` gRPC | Rooms, membership, ready state, game state |
| `redis` | Redis | `6379` | Sessions, user status, rooms |
| `mysql` | MySQL 8.4 | `3306` | User persistence |
| `adminer` | Adminer | `8080` | Development database UI |
| `redisinsight` | RedisInsight | `5540` | Development Redis UI |
| `game` | Not enabled | - | Compose entry is commented out; TCP `game` messages are currently a placeholder |

Inside the Compose network, services use:

```text
redis:6379
userdata:26529
room:26528
mysql:3306
```

For the complete field-level API, Redis, and MySQL data dictionary, see the [Chinese API reference](README_ch.md).

## Repository layout

```text
Duverge/
├── docker-compose.yml
├── .env
├── gateway/                 # Go external Gateway
├── room/                    # Go Room gRPC service
├── userData/userdata/       # Java UserData gRPC service
├── proto/                   # Protobuf source files
├── mysql/data/              # Development MySQL data directory
└── scripts/                 # Black-box and utility scripts
```

## Requirements and configuration

- Docker and Docker Compose
- Git
- Optional local development tools: Go 1.26, JDK 25, Maven

Create `.env` next to `docker-compose.yml`:

```env
MYSQL_ROOT_PASSWORD=root
MYSQL_DATABASE=user_data
MYSQL_USER=rin
MYSQL_PASSWORD=change-me
```

Do not commit real credentials. `MYSQL_DATABASE`, `MYSQL_USER`, and `MYSQL_PASSWORD` are passed to MySQL and UserData by Compose.

## Run

```bash
docker compose up -d --build
docker compose ps
docker compose logs -f gateway
docker compose down
```

Health check:

```bash
curl -i http://127.0.0.1:26526/health
# 200 OK
# ok
```

To clear the development Redis database, which deletes every key in the selected Redis database:

```bash
docker compose exec redis redis-cli FLUSHALL
```

## HTTP API

Gateway HTTP base URL:

```text
http://127.0.0.1:26526
```

All JSON requests use `Content-Type: application/json`.

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/health` | Returns `ok` |
| `POST` | `/login` | Register or login, selected by `type` |
| `POST` | `/login_session` | Validate a session |
| `POST` | `/room/user` | Create a room |
| `POST` | `/room/room` | Get room info, join, leave, or start a game |
| `POST` | `/room/userPreparation` | Set or cancel ready state |
| `POST` | `/room/endGame` | Internal game-end callback |

### Register

```http
POST /login
```

```json
{
  "type": "register",
  "accountId": "alice",
  "password": "secret"
}
```

Success:

```json
{"success": true}
```

### Login

```http
POST /login
```

```json
{
  "type": "login",
  "accountId": "alice",
  "password": "secret"
}
```

Success:

```json
{
  "success": true,
  "sessionId": "<opaque-session-id>"
}
```

### Validate a session

```http
POST /login_session
```

```json
{"sessionId": "<opaque-session-id>"}
```

Success:

```json
{"success": true}
```

### Room operations

Create:

```json
{"type": "createRoom", "sessionId": "<session-id>"}
```

```json
{"roomId": "12345"}
```

Get members:

```json
{
  "type": "getRoomInfo",
  "sessionId": "<session-id>",
  "roomId": "12345"
}
```

```json
{"success": true, "roomId": ["alice", "bob"]}
```

> The current response uses `roomId` for the user-ID array. This is a legacy field name; `userIds` would be clearer.

Join, leave, and start:

```json
{
  "type": "joinRoom",
  "sessionId": "<session-id>",
  "roomId": "12345"
}
```

```json
{"result": true}
```

Use `leaveRoom` or `startGame` as the `type` for the other operations. A room has a maximum size of 4. Starting requires the host, at least two players, a waiting room, and every member prepared.

Set ready:

```json
{
  "type": "setPreparation",
  "sessionId": "<session-id>",
  "roomId": "12345",
  "targetStatus": true
}
```

```json
{"status": true}
```

End-game callback:

```json
{
  "type": "endGame",
  "roomId": "12345"
}
```

This endpoint is intended for an internal game service and currently has no service authentication.

## TCP API

Gateway TCP address: `127.0.0.1:26527`.

Each message is framed as:

```text
[4-byte big-endian body length][JSON body]
```

The body must be between 1 byte and 64 KiB. Every message uses protocol version `"1.0"`:

```json
{
  "version": "1.0",
  "request_id": "req-001",
  "type": "heartbeat",
  "data": null
}
```

Authenticate with a login session:

```json
{
  "version": "1.0",
  "request_id": "auth-001",
  "type": "auth",
  "data": {"token": "<session-id>"}
}
```

Heartbeat:

```json
{
  "version": "1.0",
  "request_id": "hb-001",
  "type": "heartbeat",
  "data": null
}
```

Heartbeat renews the `user_status:<userId>` TTL to 30 seconds; it does not change the status value. Clients should send it before expiry, preferably every 10 seconds.

Room chat:

```json
{
  "version": "1.0",
  "request_id": "chat-001",
  "type": "chat",
  "data": {
    "type": "room",
    "data": {
      "room_id": "12345",
      "type": "text",
      "content": "hello room"
    }
  }
}
```

Private chat uses `"type": "private"` and an inner `"to"` user ID. Chat is currently delivered only to online TCP connections and is not persisted.

TCP error codes currently include:

| Code | Meaning |
|---|---|
| `INVALID_MESSAGE` | Invalid JSON, unsupported type, or invalid chat fields |
| `UNAUTHENTICATED` | Non-auth message before authentication |
| `AUTHENTICATION_FAILED` | Invalid or expired login session |
| `FORBIDDEN` | Repeated authentication |
| `NOT_FOUND` | Missing room or offline private-chat target |
| `VERSION_MISMATCH` | Unsupported protocol version |
| `INTERNAL_ERROR` | Redis, network, or unclassified service error |
| `SERVER_CLOSED_ERROR` | EOF or unexpected connection close |

## Internal gRPC APIs

### UserData service

Source: [`proto/user_info/login.proto`](proto/user_info/login.proto)
Address: `userdata:26529`

```protobuf
service UserDataService {
  rpc UserRegistration(UserRegistrationRequest)
      returns (UserRegistrationResponse);
  rpc UserLogin(UserLoginRequest)
      returns (UserLoginResponse);
}
```

`UserRegistration` creates a user and returns `result`. `UserLogin` verifies the BCrypt password and returns `result` plus a random Base64URL session ID.

### Room service

Source: [`proto/room/room_4_gateway.proto`](proto/room/room_4_gateway.proto)
Address: `room:26528`

RPCs:

```text
GetRoomInfo
CreateRoom
JoinRoom
LeaveRoom
SetPreparation
StartGame
EndGame
```

## Redis data dictionary

### Sessions

| Key | Type | TTL | Value |
|---|---|---:|---|
| `session:<sessionId>` | String | 7 days | `session_account:<accountId>` |
| `session_account:<accountId>` | String | 7 days | `session:<sessionId>` |

The two keys are a reverse index. A new login removes the previous session for that account.

### User status

Key: `user_status:<userId>`
Type: JSON string
Initial TTL: 30 seconds
Heartbeat: resets TTL to 30 seconds

```json
{
  "user_id": "alice",
  "status": "in_lobby"
}
```

Statuses are `in_lobby`, `in_room`, and `in_game`. Room operations may rewrite the key with no expiration.

### Rooms

Key: `room:<roomId>`
Type: JSON string
TTL: none

```json
{
  "roomId": "12345",
  "maximumSize": 4,
  "userIds": ["alice", "bob"],
  "hosterId": "alice",
  "prepared": ["alice"],
  "status": "waiting"
}
```

Room statuses are `waiting`, `in_game`, and `finished`. Empty rooms are deleted.

## MySQL data dictionary

Database: `${MYSQL_DATABASE}`, normally `user_data`.

Table: `users`

| Column | Meaning |
|---|---|
| `id` | Auto-increment primary key |
| `account_id` | Login account |
| `password` | BCrypt password hash |

JPA currently uses `ddl-auto: update` for development. UserData exposes registration and login through gRPC only.

## State model

```text
in_lobby ── joinRoom ──► in_room ── startGame ──► in_game
   ▲                         │                      │
   └──────── leaveRoom ◄─────┘                      │
                         endGame ◄─────────────────┘
```

## Development and testing

```bash
cd gateway && go test ./...
cd ../room && go test ./...
go run scripts/heartbeat_blackbox.go
```

The black-box flow exercises login, TCP authentication, periodic heartbeat, room presence, room chat, and private chat. Run it only against an isolated development stack.

## Security and known limitations

- Redis has no password or TLS in the default Compose setup.
- HTTP and gRPC are not configured with TLS.
- `/room/endGame` has no internal authentication.
- Online TCP connections are kept in one Gateway process's memory and are not shared across replicas.
- Chat is not persisted.
- Room keys have no expiration until an empty room is deleted.
- Production deployments should use schema migrations instead of `ddl-auto: update`.
- Add input-length validation, rate limiting, session revocation, replay protection, and fine-grained authorization before production use.
- The Game service is not enabled; TCP `game` is currently a placeholder.

## License

MIT License: see [LICENSE](LICENSE).
