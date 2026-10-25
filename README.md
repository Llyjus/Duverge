# Duverge

[English](README.md) | [中文](README_ch.md)

Duverge is an experimental multiplayer game backend. It contains a Go Gateway, a Go Room service, a Java/Spring UserData service, Redis and MySQL. Prometheus and Grafana are included in Docker Compose for service and load-test monitoring.

> This project is under active development. The API, storage format and operational behavior documented here describe the current code, not a stable production contract.

## Contents

- [Architecture](#architecture)
- [Quick start](#quick-start)
- [Monitoring](#monitoring)
- [HTTP API](#http-api)
- [TCP API](#tcp-api)
- [Timeouts](#timeouts)
- [gRPC services](#grpc-services)
- [Storage](#storage)
- [Password hashing limits](#password-hashing-limits)
- [Development and tests](#development-and-tests)
- [Security and current limitations](#security-and-current-limitations)

## Architecture

```text
HTTP :26526 ─┐
TCP  :26527 ─┴─> Gateway (Go) ──gRPC──> UserData (Java) ──> MySQL
                       │
                       ├────gRPC────> Room (Go)
                       └────────────> Redis <──────────────┘

Prometheus ──scrapes──> Gateway :9091, Room :9091, UserData :9091
Grafana ──────────────> Prometheus
```

| Compose service | Technology | Ports | Responsibility |
|---|---|---|---|
| `gateway` | Go | HTTP `26526`, TCP `26527`, metrics `9091` (internal) | External API, sessions, TCP auth/chat/heartbeat, room HTTP routes |
| `room` | Go | gRPC `26528`, metrics `9091` (internal) | Room membership, ready state and game state |
| `userdata` | Java 25 / Spring Boot | gRPC `26529`, metrics `9091` (host-bound to `127.0.0.1`) | Registration, BCrypt password verification and database access |
| `redis` | Redis | `6379` | Sessions, online user status and room state |
| `mysql` | MySQL 8.4 | `3306` | User records |
| `prometheus` | Prometheus | host `127.0.0.1:9090` | Scrapes and stores service metrics |
| `grafana` | Grafana | host `127.0.0.1:3000` | Dashboards backed by Prometheus |
| `adminer` | Adminer | `8080` | Development database UI |
| `redisinsight` | RedisInsight | `5540` | Development Redis UI |

Services use Compose DNS names internally: `gateway`, `room`, `userdata`, `redis`, and `mysql`. The optional `game` service is not enabled; TCP `game` messages remain a placeholder.

## Quick start

Requirements: Docker and Docker Compose. Go 1.26, JDK 25 and Maven are only needed for local development.

Create `.env` beside `docker-compose.yml`:

```env
MYSQL_ROOT_PASSWORD=root
MYSQL_DATABASE=user_data
MYSQL_USER=rin
MYSQL_PASSWORD=change-me
```

Use local development credentials only. Do not commit real secrets.

Start all application and monitoring services:

```bash
docker compose up -d --build
docker compose ps
```

Check Gateway and service readiness:

```bash
curl -i http://127.0.0.1:26526/health
docker compose logs -f gateway
```

Useful local URLs:

- Gateway health: <http://127.0.0.1:26526/health>
- Prometheus: <http://127.0.0.1:9090>
- Grafana: <http://127.0.0.1:3000>
- UserData Prometheus endpoint: <http://127.0.0.1:9091/actuator/prometheus>

The first Grafana login uses the Grafana image's default `admin` / `admin` credentials unless you have changed them. Change the default password before sharing the instance. Gateway and Room metrics ports are available to Prometheus on the Compose network, but are not published on the host.

Stop containers while preserving named volumes:

```bash
docker compose down
```

This project uses persistent Redis, Prometheus and Grafana volumes, and a MySQL data directory. Avoid `FLUSHALL` or volume deletion unless you intend to erase development data.

## Monitoring

### Collection and dashboard

Prometheus configuration is in [`monitoring/prometheus.yml`](monitoring/prometheus.yml). It scrapes Gateway and Room at `/metrics` every 5 seconds and UserData at `/actuator/prometheus` every 250 ms (200 ms scrape timeout). Prometheus's UI is at <http://127.0.0.1:9090>; check **Status → Targets** and confirm the `gateway`, `room` and `userdata` targets are `UP`.

Grafana is provisioned with a Prometheus datasource and loads dashboards from `monitoring/dashboards/`. Open **Dashboards → Load Tests → UserData Load Correlation**. Its source file is [`monitoring/dashboards/userdata-load-correlated.json`](monitoring/dashboards/userdata-load-correlated.json). Compose mounts the dashboard read-only, so edit the JSON and restart Grafana to apply file changes:

```bash
docker compose restart grafana
```

The correlated dashboard includes process CPU, BCrypt P50/P95/P99 and mean, UserData gRPC latency, Gateway HTTP latency, database operation latency, exported JVM/Hikari/executor data, request rates, BCrypt concurrency and waiters, peak concurrency, permit wait P95, and rejection rates.

Two dashboard panels depend on metrics that are not uniformly available: the **Process CPU** panel queries Spring's `process_cpu_usage` for Gateway and Room as well as UserData, but the Go Prometheus collector exposes `process_cpu_seconds_total` instead. Therefore the UserData CPU series is the expected one in that panel; use `rate(process_cpu_seconds_total{job="gateway"}[1m])` or the equivalent `room` query for Go process CPU (CPU-seconds per second, where `1` is one fully used core). The **Container CPU throttling** panel expects cAdvisor/cgroup metrics, but the current Prometheus configuration does not scrape cAdvisor, so that panel can be empty while application targets are healthy.

### Metrics and example PromQL

Micrometer timer names are exported to Prometheus with `_seconds` suffixes. UserData's `job="userdata"` label comes from the Prometheus scrape configuration; BCrypt timers also have `operation="encode"` or `operation="verify"`.

Current BCrypt activity by operation:

```promql
password_hash_in_flight{job="userdata",operation="encode"}
password_hash_in_flight{job="userdata",operation="verify"}
```

Total active hashes and requests waiting for a permit:

```promql
password_hash_active{job="userdata"}
password_hash_waiting{job="userdata"}
```

BCrypt execution P95, split by operation:

```promql
histogram_quantile(0.95,
  sum by (le, operation) (
    rate(userdata_password_operation_seconds_bucket{job="userdata"}[1m])
  )
)
```

BCrypt completions per second:

```promql
sum by (operation) (
  rate(userdata_password_operation_seconds_count{job="userdata"}[1m])
)
```

Permit wait P95:

```promql
histogram_quantile(0.95,
  sum by (le) (
    rate(userdata_password_permit_wait_seconds_bucket{job="userdata"}[1m])
  )
)
```

Gateway HTTP request P95 (currently instrumented for login and registration):

```promql
histogram_quantile(0.95,
  sum by (le, operation) (
    rate(service_operation_duration_seconds_bucket{
      service="gateway",stage="http_total"
    }[1m])
  )
)
```

Other useful UserData metrics include:

| Metric | Meaning |
|---|---|
| `password_hash_active` | Number of encode/verify calls currently inside BCrypt |
| `password_hash_waiting` | Number of calls currently waiting for a permit |
| `password_hash_active_max` | Peak total BCrypt execution concurrency since process start |
| `password_hash_in_flight{operation}` | Current BCrypt calls, split into `encode` and `verify` |
| `password_hash_in_flight_max{operation}` | Peak BCrypt concurrency by operation since process start |
| `userdata_password_operation_seconds_*` | Encode/verify execution timer: count, sum and histogram buckets |
| `userdata_password_permit_wait_seconds_*` | Time spent acquiring a BCrypt permit |
| `password_hash_permit_timeout_total` | Calls rejected after waiting for a permit timed out |
| `password_hash_permit_queue_full_total` | Calls rejected because the configured waiting limit was full |
| `userdata_grpc_server_requests_seconds_*` | UserData gRPC request duration |
| `userdata_database_operation_seconds_*` | Timed UserData database operations |

Peak gauges reset when UserData restarts. A short load phase may finish between Prometheus scrapes; use a longer test phase or a sufficiently short scrape interval when measuring peaks. A Prometheus `rate()` over a window is an averaged completion rate, not a single request's latency or instantaneous concurrency.

## HTTP API

Base URL: `http://127.0.0.1:26526`. Send JSON requests with `Content-Type: application/json`.

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/health` | Liveness response (`200 ok`) |
| `POST` | `/login` | Register or login, selected by `type` |
| `POST` | `/login_session` | Validate a session |
| `POST` | `/room/user` | Create a room |
| `POST` | `/room/room` | Get room info, join, leave or start a game |
| `POST` | `/room/userPreparation` | Set or cancel ready state |
| `POST` | `/room/endGame` | Internal game-end callback |

### Register and login

Register:

```json
{"type":"register","accountId":"alice","password":"secret"}
```

Login:

```json
{"type":"login","accountId":"alice","password":"secret"}
```

Successful login returns an opaque session ID:

```json
{"success":true,"sessionId":"<opaque-session-id>"}
```

### Validate session

```json
{"sessionId":"<opaque-session-id>"}
```

A valid session returns `{"success":true}`. A missing or expired session returns `401`; Redis/session storage errors return `503 Session service unavailable`.

### Room routes

Create a room:

```json
{"type":"createRoom","sessionId":"<session-id>"}
```

```json
{"roomId":"12345"}
```

Get members:

```json
{"type":"getRoomInfo","sessionId":"<session-id>","roomId":"12345"}
```

The response currently uses the legacy field `roomId` for the user ID array:

```json
{"success":true,"roomId":["alice","bob"]}
```

Join, leave and start operations use `type` values `joinRoom`, `leaveRoom` and `startGame`, with `sessionId` and `roomId`; successful operations return `{"result":true}`. A room has a maximum size of four. Starting requires the host, at least two players, a waiting room and all members prepared.

Set ready state:

```json
{"type":"setPreparation","sessionId":"<session-id>","roomId":"12345","targetStatus":true}
```

The response is `{"status":true}` or `{"status":false}`.

End-game callback:

```json
{"type":"endGame","roomId":"12345"}
```

This callback has no service authentication and should not be exposed to untrusted clients.

## TCP API

Gateway listens on `127.0.0.1:26527`. Each message is framed as `[4-byte big-endian body length][JSON body]`; body size must be from 1 byte through 64 KiB. Protocol version is `"1.0"`.

Authenticate first with the session ID:

```json
{"version":"1.0","request_id":"auth-001","type":"auth","data":{"token":"<session-id>"}}
```

Heartbeat:

```json
{"version":"1.0","request_id":"hb-001","type":"heartbeat","data":null}
```

Heartbeat renews `user_status:<userId>` to a 30-second TTL. Send it before expiry; every 10 seconds is the current recommendation.

Room chat uses `type: "room"` and a `room_id`; private chat uses `type: "private"` and a recipient user ID. Chat is delivered only to online TCP connections and is not persisted.

TCP error codes include `INVALID_MESSAGE`, `UNAUTHENTICATED`, `AUTHENTICATION_FAILED`, `FORBIDDEN`, `NOT_FOUND`, `VERSION_MISMATCH`, `INTERNAL_ERROR` and `SERVER_CLOSED_ERROR`.

## Timeouts

HTTP routes share `GATEWAY_REQUEST_TIMEOUT` (default `10s`) through `http.TimeoutHandler`. Its request context is passed to downstream gRPC and Redis calls. Expired HTTP requests return `503 Request timed out`.

TCP has separate authentication, idle-read, operation and write deadlines. Override values in the root `.env`; values must be positive Go durations such as `5s` or `1m`.

| Environment variable | Default | Meaning |
|---|---:|---|
| `GATEWAY_REQUEST_TIMEOUT` | `10s` | Total budget for one HTTP request |
| `GATEWAY_TCP_UNAUTHENTICATED_TIMEOUT` | `10s` | Time to authenticate after TCP connect |
| `GATEWAY_TCP_IDLE_TIMEOUT` | `45s` | Maximum idle time between TCP messages |
| `GATEWAY_TCP_OPERATION_TIMEOUT` | `5s` | Deadline for auth, chat, heartbeat and other TCP operations |
| `GATEWAY_TCP_WRITE_TIMEOUT` | `10s` | Deadline for each TCP write |

Example override:

```env
GATEWAY_REQUEST_TIMEOUT=8s
GATEWAY_TCP_UNAUTHENTICATED_TIMEOUT=12s
GATEWAY_TCP_IDLE_TIMEOUT=60s
GATEWAY_TCP_OPERATION_TIMEOUT=4s
GATEWAY_TCP_WRITE_TIMEOUT=5s
```

Recreate Gateway to apply changes:

```bash
docker compose up -d --build gateway
```

## gRPC services

Gateway connects to UserData at `userdata:26529` and Room at `room:26528` inside Compose.

UserData RPCs are `UserRegistration` and `UserLogin`, defined in [`proto/user_info/login.proto`](proto/user_info/login.proto). Room RPCs are `GetRoomInfo`, `CreateRoom`, `JoinRoom`, `LeaveRoom`, `SetPreparation`, `StartGame` and `EndGame`, defined in [`proto/room/room_4_gateway.proto`](proto/room/room_4_gateway.proto).

## Storage

### Redis

| Key | Value | TTL |
|---|---|---:|
| `session:<sessionId>` | `session_account:<accountId>` | 7 days |
| `session_account:<accountId>` | `session:<sessionId>` | 7 days |
| `user_status:<userId>` | User ID and status JSON | Initially 30 seconds; heartbeat renews it |
| `room:<roomId>` | Room JSON | No expiration; deleted when the room becomes empty |

User statuses include `in_lobby`, `in_room` and `in_game`. Room statuses include `waiting`, `in_game` and `finished`.

### MySQL

The `users` table stores an auto-increment ID, unique account ID and BCrypt password hash. JPA currently uses `ddl-auto: update` for development; production should use managed schema migrations. UserData uses `MYSQL_DATABASE`, `MYSQL_USER` and `MYSQL_PASSWORD` from Compose.

## Password hashing limits

UserData routes both BCrypt `encode()` (registration) and `matches()` (login) through one fair semaphore. Compose defaults are:

| Environment variable | Default | Meaning |
|---|---:|---|
| `USERDATA_PASSWORD_HASH_MAX_CONCURRENCY` | `10` | Simultaneous BCrypt executions; `0` disables the execution limit |
| `USERDATA_PASSWORD_HASH_ACQUIRE_TIMEOUT` | `5s` | Maximum time waiting for an execution permit |
| `USERDATA_PASSWORD_HASH_MAX_WAITERS` | `100` | Maximum queued waiters; `0` rejects immediately when no permit is available |

When all permits are occupied and the waiter limit is reached or its timeout expires, the request fails explicitly; it is not silently treated as a bad password or successful registration. These settings are useful for controlled performance experiments and should be evaluated alongside BCrypt latency, permit wait, throughput, CPU and rejection metrics.

## Development and tests

Go services are separate modules:

```bash
(cd gateway && go test ./...)
(cd room && go test ./...)
```

Gateway resilience tests live under `gateway/test/resiliencetest`. Some Room Redis integration tests require a local Redis and an explicit `RESILIENCE_REDIS_ADDR`; see the test source before running them. The `gateway/test/k6/benchmark.js` script requires a k6 build that includes the `k6/x/tcp` extension. It creates persistent test accounts; use a fresh `RUN_ID` for each run and clean only the generated test data afterward.

The heartbeat black-box script is:

```bash
go run scripts/heartbeat_blackbox.go
```

Run black-box and load tests only against an isolated development stack. The K6 script's full default scenario ramps through 10, 25 and 50 VUs; it can be shortened with its environment variables such as `ITERATIONS` and scenario start times. See [`gateway/test/k6/baseline.md`](gateway/test/k6/baseline.md) for recorded run conditions and interpretation notes.

## Security and current limitations

- The default Compose Redis has no password or TLS; HTTP and gRPC have no TLS/service authentication.
- `/room/endGame` has no internal authentication.
- TCP connection state is held in one Gateway process's memory and is not shared across replicas.
- Chat is not persisted.
- The `game` service is disabled and its TCP message is a placeholder.
- Add request size and field validation, rate limiting, session revocation and fine-grained authorization before production use.
- Metrics and dashboards are configured for local development/load testing. cAdvisor/container throttling metrics are not currently scraped.

## License

MIT License: see [LICENSE](LICENSE).
