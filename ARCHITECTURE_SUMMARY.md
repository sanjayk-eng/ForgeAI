# ForgeAI Backend Architecture Summary

This document explains the backend as it is actually implemented in the current codebase. It is intentionally grounded in the runtime wiring and module composition used by the application, not the ideal future architecture.

## 1. High-level purpose

ForgeAI is a full-stack AI coding workspace. The backend is the orchestration layer that ties together:

- user identity and authentication
- workspace membership and permissions
- project data and GitHub metadata
- sandboxed filesystem execution for coding tasks
- Git status, diff, commit, and push workflows
- AI agent task execution over repository files
- email notifications and async worker flows

The main runtime entrypoint is:
- `Backend/cmd/agent/server.go`

The runtime composition is a single server process that wires modules together and governs the lifecycle of their background workers.

---

## 2. Runtime startup flow

The actual boot flow is inside `runServer()` in `Backend/cmd/agent/server.go`.

The sequence is:

1. Load configuration via `config.Load()`.
2. Build the logger and initialize Gin.
3. Configure middleware for recovery, logging, and CORS.
4. Create the OAuth provider factory.
5. Create the JWT manager if `JWTSecret` is configured.
6. Open the PostgreSQL connection if `DATABASE_URL` is set.
7. Initialize the email module and start its background workers.
8. Load auth module and create the protected router.
9. Load workspace and project modules.
10. Create the terminal module and sandbox workers.
11. Register Git, terminal, and agent HTTP routes.
12. Attach project lifecycle callbacks to terminal sandbox behavior.
13. Start Gin and wait for shutdown.

This is important: the backend is not a set of isolated packages. It is a single composed runtime where modules are registered directly into the same HTTP server and share dependencies such as DB, logger, JWT manager, and sandbox service interfaces.

### Key design point

The project root uses dependency injection at startup instead of an application container framework. Each module has a `ModuleConfig` and a `LoadModule()` function that wires its repository, service, handler, and routes.

---

## 3. Runtime module composition

### 3.1 Auth module

Location:
- `Backend/internal/modules/auth/`

Responsibilities:
- registration
- login
- email verification
- OAuth GitHub login
- token issuance and validation
- identity lookup for protected requests

This module is central to the runtime. It owns user data and provides identity access for every protected API.

Important implementation details:
- password hashing is done with the bcrypt utility package
- JWTs are created and validated through `pkg/jwt`
- OAuth data is stored through `tbl_oauth_account`
- the provider factory supports GitHub and Google flows
- the auth repository is also used by the Git module to fetch the user’s GitHub token for push operations

### 3.2 Workspace module

Location:
- `Backend/internal/modules/workspaces/workspace/`
- `Backend/internal/modules/workspaces/member/`
- `Backend/internal/modules/workspaces/invite/`

Responsibilities:
- create and read workspaces
- manage workspace membership
- invite users
- accept invites
- enforce workspace boundaries for project access

This is the top-level tenant boundary of the product. A project belongs inside a workspace, and many APIs require workspace membership.

### 3.3 Project module

Location:
- `Backend/internal/modules/project/`

This is the main domain aggregate for user projects and repository-linked work.

#### Structure
- `core/` — project entity + core DB operations
- `repository/` — repository metadata operations
- `sync/` — GitHub sync polling and state refresh
- `github/` — GitHub repo discovery and metadata logic
- `orchestrator/` — orchestration logic between core project and repo state
- `service.go` — facade service with lifecycle hooks
- `handler.go` — HTTP layer
- `route.go` — route registration

#### Real responsibilities
- create/delete/update project rows
- connect a GitHub repository to the project
- persist repository URL, default branch, and GitHub metadata
- sync GitHub state into project records
- trigger sandbox provisioning and cleanup via lifecycle events

#### Important lifecycle hooks
The project service exposes methods like:
- `SetOnCreate(...)`
- `SetOnDelete(...)`
- `SetOnBranchUpdated(...)`

Those are connected in the server startup to terminal lifecycle behavior:
- project creation triggers sandbox creation
- project deletion triggers sandbox cleanup
- branch changes trigger downstream updates

This means the project module is definitely not just CRUD; it is integrated with execution runtime lifecycle management.

### 3.4 Terminal and sandbox module

Location:
- `Backend/internal/modules/terminal/`

This is the execution runtime of the system.

Responsibilities:
- create project sandboxes
- run commands inside a container
- mount workspace directories
- clone repositories into a workspace
- read/write files inside the sandbox
- check git status/diff and manage commit/push workflows

This module is the key operational layer behind the AI coding experience.

#### Execution model
The service is built around:
- store for sandbox state
- runtime for Docker operations
- filesystem abstraction for file operations
- repository cloner logic
- repository pusher logic
- sandbox policy config

The configured image is declared in `Backend/configs/sandbox.yaml`:

```yaml
sandbox:
  image: alpine/git:latest
  git_image: alpine/git:latest
```

This matches the running container output and is the built-in execution environment used for repo operations.

#### Important runtime design
The sandbox runs with `network_mode: none` and a read-only rootfs policy. The main workspace container is isolated, but repo operations requiring network interaction are done via short-lived helper containers that mount the workspace volume.

This is a strong design constraint: most project coding happens in isolated local sandbox files, while GitHub network actions are delegated with temporary scoped access.

### 3.5 Agent module

Location:
- `Backend/internal/modules/agent/`

Responsibilities:
- gather project context from sandbox files
- send prompt payload to the configured model
- validate response payloads
- apply file edits back into the sandbox

This is the AI coding loop. It reads the repository state from the running sandbox, gives the model context, and writes the returned edits into the working tree.

In plain terms, the AI does not edit the user’s GitHub repo directly. It edits the sandbox workspace and then the system can diff, commit, and push those changes.

### 3.6 Git module

Location:
- `Backend/internal/modules/git/`

Responsibilities:
- inspect git status
- get file diffs
- commit changes
- push to GitHub using OAuth access token

The Git module is registered on the protected router and uses the terminal/sandbox execution service as its backend. It does not own the sandbox. It delegates actual repository operations to the terminal runtime infrastructure.

Key detail:
- commit author name and email come from the authenticated user
- push operations load the user’s GitHub OAuth token from the auth repository
- the token is fed into a short-lived Git helper container and not stored in repo config

This is a strong security pattern because the system leaves the repository environment clean, while using auth tokens on-demand for push actions.

### 3.7 Email module

Location:
- `Backend/internal/shared/email/`

Responsibilities:
- queue transactional emails
- process them asynchronously with a worker pool

This is support infrastructure, not the primary business domain. It is started in the server process and runs as a background worker queue.

---

## 4. Request and command architecture

### HTTP layer
The backend uses Gin. Each module usually does this:

- define a `ModuleConfig`
- create a repository/service instance
- build a handler
- register routes through `RegisterRoutes(...)`

The server composes modules in one place, then mounts them on the same engine. This keeps the main server file readable while preserving clear module boundaries.

### Protected access model
Protected routes are created with `middleware.ProtectedGroup(engine, jwtManager, appLogger)`. This means any authenticated route gets JWT validation before reaching the module handler.

### Shared concerns
The system centralizes common concerns in:
- config
- logger
- errors
- filesystem helpers
- executor interfaces
- database utilities
- JWT utilities

This is a classic layered backend design: HTTP → module handler → service → repository/data access → shared infrastructure.

---

## 5. Data flow in plain English

The real working flow of the app is:

1. User signs in and receives a JWT.
2. User creates or joins a workspace.
3. User creates a project inside that workspace.
4. The project module stores project metadata and optional repository metadata in PostgreSQL.
5. The project lifecycle triggers sandbox creation via terminal workers.
6. The sandbox mounts a workspace, clones repo content, and prepares the working tree.
7. The AI agent reads files from the mounted workspace.
8. The model returns file edits.
9. The backend writes those edits into the sandbox filesystem.
10. The user can inspect git status and diff.
11. The user can commit and push these changes to GitHub with the authenticated account.

This means the core product loop is:

workspace → project → sandbox → AI edits → git review → GitHub push

Not:

frontend direct GitHub edits

---

## 6. Background workers and async processing

The backend uses lightweight in-process asynchronous workers rather than a distributed queue system.

### Sandbox event worker
The terminal system publishes project lifecycle events like:
- project created
- project deleted
- project branch updated

These are processed by a worker that provisions or tears down sandbox resources.

### Email worker queue
Email sending is queued in memory and processed asynchronously by worker goroutines.

This is good for application-level background processing, but it is not a multi-node distributed queue.

---

## 7. Architectural reality check

The codebase is already much more structured than a simple monolith, but it is still a single-process backend with modular components.

Current reality:
- Go + Gin backend
- PostgreSQL for persistence
- modular domain structure
- Docker-based execution runtime
- GitHub OAuth and repo sync
- AI task execution on a sandboxed repo
- workspace-level multi-user model

The backend is therefore best understood as:

a modular monolith with strong infrastructure boundaries and sandboxed execution capabilities

not a microservice system.

---

## 8. Most important files to read first

If you want to understand the product quickly, read these in order:

1. `Backend/cmd/agent/server.go`
2. `Backend/internal/modules/auth/module.go`
3. `Backend/internal/modules/project/module.go`
4. `Backend/internal/modules/terminal/module.go`
5. `Backend/internal/modules/agent/service.go`
6. `Backend/internal/modules/git/service.go`
7. `Backend/configs/sandbox.yaml`

These files explain the runtime composition, the security model, the project lifecycle, and the real code execution path.

---

## 9. Bottom line

The backend is built around one central idea:

AI work happens inside a secure, sandboxed workspace attached to a project, and all user, repo, and Git operations are coordinated through a modular Go service layer.

That is the true architecture of ForgeAI as implemented today.


---

## 6. The actual backend dependency pattern

The codebase follows a layered design:

- HTTP handlers at the top
- service layer for business logic
- orchestrator layer for cross-entity operations
- repository layer for DB access
- infrastructure layer for external systems (Docker, GitHub)

This is a fairly clean modular design for a backend service, even though it is not yet a fully enterprise-scale distributed system.

---

## 7. Key execution boundaries

The project has a few crucial boundaries:

### API boundary
- Gin route handlers receive HTTP requests.
- They validate input and delegate to module services.

### Service boundary
- Business rules live here.
- Example: project creation, user validation, repo linkage, sandbox lifecycle.

### Infra boundary
- Docker runtime calls and GitHub API calls live here.
- External effects are isolated from business logic.

### AI boundary
- AI is used only to transform file content inside the sandbox workspace.
- It is not directly connected to remote git operations yet.

---

## 8. What is implemented vs not implemented

### Implemented
- auth
- workspaces
- members and invites
- project creation
- repository metadata linkage
- GitHub repository discovery/import and periodic metadata sync
- sandbox creation
- Docker host-directory bind mounting
- repo clone into workspace
- file listing, read, write, create, delete, and rename
- AI coding task execution against sandbox files
- Git status, diff, commit, and authenticated push
- worker-based async event processing for project tasks

### Not fully implemented yet
- PR generation
- distributed queue infrastructure
- full production-grade async job system
- Git operations outside the sandbox workspace
- automatic sandbox provisioning when attaching a repository to an existing project

---

## 9. The actual architecture in one sentence

ForgeAI’s current backend is a modular Go service where user and project metadata live in PostgreSQL, AI coding and Git operations run against Docker-based sandbox workspaces, and authenticated GitHub pushes are initiated through the Git API rather than by the model directly.

---

## 10. The most important takeaway

The central runtime concept is:

- project metadata in DB
- sandbox execution in Docker
- code repository cloned into sandbox
- AI agent edits files in sandbox
- Git panel reviews and commits sandbox changes
- authenticated Git push updates the configured HTTPS GitHub remote

This is the current real architecture in this codebase.

---

## 11. Detailed current execution flow

This section describes the execution path currently implemented in the backend. It separates request-time work from background work so it is clear when an HTTP request finishes and which operations continue asynchronously.

### 11.1 Server startup and dependency wiring

The composition root is `Backend/cmd/agent/server.go`. The server creates shared infrastructure first and then passes those dependencies into the modules that need them.

Startup sequence:

1. Load configuration, create the Gin engine/logger, and install request logging, recovery, and CORS middleware.
2. Create the OAuth factory; create the JWT manager and PostgreSQL connection if their settings are present.
3. Construct and start the email module (one capacity-100 in-memory queue and three workers by default).
4. Load auth, create the protected router, and load workspace routes.
5. Create the GitHub provider and load the project module. Its periodic metadata-sync loop starts when a database is configured.
6. Construct the terminal project adapter, load sandbox policy, and load the terminal module. This creates the Docker runtime, file store, PostgreSQL sandbox repository, Git clone/push helpers, application service, and capacity-100 sandbox queue.
7. Start three sandbox workers.
8. Load protected Git routes using the terminal service for execution/access and auth repository for account/profile data; register terminal and agent routes.
9. Attach project create/delete/branch callbacks to terminal, load member/invite routes, and start Gin.

The result is one Go process containing the HTTP server, email and sandbox in-memory queues, three email workers, three sandbox workers, the project metadata-sync goroutine, the database client, and external-system clients. Workers start during module setup rather than in a separate queue service.

### 11.2 Authentication request flow

For a normal authenticated request:

1. The client sends credentials or an OAuth callback request.
2. The auth handler parses and validates the request.
3. The auth service checks the user and OAuth account repositories.
4. For password-based flows, the bcrypt package verifies or creates the password hash.
5. For email verification or similar messages, the auth service creates an email job and publishes it to the email queue.
6. The handler returns the authentication result or JWT response.
7. For later protected requests, auth middleware validates the JWT before the request reaches the module handler.

Email sending does not block the request until the provider finishes. The request only publishes the job. An email worker later consumes the job and calls the configured provider.

### 11.3 Workspace and invite flow

Workspace operations are synchronous database operations unless they send an email:

1. The request enters the protected workspace route.
2. Auth middleware validates the JWT and identifies the user.
3. The workspace handler validates the payload and calls the workspace service.
4. The service verifies workspace ownership or membership.
5. The repository reads or writes workspace and member records in PostgreSQL.
6. For an invite, the service creates the invite record and publishes an invite email job.
7. The HTTP response is returned while the email worker handles delivery separately.

The invite email is therefore a notification job in the same email queue as verification and general notification messages. There is not a separate invite queue.

### 11.4 Project creation flow

The project creation flow crosses the project and terminal modules:

1. The client sends a protected project creation request.
2. Auth middleware validates the JWT.
3. The project handler parses the request and passes it to the project service or orchestrator.
4. The project layer validates the workspace relationship and project input.
5. The project core repository inserts the project record in PostgreSQL.
6. If a GitHub repository is selected, the repository layer stores its URL, default branch, owner, and GitHub repository ID. The OAuth token is stored separately against the user in `tbl_oauth_account`.
7. The project service invokes its configured create callback.
8. The callback calls `terminalModule.OnProjectCreated`.
9. `OnProjectCreated` publishes a `project.created` event to the sandbox worker queue.
10. The project HTTP response can return before sandbox provisioning completes.
11. A sandbox worker receives the event and loads the project and repository metadata.
12. The worker asks the terminal application service to create the sandbox.
13. The Docker runtime creates the volume and container and mounts the project workspace.
14. Terminal infrastructure launches a short-lived Git helper with bridge networking to clone the repository into the mounted workspace when repository metadata exists.
15. The worker provisions or refreshes the files and records the resulting sandbox state.

The important boundary is step 9: project persistence is request-time work, while sandbox provisioning is background work. Publishing is non-blocking; if the in-memory sandbox event queue is full, the event is dropped after a warning and the project request can still succeed.

### 11.5 Project deletion flow

Deletion currently uses the terminal module as a cleanup boundary:

1. The project handler authenticates and authorizes the request.
2. The project service first invokes the configured delete callback, which calls `terminalModule.OnProjectDeleted`.
3. The terminal module delegates cleanup to the sandbox worker.
4. The worker destroys the sandbox and removes its host workspace directory.
5. Only after cleanup succeeds does the project service delete project records.

The cleanup is controlled by the terminal module so project services do not need to know Docker implementation details.

### 11.6 Repository branch change flow

The current branch flow is separate from the initial project creation flow:

1. The client requests a repository branch update.
2. The project service validates the project and branch input.
3. The project repository updates `default_branch` in PostgreSQL.
4. The project service invokes the configured branch-updated callback.
5. The callback calls `terminalModule.OnProjectBranchUpdated`.
6. The terminal module delegates to `SandboxWorker.RefreshProjectBranch`.
7. The worker acquires the project operation lock so concurrent branch refreshes for the same project do not overlap.
8. The worker loads the latest project and repository metadata.
9. The terminal application service and Docker runtime refresh the sandbox repository using the new branch.
10. The stale checkout is cleared or recreated when required, and the workspace is updated to the selected branch.

The database branch value is therefore only the source of truth for project metadata. Updating the database alone does not update the files already present in the Docker workspace; the terminal refresh callback is required for that second state transition. If refresh fails, the repository branch metadata has already changed and is not rolled back with the sandbox checkout.

### 11.7 AI coding task flow

The AI task flow is primarily synchronous from the API caller's perspective, but it operates on files that were prepared asynchronously during sandbox provisioning:

1. The client submits an AI task for a project.
2. The agent handler authenticates the request and validates the project context.
3. The agent service checks that the user can access the project and its sandbox.
4. The service asks the sandbox file API for the relevant files under `/workspace`.
5. The service builds a project context and prompt for the configured LLM service.
6. The LLM returns a structured response containing file changes.
7. The agent service validates the response format and rejects invalid or unsafe changes according to its validation rules.
8. The service writes accepted file changes through the sandbox file API one file at a time.
9. The response reports the task result to the client.

The AI model does not edit the host repository directly and does not call GitHub directly. The effective target is the mounted Docker workspace. The backend is the orchestrator between the model and that workspace. Writes are not transactional: a later file-write failure can leave earlier changes applied in the sandbox.

### 11.8 Email notification flow

The notification path has one queue and one worker pool:

1. Auth, invite, or another service calls the email service.
2. The email service validates recipient, subject, and message data.
3. The service creates an email `Job` with a job type such as verification, invite, or notification.
4. The service publishes the job to `InMemoryQueue.jobs`.
5. If the queue has capacity, publishing returns immediately.
6. If the queue is full, publishing returns `ErrQueueFull`; the caller can log or handle that failure.
7. One of the three email worker goroutines receives the job.
8. The worker calls the configured email provider, currently Resend.
9. The worker logs success or failure.

The queue is process-local. Jobs are lost if the process stops before they are delivered because there is no persistent queue or retry store in the current implementation.

### 11.9 Sandbox event flow

The sandbox queue has one event channel and three worker goroutines by default:

1. A project lifecycle callback creates a `ProjectEvent`.
2. `SandboxWorker.Publish` attempts a non-blocking send to `events`.
3. If the queue is full, the event is dropped and a warning is logged.
4. An available sandbox worker receives the event.
5. `process` checks cancellation and shutdown state.
6. `handleEvent` dispatches the event to the appropriate project operation.
7. The worker loads project and repository state.
8. The worker calls the terminal application service.
9. The application service uses the Docker runtime and sandbox repository.
10. The worker records errors and releases the per-project operation lock.

Unlike the email queue, the sandbox `Publish` method is intentionally non-blocking. This keeps the caller responsive, but it means a full queue can result in a dropped project event. Project detail can request sandbox creation again with `ensure=true`, but there is no durable event record or guaranteed queue retry.

### 11.10 Git panel and push flow

Git operations are request-time operations on the existing sandbox; they do not run in the sandbox event queue:

1. The project detail page passes the current access token and sandbox ID to the Git panel.
2. The panel requests Git status and diff from the protected Git routes.
3. The Git handler validates sandbox access through the terminal application service.
4. The Git service runs status, diff, and commit commands in `/workspace` through the network-disabled sandbox executor.
5. For a commit, the service optionally stages all changes and runs `git commit` with the supplied message and the authenticated user's profile name/email as command-scoped Git config.
6. For a push, the handler loads the user's GitHub OAuth token from the auth repository.
7. The Git service checks that the selected remote resolves to an HTTPS `github.com` URL.
8. The terminal application passes the repository volume, workspace path, remote, branch, and token to the terminal infrastructure pusher.
9. The pusher starts a disposable Git helper with bridge networking and mounts the same workspace volume. It reads the token from stdin and creates a temporary HTTP authorization header for Git; credentials are not persisted in `.git/config` or passed in Docker arguments.
10. The panel refreshes status and diff after the operation.

The push succeeds only when the user has a connected GitHub account with repository write access and the sandbox repository has a valid HTTPS GitHub remote. The sandbox container itself has no network access; only the one-shot clone/push helpers use bridge networking.

### 11.11 Realtime workspace synchronization

Realtime updates reuse the terminal module and existing file/Git APIs:

1. `WorkspaceWatcher` recursively watches directories under the host bind mount for a running sandbox. Newly created directories are added to the watch set.
2. Filesystem paths are normalized to project-relative paths (for example, `/workspace/src/App.tsx` becomes `src/App.tsx`). `.git`, `node_modules`, `dist`, `build`, `cache`, and other generated directories are ignored.
3. Events for one path are coalesced for 100 ms. File create/change/delete/rename events include path metadata only; the Git status event follows workspace changes.
4. The terminal module's in-memory event hub routes messages by `project_id`. Subscriber buffers are bounded; if a client falls behind, the hub replaces queued data with `sync.required` instead of blocking filesystem writes.
5. `GET /projects/:project_id/ws` uses the existing protected router. Browser clients provide the existing JWT in the `forgeai-auth.<token>` WebSocket subprotocol; the handler validates project/workspace access before subscription and periodically while connected. WebSocket origins use configured CORS/frontend origins.
6. Agent start/progress/completed and sandbox status events use this same project-scoped hub. Existing Agent, Git, and sandbox operations remain the source of truth; no duplicate file or Git APIs were added.
7. React consumes events on the project detail page. It updates the existing lazy file-tree cache for loaded directories, marks changed files, refreshes only the selected file through the existing read API, and refreshes Git status/diff through the existing Git APIs. Monaco can show a before/after `DiffEditor` snapshot; unsaved editor drafts are preserved when external changes arrive.
8. On socket open/reconnect or `sync.required`, the page resynchronizes the sandbox, root and loaded directories, current file, and Git status/diff using existing APIs. Events are notifications, not durable state.

The user's Ctrl+S path is unchanged: Monaco calls the existing save API, Terminal writes to `/workspace`, fsnotify observes that bind-mounted filesystem change, and the hub notifies project subscribers. The realtime consumer never writes the file back, so the notification cannot create a save loop.

---

## 12. Queue and worker inventory

| Queue | Location | Capacity | Workers | Job or event types | Persistence |
| --- | --- | ---: | ---: | --- | --- |
| Email queue | `internal/shared/email/queue.go` | 100 | 3 | verification, invite, notification, and other email jobs | In-memory only |
| Sandbox event queue | `internal/modules/terminal/worker/sandbox_worker.go` | 100 | 3 | project creation and sandbox lifecycle events | In-memory only |

Current totals:

- 2 queues
- 2 worker pools
- 6 default worker goroutines
- 100 buffered entries per queue
- 0 persistent or distributed queues

The word “notification” is an email job type, not a third queue. It is processed by the shared email queue.

---

## 13. Request-time versus background-time behavior

### Request-time operations

- JWT validation
- request validation
- authorization checks
- project and workspace database operations
- repository metadata updates
- LLM request execution for an agent task
- sandbox file reads and writes requested by the agent
- publishing email or sandbox events

### Background operations

- email provider delivery
- project-created sandbox provisioning
- sandbox cleanup after project deletion
- sandbox repository refresh after branch changes
- Docker container and volume operations triggered by sandbox events

This distinction explains why a successful API response does not always mean that the Docker sandbox or email provider operation has already completed.

---

## 14. Current failure and recovery behavior

The current implementation has process-local failure behavior:

- Database errors are returned through the request path or worker logs.
- Email queue overflow returns `ErrQueueFull` to the publishing service.
- A stopped email queue returns `ErrQueueStopped`.
- Sandbox queue overflow logs `sandbox worker queue full` and drops the event.
- Repository metadata sync is a special case: `PENDING` and `FAILED` rows are eligible for the periodic worker again on its next interval (30 seconds by default). This retry does not reprovision the sandbox checkout.
- Worker panics are recovered and logged by the sandbox worker process loop.
- Context cancellation stops workers.
- Queue contents are not recovered after a process restart.
- There is no general retry, dead-letter queue, durable job record, or distributed lock.
- `Logout` currently returns success without revoking a JWT; stateless tokens remain valid until expiration because revocation state is not persisted.
- Queue shutdown needs hardening: email and sandbox publishers can race with `Stop()` closing their publish channels.

Per-project sandbox operations use an in-process project lock to reduce concurrent operations for the same project. This lock does not coordinate separate backend processes.

---

## 15. Current state transition summary

The project moves through these practical states:

1. Project metadata is requested and validated.
2. Project and repository records are persisted.
3. A project event is published.
4. The sandbox is provisioned or refreshed.
5. The repository is available under the sandbox workspace.
6. The AI reads and changes workspace files.
7. The user can inspect the changed sandbox files.
8. The user can commit the changes in the sandbox and push them to its configured GitHub remote from the Git panel.

The current architecture is therefore a sandbox-first editing system. GitHub supplies repository metadata and clone source, while Docker workspace state is the active working copy and the Git panel provides the explicit commit/push path back to the remote.

---

## 16. Current architecture limitations

The design is suitable for a single backend process and development or early-stage workloads, but the following limitations are important for production planning:

- In-memory queues do not survive restarts.
- A full sandbox queue can drop events.
- Worker counts are process-local and do not scale across instances.
- There is no durable task status model for every background operation.
- There is no distributed coordination for the same project across multiple backend instances.
- Pull requests are not generated automatically after a push.
- Git diff preview is based on `git diff HEAD`; untracked files appear in status but are not represented as patch content in the current preview.
- The sandbox and Git helper use the mutable `alpine/git:latest` image tag, so their Git version can change when the image is updated.
- There is no durable audit trail for every file change produced by the model.
- OAuth access tokens are stored as raw text in `tbl_oauth_account.access_token`; no application-layer encryption is implemented in the auth repository.
- The GitHub OAuth authorization/callback flow does not generate or validate a `state` value, leaving the standard OAuth CSRF defense absent.
- GitHub OAuth requests the broad classic `repo` scope; review a narrower GitHub App permission model before production use.
- Git helper behavior has no Docker/GitHub end-to-end test; current tests focus on service-level command selection and output parsing.
- Connecting a repository to an existing project does not trigger sandbox clone/provisioning; create/import flows do.
- Agent tasks have no durable job record and no rollback if applying a multi-file response fails partway through.
- `conversation`, `llm`, `permission`, and `tool` packages are not wired into `runServer()`; the agent currently makes its model HTTP call directly.
- Realtime events are process-local and not replayed after a server restart; clients rely on reconnect resync through the existing APIs.
- A full subscriber buffer loses individual events and sends `sync.required`; exact intermediate history is intentionally not retained.
- Watcher behavior is covered by local filesystem tests, but a Docker Desktop/container-to-host file-event integration test is still needed for each supported deployment environment.

These are current-state observations, not missing pieces that are silently assumed to exist.
