# Duverge

[English](README.md) | [中文](README_ch.md)

Duverge 是一个类游戏后端项目，采用模块化、面向服务的架构进行设计。

项目目前仍在积极开发中，由多个后端服务组成，这些服务通过 HTTP、TCP 和 gRPC 进行通信。

## 架构

### 概览

整个架构分为四个主要服务：

* Gateway
* UserData
* Room
* Game

本项目采用微服务架构，并使用 Docker Compose 对各个服务进行管理和部署。

项目的主要目的，是学习独立服务之间的通信方式，以及如何处理不同容器之间的通信错误。

### 通信

来自客户端的请求，包括 HTTP 和 TCP 连接，都会首先由 Gateway 接收。

Gateway 会根据连接类型和请求用途对请求进行路由。

后端服务之间使用 gRPC 进行通信。

### 语言选择

Gateway 和 Room 服务使用 **Go** 实现，主要因为 Go 提供了轻量级的 goroutine 和高效的并发支持，因此适合处理大量并发连接。

UserData 服务使用 **Java 和 Spring Boot** 实现，主要因为其对 CRUD 操作、数据库连接以及数据持久化具有成熟的支持，例如 Spring Data JPA。

## 技术栈

* **Go**
* **Java**
* **Spring Boot**
* **gRPC / Protocol Buffers**
* **MySQL**
* **Docker Compose**
* **Maven**
* **Git**

## 项目结构

当前项目结构：

```text
Duverge/
├── docker-compose.yml
├── .env
├── .gitignore
│
├── gateway/
│   ├── cmd/
│   │   └── server/
│   ├── internal/
│   │   ├── client/
│   │   ├── router/
│   │   ├── server/
│   │   └── service/
│   └── proto/
│
├── userData/
│   ├── src/
│   └── pom.xml
│
├── game/
├── room/
├── mysql/
```

## 通信

### Client → Gateway

客户端通过 HTTP 和 TCP 与 Gateway 进行通信。

#### HTTP

Gateway HTTP 服务：

```text
127.0.0.1:26526
```

健康检查示例：

```bash
curl http://127.0.0.1:26526/health
```

预期返回：

```text
ok
```

### Gateway → UserData

Gateway 使用 gRPC 与 UserData 服务进行通信。

在 Docker Compose 网络中：

```text
userdata:9090
```

gRPC API 使用 Protocol Buffers 进行定义。

目前支持的用户操作包括：

* 用户注册
* 用户登录

示例服务定义：

```protobuf
service UsersLoginService {
    rpc UserRegistration(UserRegistrationRequest)
        returns (UserRegistrationResponse);

    rpc UserLogin(UserLoginRequest)
        returns (UserLoginResponse);
}
```

### UserData → MySQL

UserData 使用 Spring Data JPA 与 MySQL 进行通信。

在 Docker Compose 网络中：

```text
mysql:3306
```

默认数据库：

```text
user_data
```

## 配置

敏感配置信息应该存储在 `.env` 文件中，而不是直接提交到 Git。

`.env` 示例：

```env
MYSQL_ROOT_PASSWORD=your_root_password
MYSQL_DATABASE=user_data
MYSQL_USER=your_user
MYSQL_PASSWORD=your_password
```

`.env` 文件应该被 Git 忽略：

```gitignore
.env
```

Docker Compose 可以通过以下方式引用这些环境变量：

```yaml
environment:
  MYSQL_USER: ${MYSQL_USER}
  MYSQL_PASSWORD: ${MYSQL_PASSWORD}
```

UserData 也可以通过 Docker Compose 获取这些环境变量，并在 `application.yml` 中使用。

## 环境要求

运行项目之前，请安装：

* Docker
* Docker Compose
* Git

如果需要在 Docker 之外进行本地开发，还建议安装：

* Go
* JDK 21
* Maven

## 运行项目

克隆仓库：

```bash
git clone <repository-url>
cd Duverge
```

在与 `docker-compose.yml` 相同的目录下创建 `.env` 文件。

然后启动完整项目：

```bash
docker compose up -d
```

如果需要在启动前重新构建所有服务：

```bash
docker compose up -d --build
```

查看正在运行的容器：

```bash
docker compose ps
```

查看所有服务的日志：

```bash
docker compose logs
```

查看指定服务的日志：

```bash
docker compose logs gateway
docker compose logs userdata
docker compose logs mysql
```

停止项目：

```bash
docker compose down
```

## 数据库

项目使用 MySQL。

MySQL 数据库通过 Docker Compose 进行配置。

UserData 在开发环境中使用 JPA 管理数据库 Schema。

开发环境配置：

```yaml
spring:
  jpa:
    hibernate:
      ddl-auto: update
```

## API 示例

### 用户注册

请求：

```text
POST /login
```

请求内容：

```json
{
  "type": "register",
  "accountId": "123456",
  "password": "abc123"
}
```

示例：

```bash
curl -X POST http://127.0.0.1:26526/login \
  -H "Content-Type: application/json" \
  -d '{
    "type": "register",
    "accountId": "123456",
    "password": "abc123"
  }'
```

注册成功后返回：

```text
Registration successful
```

### 用户登录

请求内容：

```json
{
  "type": "login",
  "accountId": "123456",
  "password": "abc123"
}
```

示例：

```bash
curl -X POST http://127.0.0.1:26526/login \
  -H "Content-Type: application/json" \
  -d '{
    "type": "login",
    "accountId": "123456",
    "password": "abc123"
  }'
```

## 测试

项目在开发过程中使用自动化测试。

目前仅使用简单黑盒测试，预计正式跑通HTTP链路后会添加

* pytest
* 基于 Docker 的集成测试

随着项目开发的推进，测试基础设施将进一步完善。

## 安全

本项目目前主要用于开发和学习目的。

如果用于生产环境，还需要进一步实现以下安全措施：

* 对外部 HTTP 通信使用 HTTPS
* 使用密码哈希，而不是以明文形式存储密码
* 安全的 Session 管理
* Authentication 和 Authorization
* 在需要时为 gRPC 使用 TLS
* 安全管理数据库凭据
* 针对不同环境进行配置管理
* 数据库迁移管理

## 项目状态

Duverge 目前仍处于积极开发阶段。

### 当前进展

目前已完成以下功能：

* **用户注册与登录链路**：Gateway → gRPC → UserData → MySQL → 返回结果
* **健康检查（Health Check）**：用于检查服务运行状态
* **Docker Compose 部署**：使用 Docker Compose 管理和部署服务
* **MySQL 数据持久化**：实现用户数据的持久化存储

### 后续开发规划

后续计划实现以下功能：

* **用户 Session 管理**：基于凭证的 Session 机制，支持过期时间及独立的 Session 管理
* **UserData 服务扩展**：增加更多用户相关数据及功能
* **游戏房间系统**：实现房间创建、管理以及玩家管理等功能
* **Gateway TCP 服务器**：支持持久化 TCP 连接
* **游戏服务器开发**：实现游戏核心逻辑及服务端游戏机制

项目的 API 和内部架构可能会随着开发进度发生变化。

## License

本项目使用 [MIT License](LICENSE)。
