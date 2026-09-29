# ForgeAI

ForgeAI is a workspace-based AI development platform that combines a Go backend with a React + TypeScript frontend. The product is structured around user auth, project management, GitHub integration, AI agent orchestration, and sandboxed terminal execution.

## Overview

### Backend
The backend is built in Go using Gin and follows a modular architecture. It is organized under the Backend directory and exposes a REST API for authentication, workspaces, project management, terminal execution, Git operations, and AI-agent orchestration.

### Frontend
The frontend is built in React 19 with Vite and TypeScript. It uses React Router for navigation, TanStack Query for data fetching, Zustand-ready patterns, and a modular workspace UI structure.

---

## Backend architecture

### Core entry point
- Backend/cmd/agent/main.go starts the application.
- Backend/cmd/agent/server.go sets up the Gin server, config, middleware, JWT, OAuth providers, Postgres wiring, sandbox services, email workers, and all module registrations.

### Key responsibilities
- Auth and sessions
  - Login, registration, refresh, OAuth callback, email verification, and JWT-protected access.
  - Managed by Backend/internal/modules/auth and middleware.
- Workspace management
  - Workspace creation, membership, invitations, and access control.
  - Located under Backend/internal/modules/workspaces.
- Project management
  - Project CRUD, repository integration, GitHub sync, and project orchestration.
  - Located under Backend/internal/modules/project.
- Terminal and sandbox execution
  - Secure command execution, container lifecycle management, and workspace operations.
  - Located under Backend/internal/modules/terminal and related shared executor packages.
- Git and agent flows
  - Git automation, repository actions, and AI agent integrations.
  - Managed through Backend/internal/modules/git and Backend/internal/modules/agent.
- Shared infrastructure
  - Config, logger, database, validation, email service, executor, filesystem utilities, and error handling.

### Current working config
The development sandbox uses a dedicated image with Go, Node.js/npm, Git, and Bash:
- Backend/configs/sandbox.yaml
- Image: forgeai-sandbox:latest

React, TypeScript, and JavaScript projects use their normal npm dependencies. The sandbox currently has `network_mode: none`, so npm and Go dependencies must already be available locally; registry downloads are unavailable during sandbox execution.

### Backend structure
```text
Backend/
├─ cmd/
│  └─ agent/
│     ├─ main.go
│     └─ server.go
├─ internal/
│  ├─ config/
│  ├─ middleware/
│  ├─ modules/
│  │  ├─ agent/
│  │  ├─ auth/
│  │  ├─ conversation/
│  │  ├─ git/
│  │  ├─ llm/
│  │  ├─ permission/
│  │  ├─ project/
│  │  ├─ terminal/
│  │  ├─ tool/
│  │  └─ workspaces/
│  ├─ shared/
│  └─ database/
├─ pkg/
│  ├─ bcrypt/
│  ├─ database/
│  ├─ jwt/
│  └─ validate/
├─ configs/
├─ migrations/
├─ prompts/
├─ go.mod
├─ go.sum
└─ schema.sql
```

---

## Frontend architecture

### App shell
- Frontend/src/App.tsx wraps the app in:
  - QueryClientProvider
  - ThemeProvider
  - ToastProvider
  - AuthProvider
  - RouterProvider

### Routing
- Frontend/src/app/router.tsx defines the app routes.
- Auth routes include login, register, OAuth callback, and email verification.
- Protected routes include workspace layout, overview, projects, project detail, members, and settings.

### Frontend features
- Auth flow:
  - Login, register, protected route gating, OAuth callback handling.
- Workspace UI:
  - Overview pages, members, settings, invites, and project dashboard.
- Project UI:
  - Project listing and detail pages for repository-driven work.
- UI stack:
  - React Router
  - TanStack Query
  - Monaco editor for code editing
  - Lucide icons
  - Tailwind styling

### Frontend structure
```text
Frontend/
├─ public/
├─ src/
│  ├─ app/
│  │  ├─ queryClient.ts
│  │  └─ router.tsx
│  ├─ modules/
│  │  ├─ auth/
│  │  └─ workspace/
│  ├─ shared/
│  │  ├─ api/
│  │  ├─ auth/
│  │  ├─ hooks/
│  │  ├─ realtime/
│  │  └─ ui/
│  ├─ App.tsx
│  ├─ index.css
│  └─ main.tsx
├─ index.html
├─ package.json
├─ tsconfig.json
├─ vite.config.ts
└─ README.md
```

---

## How the system fits together

1. The frontend authenticates users through the backend auth module.
2. Protected pages rely on JWT-based access checks.
3. Workspace and project APIs are served by the Go backend.
4. The backend can create or manage sandboxed terminal containers for repos and commands.
5. GitHub and project sync logic connect user accounts and repositories to workspace activities.
6. The AI agent layer sits on top of these services and can interact with project and terminal execution workflows.

---

## Run locally

### Backend
```bash
cd Backend
go mod download
go run ./cmd/agent
```

### Frontend
```bash
cd Frontend
npm install
npm run dev
```

### Build the sandbox image
```yaml
sandbox:
  image: forgeai-sandbox:latest
```

Build the image from the repository root before starting the backend:

```bash
docker build -f Backend/Dockerfile.sandbox -t forgeai-sandbox:latest Backend
```

---

## Notes

This project is currently in an active product-development phase: the backend is already modular and service-oriented, while the frontend is structured around a protected workspace experience. The overall direction is a full-stack AI coding workspace where project data, git operations, sandbox execution, and user workflows are all connected through a single platform.

