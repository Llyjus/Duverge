# Duverge

[English](README.md) | [中文](README_ch.md)

Duverge is a game-like backend project designed with a modular, service-oriented architecture.

The project is currently under active development and consists of several backend services communicating through HTTP, TCP, and gRPC.

## Architecture

### Overview

The architecture is divided into four main services:

* Gateway
* UserData
* Room
* Game

This project adopts a microservices architecture, with Docker Compose used to manage and deploy the services. The main purpose is to explore how independent services communicate with each other and how communication failures between containers can be handled.

### Communication

Requests from clients, including both HTTP and TCP connections, are received by the Gateway. The Gateway then routes requests according to their connection type and purpose.

The backend services communicate with each other using gRPC.

### Languages Chosen

The Gateway and Room services are implemented in **Go**, mainly because Go provides lightweight goroutines and efficient concurrency support, which makes it suitable for handling a large number of concurrent connections.

The UserData service is implemented using **Java and Spring Boot**, mainly because of its mature support for CRUD operations, database connectivity, and data persistence through technologies such as Spring Data JPA.


## Technology Stack

* **Go**
* **Java**
* **Spring Boot**
* **gRPC / Protocol Buffers**
* **MySQL**
* **Docker Compose**
* **Maven**
* **Git**

## Project Structure

Current:

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

## Communication

### Client → Gateway

The client communicates with the Gateway through HTTP and TCP.

#### HTTP

Gateway HTTP server:

```text
127.0.0.1:26526
```

Example health check:

```bash
curl http://127.0.0.1:26526/health
```

Expected response:

```text
ok
```

### Gateway → UserData

The Gateway communicates with the UserData service using gRPC.

Inside the Docker Compose network:

```text
userdata:9090
```

The gRPC API is defined using Protocol Buffers.

Currently supported user operations include:

* User registration
* User login

Example service definition:

```protobuf
service UsersLoginService {
    rpc UserRegistration(UserRegistrationRequest)
        returns (UserRegistrationResponse);

    rpc UserLogin(UserLoginRequest)
        returns (UserLoginResponse);
}
```

### UserData → MySQL

UserData uses Spring Data JPA to communicate with MySQL.

Inside the Docker Compose network:

```text
mysql:3306
```

The default database is:

```text
user_data
```

## Configuration

Sensitive configuration should be stored in `.env` rather than committed directly to Git.

Example `.env`:

```env
MYSQL_ROOT_PASSWORD=your_root_password
MYSQL_DATABASE=user_data
MYSQL_USER=your_user
MYSQL_PASSWORD=your_password
```

The `.env` file should be excluded from Git:

```gitignore
.env
```

Docker Compose can then reference the variables:

```yaml
environment:
  MYSQL_USER: ${MYSQL_USER}
  MYSQL_PASSWORD: ${MYSQL_PASSWORD}
```

UserData can receive the same variables through Docker Compose and use them in `application.yml`.

## Requirements

Before running the project, install:

* Docker
* Docker Compose
* Git

For local development outside Docker, the following are also useful:

* Go
* JDK 21
* Maven

## Running the Project

Clone the repository:

```bash
git clone <repository-url>
cd Duverge
```

Create a `.env` file in the same directory as `docker-compose.yml`.

Then start the complete project:

```bash
docker compose up -d
```

To build all services before starting:

```bash
docker compose up -d --build
```

Check running containers:

```bash
docker compose ps
```

View logs:

```bash
docker compose logs
```

View logs for a specific service:

```bash
docker compose logs gateway
docker compose logs userdata
docker compose logs mysql
```

Stop the project:

```bash
docker compose down
```

## Database

The project uses MySQL.

The MySQL database is configured through Docker Compose.

UserData uses JPA to manage the database schema during development.

For development:

```yaml
spring:
  jpa:
    hibernate:
      ddl-auto: update
```

## API Example

### User Registration

Endpoint:

```text
POST /login
```

Request:

```json
{
  "type": "register",
  "accountId": "123456",
  "password": "abc123"
}
```

Example:

```bash
curl -X POST http://127.0.0.1:26526/login \
  -H "Content-Type: application/json" \
  -d '{
    "type": "register",
    "accountId": "123456",
    "password": "abc123"
  }'
```

A successful registration returns:

```text
Registration successful
```

### User Login

Request:

```json
{
  "type": "login",
  "accountId": "123456",
  "password": "abc123"
}
```

Example:

```bash
curl -X POST http://127.0.0.1:26526/login \
  -H "Content-Type: application/json" \
  -d '{
    "type": "login",
    "accountId": "123456",
    "password": "abc123"
  }'
```

## Testing

Automated testing is used throughout the development process.

Currently, only basic black-box testing is implemented. More comprehensive automated tests will be added once the HTTP request flow is fully implemented and operational.

* pytest
* Docker-based integration testing

Testing infrastructure will be expanded as the project develops.

## Security

This project is currently intended for development and educational purposes.

For production deployment, additional security measures should be implemented, including:

* HTTPS for external HTTP communication
* Password hashing instead of storing plaintext passwords
* Secure session management
* Authentication and authorization
* gRPC TLS where required
* Secure management of database credentials
* Environment-specific configuration
* Database migration management

## Project Status

Duverge is currently under active development.

### Current Progress

The following features have been implemented:

* **User registration and login flow**: Gateway → gRPC → UserData → MySQL → response
* **Health check endpoint** for service availability monitoring
* **Docker Compose deployment** for managing the service environment
* **MySQL persistence** for user data

### Planned Development

The following features are planned for future development:

* **User session management**: credential-based sessions with expiration and dedicated session handling
* **UserData service expansion**: additional user-related data and functionality
* **Game room system**: room creation, management, and player handling
* **Gateway TCP server**: support for persistent TCP connections
* **Game server development**: implementation of core game logic and server-side gameplay systems


APIs and internal architecture may change during development.

## License

This project is licensed under the [MIT License](LICENSE).
