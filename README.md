# ForgeAI

ForgeAI is a multi-tenant AI coding workspace. Users work inside organizations, create projects, and use an isolated Docker-backed sandbox to edit files, run a terminal, review Git changes, preview applications, and request AI-generated file changes.

This guide describes the behavior wired into the current Go backend and React frontend. The active product model is **organizations and projects**; older workspace-oriented documentation does not describe the current runtime.

## Application structure

```text
Backend/
  cmd/agent/                 Server entry point and dependency wiring
  internal/config/           Environment and .env configuration
  internal/middleware/       HTTP, CORS, logging, recovery, and authentication
  internal/modules/
    auth/                    Registration, login, OAuth, JWT identity
    organization/            Organizations, membership, invitations
    project/                 Organization-scoped project records
    github/                  GitHub App integration and repository export
    terminal/                Sandboxes, files, terminals, and project events
    git/                     Sandbox Git status, history, commit, and push
    preview/                 Signed access to project previews
    agent/                   AI task requests and validated file changes
  migrations/                PostgreSQL schema migrations
  configs/sandbox.yaml       Sandbox execution policy
  compose.yaml               Optional PostgreSQL service definition

Frontend/
  src/app/                   Query client and browser routes
  src/modules/auth/          Session, authentication pages, and guards
  src/modules/organization/  Organization and project UI, API, and state
  src/shared/                API client, auth storage, and shared UI
```

The backend is composed as one Gin server. Each module wires its repository, service, handler, and routes at startup. PostgreSQL, the JWT manager, logger, email service, and sandbox services are passed to modules through their configuration.

## Runtime flow

### Backend startup

`Backend/cmd/agent/main.go` invokes `runServer()` in `server.go`. Startup loads settings, configures logging and Gin middleware, opens PostgreSQL when `DATABASE_URL` is set, and prepares JWT and OAuth services. It starts the email workers and wires authentication, organizations, GitHub, projects, terminal/sandbox, preview, Git, and agent modules into the `/api/v1` router.

The project module's creation and deletion hooks enqueue sandbox lifecycle events for the terminal worker. The preview gateway is mounted as a catch-all handler for preview hosts. On shutdown, the server stops email and sandbox workers, closes terminal sessions, and closes the database.

### Frontend startup and authentication

`Frontend/src/main.tsx` mounts the app. `App.tsx` provides React Query, theme, toast, authentication, and routing contexts.

On startup, the auth provider reads stored tokens and tries to load the current user. If that fails, it tries the refresh token; if refresh or identity lookup also fails, it clears the session. Registration sends an email verification link; password login requires a verified account. OAuth callback and verification pages are also part of the frontend flow. Protected routes redirect signed-out users to login. API calls use the shared client, which prefixes requests with `/api/v1` and sends bearer tokens from feature API helpers.

After login, the app honors a saved invite redirect first. Otherwise it loads the user's organizations, creates **My Organization** if none exist, and navigates to the first organization. The organization layout loads organizations and keeps the selected organization in the route.

### Organization invitations

Organization owners and admins can invite a user by email with the `ADMIN` or `MEMBER` role. The backend creates a pending invite with a seven-day expiry, stores a hash of its random token, prevents duplicate pending invites through the database, and queues an email link.

The invite link opens the public invite page. A signed-out user is redirected to login or registration and then returned to the invite. An authenticated user can accept only when their account email matches the invite email. Acceptance adds them to the organization; an invite can also be rejected, revoked by an organization manager, or marked expired when an expired invite is accessed. Signed-in users can also load their pending invitations.

### Projects and sandboxes

Projects belong to an organization. Creating a project checks organization access, persists the project, and invokes the project-created hook. The hook queues a `project.created` event for the sandbox worker, which provisions the sandbox asynchronously. Deleting a project similarly queues sandbox cleanup.

When a project detail page opens, it loads the sandbox and watches transitional sandbox states. It subscribes to a project WebSocket for file, Git, sandbox, and agent events, reconnecting and refreshing API-backed state when needed. The project workspace provides:

- **Files and editor:** browse, read, create, save, rename, and delete files in the sandbox.
- **Terminal:** create interactive shell sessions over WebSockets.
- **Git:** inspect status, diffs, and history; commit, revert, and push changes. GitHub export is available through the organization GitHub integration.
- **Preview:** request project preview information and access the running preview through the signed preview gateway.
- **Agent:** submit a task when the configured model and sandbox are available. The backend reads project context, requests a model response, validates proposed file changes, writes them into the sandbox, and publishes progress events. The agent writes files; it does not run shell commands.

Sandbox, file, terminal, Git, preview, and agent operations validate the signed-in user's access to the project or sandbox.

## Main browser routes

| Route | Purpose |
| --- | --- |
| `/login`, `/register` | Sign in or create an account |
| `/auth/callback`, `/auth/verify` | OAuth callback and email verification |
| `/accept-invite/:token` | View and respond to an invitation link |
| `/accept-invite/id/:inviteId` | Respond to an invitation after signing in |
| `/organizations/:organizationId` | Organization overview |
| `/organizations/:organizationId/projects` | Organization project list |
| `/organizations/:organizationId/projects/:projectId` | Project workspace |
| `/organizations/:organizationId/members` | Organization members and invitations |
| `/organizations/:organizationId/settings` | Organization settings |
| `/organizations/:organizationId/settings/integrations` | Organization integrations |

Unknown frontend routes redirect to `/`.

## Backend API surface

All API routes use the `/api/v1` prefix. Unless listed as public below, routes are registered on the JWT-protected router.

| Area | Main routes |
| --- | --- |
| Health | `GET /health` (outside `/api/v1`) |
| Authentication | `POST /auth/register`, `/auth/login`, `/auth/refresh`, `/auth/logout`; `GET /auth/me`, `/:provider/connect`, `/auth/callback`, `/auth/verify-email` |
| Organizations | `GET/POST /organizations`; `GET/PATCH/PUT/DELETE /organizations/:organization_id` |
| Members | `GET /organizations/:organization_id/members`; `PATCH /organizations/:organization_id/members/:user_id/role`; `DELETE /organizations/:organization_id/members/:user_id` |
| Invitations | `POST/GET /organizations/:organization_id/invites`; `POST /invites/:token/accept`; `GET /invites/my-pending`; `PATCH /invites/:invite_id`; public `GET /public/invites/:token` and `POST /public/invites/:token/reject` |
| Projects | `GET/POST /organizations/:organization_id/projects`; `GET /organizations/:organization_id/projects/:project_id`; `GET/PATCH/PUT/DELETE /projects/:project_id` |
| GitHub integration | Organization integration and project export routes; public GitHub callback and webhook |
| Sandbox and files | Project sandbox, file listing/read/write/create/delete/rename routes under `/projects` and `/sandboxes` |
| Terminal and realtime | Project WebSocket, terminal shell/session routes, and terminal session WebSocket |
| Git | `/sandboxes/:sandbox_id/git/status`, `diff`, `history`, `commit`, `revert`, and `push` |
| Preview | `GET /projects/:project_id/preview`; preview-host requests are served by the gateway |
| Agent | `GET /sandboxes/:sandbox_id/agent/status`; `POST /sandboxes/:sandbox_id/agent/tasks` |

The backend returns API responses through shared success/error helpers. See each module's `route.go` (or the terminal handler's route registration) for the authoritative route definitions and HTTP methods.

## Configuration and local development

Use Go 1.26, Node.js compatible with the Vite toolchain, PostgreSQL 16+, and Docker for project sandboxes. The backend searches the current directory and its parent directories for a `.env` file. Configuration defaults include host `127.0.0.1`, port `8080`, and frontend URL `http://localhost:5173`.

Set at least `DATABASE_URL` and `JWT_SECRET` for a usable database-backed authenticated application. Apply the SQL migrations in `Backend/migrations` to PostgreSQL before using database-backed routes. Add provider credentials only for the integrations you use.

| Variable | Purpose |
| --- | --- |
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Signs access and refresh tokens |
| `HOST`, `PORT`, `APP_ENV` | Backend bind address and runtime mode |
| `FRONTEND_URL`, `CORS_ALLOWED_ORIGINS` | Frontend links and allowed browser origins |
| `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL` | Optional Google OAuth |
| `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GITHUB_REDIRECT_URL` | Optional GitHub OAuth |
| `GITHUB_APP_ID`, `GITHUB_APP_SLUG`, `GITHUB_APP_PRIVATE_KEY` or `GITHUB_APP_PRIVATE_KEY_PATH`, `GITHUB_APP_WEBHOOK_SECRET` | Optional organization-level GitHub App |
| `RESEND_API_KEY`, `RESEND_FROM_EMAIL` | Invite and account email delivery |
| `AI_API_KEY`, `AI_BASE_URL`, `AI_MODEL` | AI coding agent; the key is required to enable model calls |
| `PREVIEW_PUBLIC_ORIGIN`, `PREVIEW_SIGNING_KEY` | Preview gateway; production requires an HTTPS origin and a signing key of at least 16 bytes |
| `JWT_ACCESS_TTL`, `JWT_REFRESH_TTL`, `LOG_LEVEL`, `LOG_FORMAT`, `LOG_SOURCE` | Token lifetime and logging options |

The coding agent sends OpenAI-compatible Chat Completions requests to the configured `AI_BASE_URL` with `AI_API_KEY` and `AI_MODEL`. For OpenRouter, configure the backend environment like this (use the model ID supported by your OpenRouter account):

```dotenv
AI_BASE_URL=https://openrouter.ai/api/v1
AI_API_KEY=your-openrouter-api-key
AI_MODEL=provider/model
```

The backend's built-in fallback URL and model are OpenAI defaults; setting these environment variables selects OpenRouter instead. Keep the API key in the backend environment and never expose it in the frontend.

Run the backend from `Backend` so the sandbox policy at `Backend/configs/sandbox.yaml` is found:

```powershell
Set-Location Backend
go run ./cmd/agent
```

Run the frontend in a separate terminal:

```powershell
Set-Location Frontend
npm install
npm run dev
```

The frontend API base URL defaults to `http://127.0.0.1:8080`. Set `VITE_API_URL` when the backend is hosted elsewhere. The PostgreSQL service definition is in `Backend/compose.yaml`.

## Validation commands

```powershell
Set-Location Backend
go test ./...
```

```powershell
Set-Location Frontend
npm run build
npm run lint
```
