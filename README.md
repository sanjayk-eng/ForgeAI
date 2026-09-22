# ForgeAI
daily and wikely task

## Current Backend Overview

The backend is a Go API built with Gin and centered around a modular architecture. The main server entry point lives in Backend/cmd/agent/server.go and starts the HTTP server, loads environment configuration, builds logger and database clients, sets up middleware, and mounts the auth module.

### What the current code is doing

- Server startup:
  - Backend/cmd/agent/main.go calls runServer().
  - Backend/cmd/agent/server.go loads config with config.Load(), sets Gin mode based on APP_ENV, creates the engine, applies middleware, exposes /health, initializes OAuth providers, connects to Postgres if DATABASE_URL is set, creates JWT manager if configured, then starts the app on HOST:PORT.

- Configuration and environment:
  - Backend/internal/config/config.go reads .env files and environment variables like APP_ENV, HOST, PORT, DATABASE_URL, JWT_SECRET, CORS origins, and OAuth credentials.
  - It also supports log level, format, and source configuration.

- Authentication module:
  - Backend/internal/modules/auth/route.go defines endpoints such as /auth/register, /auth/login, /auth/refresh, /auth/logout, /auth/callback, and the protected /auth/me route.
  - Backend/internal/modules/auth/module.go wires the repository, service, handler, JWT, provider factory, and routes.
  - Backend/internal/modules/auth/handler.go handles request validation, response mapping, and HTTP status codes.
  - Backend/internal/middleware/auth.go enforces Bearer JWT authentication and stores the user ID in request context for protected endpoints.

- Shared infrastructure:
  - Backend/internal/shared/logger provides logging abstraction and zap-based implementation.
  - Backend/internal/shared/errors defines consistent API error responses and success wrappers.
  - Backend/internal/shared/executor defines the command execution contract used for terminal-like actions.
  - Backend/pkg/database/postgres.go and Backend/pkg/jwt/jwt.go provide DB access and JWT creation/validation utilities.

- Module structure:
  - The codebase already has scaffolded modules for agent, conversation, git, llm, permission, tool, workspace, and terminal; the active production path is currently centered on the auth flow and the server bootstrapping.
  - Some folders like the git package are still minimal or incomplete, which means the architecture is intentionally prepared for expansion but not all features are implemented yet.

### Backend structure

```text
Backend/
|-- cmd/
|   `-- agent/
|       |-- main.go
|       `-- server.go
|-- internal/
|   |-- config/
|   |   `-- config.go
|   |-- middleware/
|   |   |-- auth.go
|   |   |-- http.go
|   |   |-- logging.go
|   |   |-- recovery.go
|   |-- modules/
|   |   |-- agent/
|   |   |-- auth/
|   |   |-- conversation/
|   |   |-- git/
|   |   |-- llm/
|   |   |-- permission/
|   |   |-- terminal/
|   |   |-- tool/
|   |   `-- workspace/
|   |-- router/
|   `-- shared/
|       |-- errors/
|       |-- executor/
|       |-- filesystem/
|       |-- logger/
|-- migrations/
|-- pkg/
|   |-- bcrypt/
|   |-- database/
|   |-- jwt/
|   `-- validate/
|-- prompts/
|-- go.mod
`-- go.sum
```

### Current status

This backend is in an early-to-mid modular architecture stage. The server boots and exposes an auth API, the configuration layer is active, and the system is designed to grow into a more complete AI-agent platform with workspace, git, terminal, tool, and permission features. The next implementation steps are usually to complete the remaining modules and wire them into the router and service layer.

## Backend Structure

```text
Backend/
|-- cmd/
|   `-- agent/
|       `-- main.go
|-- internal/
|   |-- modules/
|   |   |-- agent/        (handler.go, service.go, repository.go, model.go, route.go)
|   |   |-- conversation/ (handler.go, service.go, repository.go, model.go)
|   |   |-- llm/          (service.go, repository.go, model.go)
|   |   |-- tool/         (service.go, registry.go, model.go)
|   |   |-- workspace/    (service.go, repository.go, model.go)
|   |   |-- permission/   (service.go, repository.go, model.go)
|   |   `-- git/          (service.go, repository.go, model.go)
|   |-- shared/
|   |   |-- executor/     (contract and result types)
|   |   |   |-- runtime/  (runner.go, runner_test.go)
|   |   |   |-- shell/    (executor.go, bash.go, cmd.go, powershell.go, zsh.go)
|   |   |   `-- factory/  (factory.go, factory_test.go)
|   |   |-- filesystem/   (filesystem.go)
|   |   |-- errors/       (errors.go)
|   |   |-- logger/       (logger.go)
|   |   `-- types/        (common.go)
|   |-- middleware/       (recovery.go, logging.go)
|   |-- config/           (config.go)
|   `-- database/         (postgres.go)
|-- prompts/
|-- configs/
|-- migrations/
|-- go.mod
`-- go.sum
```
