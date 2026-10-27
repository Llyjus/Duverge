# Duverge

[English](README.md) | 中文

Duverge 是一个面向多人游戏场景的后端实验项目。项目采用 Gateway、Room、UserData 分层服务，使用 HTTP、TCP、gRPC、Redis 和 MySQL 实现用户认证、会话管理、房间管理、准备状态、游戏状态和聊天，并通过 Prometheus 与 Grafana 监测服务和压测指标。

> 当前项目仍在开发中。本文档描述的是仓库当前代码中的实际接口和数据结构，不代表已经稳定的生产协议。接口、字段和 Redis key 可能继续变化。

## 目录

- [架构与服务](#架构与服务)
- [运行环境与启动](#运行环境与启动)
- [监控系统](#监控系统)
- [HTTP API](#http-api)
- [TCP 协议](#tcp-协议)
- [Gateway 超时设置](#gateway-超时设置)
- [密码哈希并发限制](#密码哈希并发限制)
- [内部 gRPC API](#内部-grpc-api)
- [Redis 数据字典](#redis-数据字典)
- [MySQL 数据字典](#mysql-数据字典)
- [MySQL 连接池与查询超时](#mysql-连接池与查询超时)
- [数据库操作准入限制](#数据库操作准入限制)
- [状态机与业务约束](#状态机与业务约束)
- [测试与验证](#测试与验证)
- [安全与已知限制](#安全与已知限制)

## 架构与服务

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
             (Java)           │
                  │             │
                MySQL ◄──── Redis :6379
```

| 服务 | 语言 | 容器/端口 | 职责 |
|---|---|---|---|
| `gateway` | Go 1.26 | `26526` HTTP、`26527` TCP、`9091` metrics（仅内部） | 对外入口、会话、TCP 认证、heartbeat、聊天、HTTP 房间 API |
| `userdata` | Java 25 / Spring Boot | `26529` gRPC、`127.0.0.1:9091` metrics | 用户注册、BCrypt 密码校验、生成登录 session |
| `room` | Go 1.26 | `26528` gRPC、`9091` metrics（仅内部） | 房间、成员、准备状态、开始/结束游戏 |
| `redis` | Redis | `6379` | session、用户状态、房间状态 |
| `mysql` | MySQL 8.4 | `3306` | UserData 用户持久化 |
| `prometheus` | Prometheus | `127.0.0.1:9090` | 抓取并保存各服务指标 |
| `grafana` | Grafana | `127.0.0.1:3000` | 查询 Prometheus 并展示仪表盘 |
| `adminer` | Adminer | `8080` | 开发环境数据库管理界面 |
| `redisinsight` | RedisInsight | `5540` | 开发环境 Redis 管理界面 |
| `game` | 未启用 | - | Compose 中暂时注释，TCP `game` 消息目前没有业务实现 |

Gateway 使用 Docker Compose 网络中的服务名访问依赖：

```text
redis:6379
userdata:26529
room:26528
mysql:3306
```

## 项目结构

```text
Duverge/
├── docker-compose.yml
├── .env                         # 本地环境变量，不应提交
├── gateway/                     # 对外 Gateway，Go
│   ├── cmd/server/main.go
│   ├── internal/router/         # HTTP 路由
│   ├── internal/server/         # HTTP/TCP 服务
│   ├── internal/service/html/   # HTTP handlers
│   ├── internal/service/tcp/    # TCP 协议、认证、heartbeat、聊天
│   ├── proto/                   # Gateway 生成的 gRPC 代码
│   └── test/                    # Gateway 韧性测试和 K6 脚本
├── room/                        # Room gRPC 服务，Go
├── userData/userdata/           # UserData gRPC 服务，Java
├── proto/                       # gRPC/Protobuf 源文件
├── mysql/data/                  # 开发环境 MySQL 持久化目录
├── monitoring/                  # Prometheus 配置、Grafana provisioning 和 dashboard
├── scripts/                     # 黑盒/辅助脚本
└── runtime-validation-report.md # 最近一次运行验证记录（如存在）
```

## 运行环境与启动

### 要求

- Docker
- Docker Compose
- Git
- 本地开发时可选：Go 1.26、JDK 25、Maven

### 配置 `.env`

在 `docker-compose.yml` 同级目录创建 `.env`：

```env
MYSQL_ROOT_PASSWORD=root
MYSQL_DATABASE=user_data
MYSQL_USER=rin
MYSQL_PASSWORD=change-me
```

不要在源码或提交中保存真实密码。`MYSQL_DATABASE`、`MYSQL_USER` 和 `MYSQL_PASSWORD` 会传给 `userdata` 与 `mysql`。

### 启动、查看和停止

```bash
# 首次启动或代码/Dockerfile 改动后
docker compose up -d --build

# 查看状态
docker compose ps

# 查看日志
docker compose logs -f gateway
docker compose logs -f room
docker compose logs -f userdata

# 停止容器但保留 MySQL 数据卷目录
docker compose down
```

健康检查：

```bash
curl -i http://127.0.0.1:26526/health
# HTTP/1.1 200 OK
# ok
```

清空开发环境 Redis（会删除当前 Redis 数据库中的全部 key，谨慎执行）：

```bash
docker compose exec redis redis-cli FLUSHALL
```

## 监控系统

`docker compose up -d --build` 会同时启动 Prometheus 和 Grafana。Prometheus 配置位于 [`monitoring/prometheus.yml`](monitoring/prometheus.yml)，Grafana 会自动配置 Prometheus 数据源并加载 [`monitoring/dashboards/userdata-load-correlated.json`](monitoring/dashboards/userdata-load-correlated.json)。

| 页面/服务 | 地址 | 用途 |
|---|---|---|
| Prometheus | <http://127.0.0.1:9090> | PromQL 查询；在 **Status → Targets** 检查抓取目标是否为 `UP` |
| Grafana | <http://127.0.0.1:3000> | **Dashboards → Load Tests → UserData Load Correlation** |
| UserData 指标 | <http://127.0.0.1:9091/actuator/prometheus> | 本机直接查看 Spring/Micrometer 指标 |

Grafana 初始账号密码为镜像默认值 `admin` / `admin`（若已更改则使用现有凭据）；首次进入后请修改默认密码。Gateway 和 Room 的 `9091` 指标端口只在 Compose 网络内提供，Prometheus 可以访问，但没有映射到宿主机。

Prometheus 每 5 秒抓取 Gateway 和 Room 的 `/metrics`；UserData 的采集间隔是 250 ms，抓取超时是 200 ms。仪表盘可关联查看进程 CPU、BCrypt 耗时/并发/等待、UserData gRPC 与数据库耗时、Gateway HTTP 耗时、请求速率及 JVM/连接池指标。

常用 PromQL：

```promql
# BCrypt 当前执行数，按 encode / verify 区分
password_hash_in_flight{job="userdata",operation="encode"}
password_hash_in_flight{job="userdata",operation="verify"}

# BCrypt 总执行数及等待许可数
password_hash_active{job="userdata"}
password_hash_waiting{job="userdata"}

# BCrypt 执行耗时 P95
histogram_quantile(0.95,
  sum by (le, operation) (
    rate(userdata_password_operation_seconds_bucket{job="userdata"}[1m])
  )
)

# BCrypt 每秒完成次数
sum by (operation) (
  rate(userdata_password_operation_seconds_count{job="userdata"}[1m])
)

# 获取 BCrypt 执行许可的等待时间 P95
histogram_quantile(0.95,
  sum by (le) (
    rate(userdata_password_permit_wait_seconds_bucket{job="userdata"}[1m])
  )
)
```

主要 UserData 指标：

| 指标 | 含义 |
|---|---|
| `password_hash_active` | 正在执行的 BCrypt 任务总数 |
| `password_hash_waiting` | 当前等待执行许可的请求数 |
| `password_hash_active_max` | 进程启动以来观测到的 BCrypt 执行并发峰值 |
| `password_hash_in_flight{operation}` | 当前正在执行的哈希数，`operation` 为 `encode` 或 `verify` |
| `password_hash_in_flight_max{operation}` | 按操作区分的进程生命周期并发峰值 |
| `userdata_password_operation_seconds_*` | encode/verify 执行耗时、次数和直方图桶 |
| `userdata_password_permit_wait_seconds_*` | 获取 BCrypt 执行许可的等待耗时 |
| `password_hash_permit_timeout_total` | 等待许可超时而拒绝的次数 |
| `password_hash_permit_queue_full_total` | 等待队列达到上限而拒绝的次数 |
| `userdata_grpc_server_requests_seconds_*` | UserData gRPC 请求耗时 |
| `userdata_database_operation_seconds_*` | UserData 数据库操作耗时 |
| `hikaricp_connections_active` | 当前从连接池借出、正在使用的连接数 |
| `hikaricp_connections_idle` | 已建立但空闲、可供借用的连接数 |
| `hikaricp_connections_max` | 连接池配置的最大连接数 |
| `hikaricp_connections_pending` | 当前等待从连接池获取连接的调用数 |
| `userdata_database_active` | 已准入且当前正在执行的数据库操作数 |
| `userdata_database_waiting` | 当前等待数据库操作准入许可的调用数 |
| `userdata_database_active_max` | 进程启动以来观测到的数据库操作执行并发峰值 |
| `userdata_database_permit_wait_seconds_*` | 等待数据库操作准入许可的耗时 |
| `userdata_database_permit_timeout_total` | 等待准入许可超时而拒绝的请求数 |
| `userdata_database_permit_queue_full_total` | 有界等待队列已满而拒绝的请求数 |

并发峰值 Gauge 在 UserData 重启时重置。短暂的压测阶段可能刚好发生在两次采集之间，导致瞬时峰值没有被抓到；可延长测试阶段或缩短采集间隔。Prometheus `rate()` 是窗口平均速率，并非瞬时并发数或单次请求耗时。

注意：仪表盘的 **Container CPU throttling** 面板依赖 cAdvisor/cgroup 指标，但当前 Prometheus 没有配置 cAdvisor target，因此该面板可能为空。**Process CPU** 面板使用 Spring 的 `process_cpu_usage` 查询 UserData；Gateway/Room 的 Go collector 提供的是 `process_cpu_seconds_total`，所以它们在当前面板查询中可能没有数据。Go 进程 CPU 可用以下查询查看，结果单位是 CPU-seconds/second，`1` 相当于一个 CPU 核满载：

```promql
rate(process_cpu_seconds_total{job="gateway"}[1m])
rate(process_cpu_seconds_total{job="room"}[1m])
```

## HTTP API

Gateway HTTP 地址：

```text
http://127.0.0.1:26526
```

所有 JSON 请求都应使用：

```http
Content-Type: application/json
```

### `GET /health`

服务存活检查，不依赖认证。

响应：

```text
ok
```

状态码：`200`

### `POST /login`

同一路由通过 `type` 区分注册和登录。

#### 注册

请求：

```json
{
  "type": "register",
  "accountId": "alice",
  "password": "secret"
}
```

成功响应 `200`：

```json
{
  "success": true
}
```

错误：

| 状态码 | 条件 |
|---|---|
| `400` | JSON 非法或 `type` 不支持 |
| `401` | `accountId` 已存在 |
| `503` | 请求超时（`Request timed out`）或 UserData gRPC 服务不可用（`UserData service unavailable`） |
| `500` | 其他 UserData、Redis 或数据库错误 |

密码由 UserData 使用 BCrypt 保存，Gateway 不直接访问 `users` 表。

#### 登录

请求：

```json
{
  "type": "login",
  "accountId": "alice",
  "password": "secret"
}
```

成功响应 `200`：

```json
{
  "success": true,
  "sessionId": "<opaque-session-id>"
}
```

错误：

| 状态码 | 条件 |
|---|---|
| `400` | JSON 非法或 `type` 不支持 |
| `401` | 用户名或密码错误 |
| `503` | 请求超时（`Request timed out`）或 UserData gRPC 服务不可用（`UserData service unavailable`） |
| `500` | 其他 UserData、Redis 或 session 创建错误 |

登录成功后，Gateway 会在 Redis 创建两个 7 天 TTL 的 session key，并创建一个 30 秒 TTL 的用户状态 key（详见 [Redis 数据字典](#redis-数据字典)）。

### `POST /login_session`

校验 Gateway 创建的 session。

请求：

```json
{
  "sessionId": "<opaque-session-id>"
}
```

成功响应 `200`：

```json
{
  "success": true
}
```

错误：

| 状态码 | 条件 |
|---|---|
| `400` | JSON 非法 |
| `401` | session 不存在、为空或已过期 |
| `503` | Redis/session 存储不可用，服务暂时无法校验 session |

### `POST /room/user`

当前支持的 `type`：

#### 创建房间：`createRoom`

请求：

```json
{
  "type": "createRoom",
  "sessionId": "<session-id>"
}
```

成功响应 `200`：

```json
{
  "roomId": "12345"
}
```

创建者自动成为房主并加入房间，初始房间状态为 `waiting`，最大人数为 `4`。

#### 错误

| 状态码 | 条件 |
|---|---|
| `400` | JSON 非法或 type 不支持 |
| `500` | session 查询或 Room gRPC 调用失败 |

### `POST /room/room`

请求必须包含：

```json
{
  "type": "<operation>",
  "sessionId": "<session-id>",
  "roomId": "<room-id>"
}
```

#### 查询房间：`getRoomInfo`

响应 `200`：

```json
{
  "success": true,
  "roomId": ["alice", "bob"]
}
```

> 当前响应字段名 `roomId` 实际承载用户 ID 数组，这是现有代码行为；后续建议改成 `userIds`。

#### 加入房间：`joinRoom`

响应 `200`：

```json
{
  "result": true
}
```

约束：用户必须处于 `in_lobby`，房间必须处于 `waiting`，且房间未满。

#### 离开房间：`leaveRoom`

响应 `200`：

```json
{
  "result": true
}
```

用户状态变为 `in_lobby`；房间空后会删除 `room:<roomId>`。

#### 开始游戏：`startGame`

响应 `200`：

```json
{
  "result": true
}
```

约束：请求者必须是房主、房间处于 `waiting`、所有成员已准备、至少有 2 名玩家。成功后房间变为 `in_game`，所有成员变为 `in_game`。

#### 错误

| 状态码 | 条件 |
|---|---|
| `400` | JSON 非法或 type 不支持 |
| `500` | session 查询、Redis 或 Room gRPC 错误 |

当前 HTTP handler 对 Room 业务错误多数统一返回 `500`，未细分为 `404/409/403`。

### `POST /room/userPreparation`

设置当前用户的准备状态。

请求：

```json
{
  "type": "setPreparation",
  "sessionId": "<session-id>",
  "roomId": "12345",
  "targetStatus": true
}
```

响应：

```json
{
  "status": true
}
```

约束：用户必须在房间内，房间必须处于 `waiting`。重复设置 `true` 是幂等的；设置 `false` 会取消准备。

### `POST /room/endGame`

该接口设计为由 Game 服务调用，不是普通用户接口。

请求：

```json
{
  "type": "endGame",
  "roomId": "12345"
}
```

响应：

```json
{
  "result": true
}
```

结束游戏后房间回到 `waiting`，准备列表清空，成员状态回到 `in_room`。

当前接口没有额外的内部服务认证，不能直接暴露到不可信网络。

## TCP 协议

Gateway TCP 地址：

```text
127.0.0.1:26527
```

### 帧格式

每条消息使用长度前缀帧：

```text
[4 字节 Big-Endian body length][JSON body]
```

- body 长度必须大于 `0`
- body 最大 `64 KiB`
- JSON 协议版本为 `"1.0"`
- `request_id` 可选，用于客户端匹配 response

基础消息：

```json
{
  "version": "1.0",
  "request_id": "req-001",
  "type": "heartbeat",
  "data": null
}
```

### 认证：`auth`

必须是连接建立后的第一类业务消息：

```json
{
  "version": "1.0",
  "request_id": "auth-001",
  "type": "auth",
  "data": {
    "token": "<session-id>"
  }
}
```

成功响应：

```json
{
  "version": "1.0",
  "type": "response",
  "success": true
}
```

认证成功后 Gateway 会：

1. 校验 `session:<sessionId>`；
2. 将当前 TCP 连接注册到内存连接管理器；
3. 创建/覆盖 `user_status:<userId>`；
4. 允许 heartbeat、chat、game 消息。

### Heartbeat：`heartbeat`

请求：

```json
{
  "version": "1.0",
  "request_id": "hb-001",
  "type": "heartbeat",
  "data": null
}
```

成功响应：

```json
{
  "version": "1.0",
  "type": "response",
  "success": true
}
```

Heartbeat 只执行 Redis `EXPIRE user_status:<userId> 30`，不会改变用户状态字段。客户端应至少在 30 秒 TTL 到期前发送一次，推荐每 10 秒发送一次。

### Chat：`chat`

#### 房间聊天

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

Gateway 会读取 `room:<roomId>` 的 `userIds`，向当前在线成员发送：

```json
{
  "version": "1.0",
  "type": "chat",
  "success": true,
  "data": {
    "message_id": "<uuid>",
    "sender_id": "alice",
    "type": "text",
    "content": "hello room"
  }
}
```

发送者随后收到对应 request 的成功 response。

#### 私聊

```json
{
  "version": "1.0",
  "request_id": "chat-002",
  "type": "chat",
  "data": {
    "type": "private",
    "data": {
      "to": "bob",
      "type": "text",
      "content": "hello"
    }
  }
}
```

目标用户必须存在于 Gateway 的在线连接管理器中。当前聊天消息没有持久化到 MySQL 或 Redis。

### `game`

`game` 消息类型已经保留，但当前 Gateway 只记录日志，没有业务响应或游戏逻辑。

### TCP 错误码

错误 response 的 `data`：

```json
{
  "error_code": "INVALID_MESSAGE",
  "detail": "非法消息格式"
}
```

当前错误码：

| error_code | 含义 | 常见触发条件 |
|---|---|---|
| `INVALID_MESSAGE` | 消息格式错误 | 非法 JSON、未知 type、chat 字段缺失 |
| `UNAUTHENTICATED` | 未认证 | 未认证连接发送非 auth |
| `AUTHENTICATION_FAILED` | 认证失败 | session 不存在、过期或没有用户 |
| `FORBIDDEN` | 禁止操作 | 已认证连接重复 auth |
| `NOT_FOUND` | 资源不存在 | 房间不存在、私聊目标不在线 |
| `VERSION_MISMATCH` | 协议版本不支持 | version 不是 `1.0` |
| `INTERNAL_ERROR` | 内部错误 | Redis、网络或未分类服务错误 |
| `SERVER_CLOSED_ERROR` | 连接已关闭 | EOF/UnexpectedEOF |

认证失败和版本错误返回错误后会关闭 TCP 连接；普通业务错误通常保持连接。

## Gateway 超时设置

Gateway 的 HTTP 路由通过 `http.TimeoutHandler` 使用统一请求超时。该 deadline 会随请求 context 传入 UserData/Room gRPC 调用和 Redis 操作；请求超过预算时返回 HTTP `503`、`Request timed out`。Redis/session 存储自身不可用时，`POST /login_session` 返回 HTTP `503`、`Session service unavailable`；无效或过期 session 仍返回 HTTP `401`。

TCP 连接分别设置认证前等待、空闲读取、单次操作和写入超时。超时配置必须是正数 Go duration（例如 `3s`、`1m`）；无效配置会阻止 Gateway 启动。

| 环境变量 | 默认值 | 作用 |
|---|---:|---|
| `GATEWAY_REQUEST_TIMEOUT` | `10s` | 每个 HTTP 请求的总预算 |
| `GATEWAY_TCP_UNAUTHENTICATED_TIMEOUT` | `10s` | TCP 建连后完成认证的最长时间 |
| `GATEWAY_TCP_IDLE_TIMEOUT` | `45s` | 两次 TCP 消息读取之间允许的最长空闲时间 |
| `GATEWAY_TCP_OPERATION_TIMEOUT` | `5s` | TCP auth、chat、heartbeat 等操作的 context deadline |
| `GATEWAY_TCP_WRITE_TIMEOUT` | `10s` | 单次 TCP 写入的 deadline |

Docker Compose 默认使用上述值。可在项目根目录 `.env` 覆盖，例如：

```env
GATEWAY_REQUEST_TIMEOUT=8s
GATEWAY_TCP_UNAUTHENTICATED_TIMEOUT=12s
GATEWAY_TCP_IDLE_TIMEOUT=60s
GATEWAY_TCP_OPERATION_TIMEOUT=4s
GATEWAY_TCP_WRITE_TIMEOUT=5s
```

修改后重建 Gateway：

```bash
docker compose up -d --build gateway
```

## 密码哈希并发限制

UserData 的注册 `PasswordEncoder.encode()` 和登录 `PasswordEncoder.matches()` 共用一个公平 Semaphore。Compose 默认配置如下：

| 环境变量 | 默认值 | 作用 |
|---|---:|---|
| `USERDATA_PASSWORD_HASH_MAX_CONCURRENCY` | `10` | 同时执行 BCrypt 的最大数量；设为 `0` 可关闭执行并发上限 |
| `USERDATA_PASSWORD_HASH_ACQUIRE_TIMEOUT` | `5s` | 等待执行许可的最长时间 |
| `USERDATA_PASSWORD_HASH_MAX_WAITERS` | `100` | 等待队列最大长度；设为 `0` 表示无可用许可时立即拒绝 |

修改根目录 `.env` 后重建 UserData：

```env
USERDATA_PASSWORD_HASH_MAX_CONCURRENCY=5
USERDATA_PASSWORD_HASH_ACQUIRE_TIMEOUT=5s
USERDATA_PASSWORD_HASH_MAX_WAITERS=100
```

```bash
docker compose up -d --build userdata
```

达到等待队列上限或获取许可超时会返回明确错误，不会被当成密码错误或注册成功。对照压测时应同时比较 BCrypt 执行 P95、等待 P95、端到端 P95、完成速率、CPU 和拒绝次数。

## 内部 gRPC API

### UserData：`UserDataService`

Proto：[`proto/user_info/login.proto`](proto/user_info/login.proto)

服务地址：`userdata:26529`

| RPC | 请求 | 响应 | 作用 |
|---|---|---|---|
| `UserRegistration` | `account_id`, `password` | `result` | 创建用户，账号重复时 `result=false` |
| `UserLogin` | `account_id`, `password` | `result`, `session_id` | BCrypt 校验密码；成功生成 32 字节随机 Base64URL session |

UserData 负责用户凭据和密码哈希；Gateway 负责将 UserData 返回的 session 写入 Redis。

### Room：`RoomService`

Proto：[`proto/room/room_4_gateway.proto`](proto/room/room_4_gateway.proto)

服务地址：`room:26528`

| RPC | 作用 |
|---|---|
| `GetRoomInfo` | 获取房间成员 ID |
| `CreateRoom` | 创建 5 位房间号，房主自动加入 |
| `JoinRoom` | 加入 waiting 且未满的房间 |
| `LeaveRoom` | 离开房间，空房间删除 |
| `SetPreparation` | 设置/取消准备 |
| `StartGame` | 房主在所有成员准备后开始游戏 |
| `EndGame` | 结束游戏并恢复 waiting 状态 |

## Redis 数据字典

Redis 默认无密码，开发环境地址为 `redis:6379`。下面的 key 是当前代码实际使用的格式。

### Session 双向索引

#### `session:<sessionId>`

类型：String
TTL：`7 days`
值：

```text
session_account:<accountId>
```

用途：由 TCP auth、`/login_session` 和 Gateway 房间 HTTP API 反查账号。

#### `session_account:<accountId>`

类型：String
TTL：`7 days`
值：

```text
session:<sessionId>
```

用途：保证同一账号只有一个当前 session。重新登录时 Gateway 会删除旧的双向索引。

### 用户在线/房间状态

#### `user_status:<userId>`

类型：String
首次创建 TTL：`30 seconds`
Heartbeat：只刷新 TTL 到 `30 seconds`
Room/Game 服务更新时：TTL 通常为 `0`（永久，直到业务显式覆盖或删除）

当前 JSON 值：

```json
{
  "user_id": "alice",
  "status": "in_lobby"
}
```

允许的状态：

| 状态 | 含义 |
|---|---|
| `in_lobby` | 已登录，在大厅 |
| `in_room` | 已加入房间 |
| `in_game` | 游戏进行中 |

注意：Gateway 的 heartbeat 不更新 `status`，只续期；Room 的加入、离开、开始游戏和结束游戏负责更新状态 JSON。

### 房间

#### `room:<roomId>`

类型：String
TTL：`0`（无过期时间，除非房间为空时删除）
值：

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

字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `roomId` | string | 5 位房间号 |
| `maximumSize` | integer | 当前固定为 `4` |
| `userIds` | string[] | 房间成员 |
| `hosterId` | string | 房主用户 ID |
| `prepared` | string[] | 已准备成员 |
| `status` | string | `waiting`、`in_game`、`finished` |

Room 使用 Redis `WATCH`/事务更新 `room:<roomId>` 和相关 `user_status:<userId>`，避免并发加入、离开和准备操作覆盖彼此的更新。

### Redis key 总览

| Pattern | 类型 | TTL | 写入者 | 读取者 |
|---|---|---:|---|---|
| `session:<sessionId>` | String | 7 天 | Gateway | Gateway |
| `session_account:<accountId>` | String | 7 天 | Gateway | Gateway |
| `user_status:<userId>` | JSON String | 创建时 30 秒；Room 更新后通常永久 | Gateway/Room | Gateway/Room |
| `room:<roomId>` | JSON String | 永久 | Room | Gateway/Room |

## MySQL 数据字典

数据库由 `.env` 中的 `MYSQL_DATABASE` 指定，默认是 `user_data`。UserData 使用 JPA `ddl-auto: update`，开发环境会自动更新表结构。

### `users`

实体：`com.duverge.userdata.entity.User`
表名：`users`

| 列 | 类型/属性 | 说明 |
|---|---|---|
| `id` | BIGINT，自增，主键 | JPA `IDENTITY` 主键 |
| `account_id` | VARCHAR（JPA 默认命名策略） | 用户登录账号，代码通过 `findByAccountId` 查询 |
| `password` | VARCHAR | BCrypt 哈希，不保存明文密码 |

当前 UserData 只通过 gRPC 暴露注册和登录，不提供用户查询、修改密码或删除用户的公开 API。

## MySQL 连接池与查询超时

UserData 使用 HikariCP。默认最多建立 10 个数据库连接；连接池无空闲连接时，借用连接最多等待 3 秒。MySQL TCP 建连超时为 3 秒，socket I/O 超时为 5 秒；Hibernate/JPA 默认查询超时设为 3 秒（配置单位为毫秒）。

可在根目录 `.env` 覆盖：

| 环境变量 | 默认值 | 含义 |
|---|---:|---|
| `MYSQL_POOL_MAX_SIZE` | `10` | Hikari 最大连接数 |
| `MYSQL_POOL_CONNECTION_TIMEOUT_MS` | `3000` | 等待获取连接的上限，毫秒 |
| `MYSQL_POOL_VALIDATION_TIMEOUT_MS` | `2000` | 校验连接的最长时间；必须小于获取连接超时 |
| `MYSQL_QUERY_TIMEOUT_MS` | `3000` | JPA 查询超时提示，毫秒 |
| `MYSQL_CONNECT_TIMEOUT_MS` | `3000` | MySQL TCP 建连超时，毫秒 |
| `MYSQL_SOCKET_TIMEOUT_MS` | `5000` | MySQL socket I/O 超时，毫秒 |

修改后重建 UserData：

```bash
docker compose up -d --build userdata
```

连接池上限限制实际数据库连接数，连接获取超时限制单个调用等待连接的时长。UserData 还会在 repository 操作前执行有界准入：默认最多 10 个操作同时执行、最多 10 个调用等待，等待上限为 1 秒。请让执行上限不大于 Hikari 最大连接数，避免准入后的操作继续堆积在连接池中。JPA 查询超时属于 JDBC/provider 提示；socket timeout 限制网络 I/O，并不保证 MySQL 服务端立即停止正在执行的 SQL。应同时观察 Hikari active/idle/pending 和数据库准入指标，并在隔离数据库中验证慢查询行为。

## 数据库操作准入限制

`UserService` 中所有 repository 操作共用一个公平信号量。默认最多 10 个操作同时执行，最多 10 个额外调用等待，每个等待调用最多等 1 秒。等待队列满或等待超时后，gRPC 返回 `RESOURCE_EXHAUSTED`；不会把过载伪装成用户不存在或注册成功。等待者使用现有 gRPC 请求线程，不会额外创建线程池。

可在根目录 `.env` 调整并重建 UserData：

| 环境变量 | 默认值 | 含义 |
|---|---:|---|
| `USERDATA_DATABASE_MAX_CONCURRENCY` | `10` | 同时执行的 repository 操作上限；建议与 `MYSQL_POOL_MAX_SIZE` 对齐 |
| `USERDATA_DATABASE_MAX_WAITERS` | `10` | 等待准入许可的调用上限；设为 `0` 时无空闲许可就立即拒绝 |
| `USERDATA_DATABASE_ACQUIRE_TIMEOUT` | `1s` | 等待准入许可的最长时间 |

Grafana 的 **Database admission and bounded waiters** 面板展示正在执行数、等待数、历史执行峰值及拒绝速率。Hikari pending 指标仍可用于观察此准入入口之外的连接池竞争。

## 状态机与业务约束

### 用户状态

```text
登录成功
   │
   ▼
in_lobby ── joinRoom ──► in_room ── startGame ──► in_game
   ▲                         │                      │
   └──────── leaveRoom ◄─────┘                      │
                                                     │
                         endGame ◄───────────────────┘
                              ▼
                          in_room
```

### 房间状态

```text
创建 ──► waiting ── startGame ──► in_game
  ▲          │                         │
  │          └── leave/准备操作         │
  └──────────── endGame ◄──────────────┘
```

开始游戏的必要条件：

1. 请求者是 `hosterId`；
2. 房间状态为 `waiting`；
3. 所有成员都在 `prepared`；
4. 成员数量至少为 `2`；
5. 成员数量不能超过 `maximumSize`（当前为 `4`）。

## 测试与验证

Gateway 和 Room 是独立 Go module：

```bash
cd gateway && go test ./...
cd ../room && go test ./...
```

Gateway 的 deadline、下游不可用、调用重试次数和 Redis 断连语义测试在 `gateway/test/resiliencetest`，会由 `go test ./...` 一并运行。Room 的 Redis 集成测试需要显式设置 `RESILIENCE_REDIS_ADDR` 并连接测试 Redis；运行前请阅读对应测试文件，确保测试 key 使用隔离环境。

仓库里的 K6 生命周期脚本位于 `gateway/test/k6/benchmark.js`，需要包含 `k6/x/tcp` 扩展的 k6 构建。脚本会创建并持久化测试账号，每轮使用新的 `RUN_ID`，不要直接在生产数据上运行。默认场景按 10、25、50 VU 分阶段启动；压测记录和结果解释见 [`gateway/test/k6/baseline.md`](gateway/test/k6/baseline.md)。

Heartbeat 黑盒脚本：

```bash
go run scripts/heartbeat_blackbox.go
```

该脚本覆盖登录、TCP auth、heartbeat、房间和聊天链路。运行前请确认服务已启动；测试账号和 Redis 数据应使用隔离的开发环境。

## 安全与已知限制

当前项目适合本地开发和学习，不应直接作为生产服务部署：

- Compose 中 Redis 没有密码和 TLS；
- HTTP 没有 HTTPS；
- gRPC 没有 TLS 或服务间认证；
- `/room/endGame` 没有内部调用认证；
- TCP 连接管理器是单实例内存状态，Gateway 多副本时不会共享在线连接；
- 聊天消息目前不持久化；
- Room key 永久保存，只有空房间会删除；
- UserData 使用 `ddl-auto: update`，生产环境应使用正式 migration；
- 应增加请求体大小、字段长度、账号格式和速率限制；
- 应考虑 session 撤销、重放防护和更细粒度的授权；
- `game` 服务尚未启用，TCP `game` 消息目前只是占位；
- 当前若客户端正常关闭 TCP，服务端会记录 EOF 日志，这是连接生命周期日志，不等同于业务失败。

## License

本项目使用 [MIT License](LICENSE)。
