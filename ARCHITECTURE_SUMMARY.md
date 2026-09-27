# ForgeAI Backend Architecture Summary

This document explains the backend architecture as it is currently implemented in the codebase, not the ideal future design.

## 1. High-level overview

The backend is a Go service built around Gin HTTP handlers, modular service packages, PostgreSQL persistence, and Docker-based sandbox execution for AI coding tasks.

The main runtime entrypoint is:
- `Backend/cmd/agent/server.go`

The server wires together these major modules:
- Auth
- Workspace
- Project
- Terminal / sandbox
- Agent
- Member / invite
- Email

The backend is not a single monolith service with one giant flow. It is organized into modules, each with its own service, repository, handler, and routes.

---

## 2. Runtime startup flow

The main server bootstraps all modules in `cmd/agent/server.go`.

Execution order:
1. Initialize database connection
2. Initialize logger and email infrastructure
3. Initialize auth module
4. Initialize workspace modules
5. Initialize GitHub provider and auth factory
6. Initialize project module
7. Initialize terminal module
8. Start background worker(s)
9. Register HTTP route groups
10. Start the Gin server and wait for shutdown

The server config also wires together project lifecycle hooks with terminal lifecycle behavior:
- project creation triggers sandbox creation
- project deletion triggers sandbox cleanup

This wiring is important because it shows the backend is designed as a connected system, not just isolated services.

---

## 3. Module structure and responsibilities

### 3.1 Auth module

Location:
- `Backend/internal/modules/auth/`

Responsibilities:
- user registration
- login
- OAuth login with GitHub
- JWT issuance and validation
- email verification flow

Core files:
- `module.go`
- `service.go`
- `handler.go`
- `route.go`
- `provider/`

Important concept:
- Auth is a real first-class backend module and is not just a middleware.
- It owns user identity and identity-related access control.

---

### 3.2 Workspace module

Location:
- `Backend/internal/modules/workspaces/workspace/`
- `Backend/internal/modules/workspaces/member/`
- `Backend/internal/modules/workspaces/invite/`

Responsibilities:
- create workspace
- load workspace by ID
- manage workspace members
- invite users
- accept workspace invites

Important concept:
- Workspaces are the top-level tenant/container for projects and team access.
- Project data belongs to a workspace.

---

### 3.3 Project module

Location:
- `Backend/internal/modules/project/`

Responsibilities:
- create projects
- list and find projects
- update project metadata
- delete projects
- connect a GitHub repository to a project
- store repository URL and default branch
- sync project data with GitHub metadata

Main implementation layers:
- `core/` - project core entity and repository operations
- `repository/` - repository connection metadata to GitHub repo
- `orchestrator/` - composite orchestration logic between core project and repo
- `sync/` - sync logic with GitHub and project state
- `github/` - GitHub repo discovery and import logic

This module is the bridge between the application’s project domain and external GitHub metadata.

Important logic:
- `ProjectOrchestrator.Create()` creates the project row and optionally the repository metadata row.
- `SetOnCreate` and `SetOnDelete` are used to trigger sandbox worker lifecycle events.

---

### 3.4 Terminal / sandbox module

Location:
- `Backend/internal/modules/terminal/`

This is one of the most important modules because it provides the execution environment for code work.

Responsibilities:
- create a sandbox for a project
- start/stop/restart/destroy sandbox instances
- track sandbox status
- create a Docker container and workspace volume
- clone a repository into a workspace folder
- list, read, and write files inside the sandbox

Subareas:
- `application/` - service layer for sandbox lifecycle and file operations
- `infrastructure/` - Docker runtime and DB-backed repository persistence
- `policy/` - sandbox configuration policy
- `worker/` - background async workers for project lifecycle events
- `domain/` - sandbox domain entities and status rules

This module is effectively the execution environment where the AI actually interacts with code.

---

### 3.5 Agent module

Location:
- `Backend/internal/modules/agent/`

Responsibilities:
- read project files from the sandbox workspace
- build a context payload for the LLM
- send prompts to the configured AI model
- validate AI response structure
- apply returned file changes into the sandbox

Main file:
- `service.go`

Important behavior:
- `RunTask()` validates access, collects file context, calls the model, validates returned changes, and writes files into the sandbox.
- It does not directly push to GitHub; it modifies files inside the sandboxed repository workspace.

This is the key “AI coding loop” in the current implementation.

---

### 3.6 Email module

Location:
- `Backend/internal/shared/email/`

Responsibilities:
- send transactional emails in the background
- queue jobs for email sending
- worker pool consumes jobs

This is not the main business flow, but it is an important async infrastructure layer.

---

## 4. Current execution flow in plain English

The real implemented flow is:

1. User logs in through auth and receives a valid JWT.
2. User creates or selects a workspace.
3. User creates a project inside that workspace.
4. Project metadata is stored in PostgreSQL.
5. If the project includes a GitHub repo, that repo metadata is also stored.
6. The project lifecycle triggers a sandbox creation event.
7. The terminal module creates a sandbox and starts a container.
8. The runtime mounts a host directory into the container workspace.
9. The runtime clones the GitHub repository into the sandbox workspace.
10. The AI agent reads files from the sandbox workspace.
11. The AI model creates file-level changes.
12. The backend writes those changes back into the sandbox files.
13. The user sees the outcome through the app, but the repo is not yet being committed and pushed remotely in the current implementation.

This means the actual AI interaction is with the sandboxed workspace, not directly with the remote GitHub repo itself.

---

## 5. Background workers

There are worker patterns in the backend, but they are lightweight in-process workers, not distributed queue systems.

### 5.1 Terminal sandbox worker

Location:
- `Backend/internal/modules/terminal/worker/sandbox_worker.go`

This worker owns a buffered channel:
- `events chan ProjectEvent`

It works like this:
- `Publish(event)` adds a project event to the channel
- `process()` consumes events in goroutines
- `handleEvent()` dispatches by type
- `project.created` causes sandbox provisioning
- `project.deleted` causes sandbox cleanup

This is the core async background system for project state transitions.

### 5.2 Email worker queue

Location:
- `Backend/internal/shared/email/queue.go`
- `Backend/internal/shared/email/worker.go`
- `Backend/internal/shared/email/module.go`

This is a classic FIFO in-memory queue for async email sending.

Important note:
- this is not a distributed queue like Kafka or RabbitMQ
- it is a local Go-channel-based async queue

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
- sandbox creation
- Docker workspace mounting
- repo clone into workspace
- file listing, read, write
- AI coding task execution against sandbox files
- worker-based async event processing for project tasks

### Not fully implemented yet
- direct GitHub commit/push after AI edits
- branch sync to remote after code changes
- PR generation
- distributed queue infrastructure
- full production-grade async job system
- repo mutation flow outside the sandbox

---

## 9. The actual architecture in one sentence

ForgeAI’s current backend is a modular Go service where user and project metadata live in PostgreSQL, AI coding runs against Docker-based sandbox workspaces, and the backend manages the lifecycle of those sandboxes and repo clones while the model modifies files inside the sandbox rather than directly pushing to GitHub.

---

## 10. The most important takeaway

The central runtime concept is:

- project metadata in DB
- sandbox execution in Docker
- code repository cloned into sandbox
- AI agent edits files in sandbox
- GitHub is connected as repository metadata and source, but not as the active mutation target yet

This is the current real architecture in this codebase.

---

## 11. Detailed current execution flow

This section describes the execution path currently implemented in the backend. It separates request-time work from background work so it is clear when an HTTP request finishes and which operations continue asynchronously.

### 11.1 Server startup and dependency wiring

The composition root is `Backend/cmd/agent/server.go`. The server creates shared infrastructure first and then passes those dependencies into the modules that need them.

Startup sequence:

1. Load application configuration, including database, OAuth, JWT, email, Docker, and frontend settings.
2. Create the Gin router and shared logger.
3. Create the PostgreSQL connection when a database URL is configured.
4. Create the email module.
	- The module creates one in-memory queue with capacity 100.
	- The module creates one email worker pool with 3 goroutines by default.
	- The queue and service are injected into auth and workspace invite flows.
5. Create authentication dependencies, including the OAuth provider and JWT manager.
6. Load the auth and workspace modules.
7. Load the project module and its project, repository, GitHub, orchestrator, and sync services.
8. Load the terminal module.
	- The terminal module creates the Docker runtime, sandbox repository, application service, and sandbox worker.
	- The sandbox worker creates one in-memory event queue with capacity 100.
9. Register project lifecycle callbacks.
	- Project creation calls the terminal module's `OnProjectCreated` callback.
	- Project deletion calls the terminal module's `OnProjectDeleted` callback.
	- Branch changes call the terminal module's `OnProjectBranchUpdated` callback.
10. Start the email workers and the sandbox workers.
11. Register protected and public HTTP routes.
12. Start the Gin HTTP server.

The result is one Go process containing the HTTP server, the email queue, the sandbox event queue, six default worker goroutines, database clients, and external-system clients.

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
6. If a GitHub repository is selected, the repository layer stores the repository URL, branch, and access-token metadata.
7. The project service invokes its configured create callback.
8. The callback calls `terminalModule.OnProjectCreated`.
9. `OnProjectCreated` publishes a `project.created` event to the sandbox worker queue.
10. The project HTTP response can return before sandbox provisioning completes.
11. A sandbox worker receives the event and loads the project and repository metadata.
12. The worker asks the terminal application service to create the sandbox.
13. The Docker runtime creates the volume and container and mounts the project workspace.
14. The runtime clones the repository into the mounted workspace when repository metadata exists.
15. The worker provisions or refreshes the files and records the resulting sandbox state.

The important boundary is step 9: project persistence is request-time work, while sandbox provisioning is background work.

### 11.5 Project deletion flow

Deletion currently uses the terminal module as a cleanup boundary:

1. The project handler authenticates and authorizes the request.
2. The project service deletes or updates the project records through its repositories.
3. The configured delete callback calls `terminalModule.OnProjectDeleted`.
4. The terminal module delegates cleanup to the sandbox worker.
5. The worker stops or removes the Docker sandbox and associated workspace resources.

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

The database branch value is therefore only the source of truth for project metadata. Updating the database alone does not update the files already present in the Docker workspace; the terminal refresh callback is required for that second state transition.

### 11.7 AI coding task flow

The AI task flow is primarily synchronous from the API caller's perspective, but it operates on files that were prepared asynchronously during sandbox provisioning:

1. The client submits an AI task for a project.
2. The agent handler authenticates the request and validates the project context.
3. The agent service checks that the user can access the project and its sandbox.
4. The service asks the sandbox file API for the relevant files under `/workspace`.
5. The service builds a project context and prompt for the configured LLM service.
6. The LLM returns a structured response containing file changes.
7. The agent service validates the response format and rejects invalid or unsafe changes according to its validation rules.
8. The service writes accepted file changes through the sandbox file API.
9. The response reports the task result to the client.

The AI model does not edit the host repository directly and does not call GitHub directly. The effective target is the mounted Docker workspace. The backend is the orchestrator between the model and that workspace.

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

Unlike the email queue, the sandbox `Publish` method is intentionally non-blocking. This keeps the caller responsive, but it means a full queue can result in a dropped project event.

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
- Worker panics are recovered and logged by the sandbox worker process loop.
- Context cancellation stops workers.
- Queue contents are not recovered after a process restart.
- There is no general retry, dead-letter queue, durable job record, or distributed lock.

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
8. The remote GitHub repository remains unchanged until a future commit and push workflow is implemented.

The current architecture is therefore a sandbox-first editing system. GitHub supplies repository metadata and clone source, while Docker workspace state is the active working copy.

---

## 16. Current architecture limitations

The design is suitable for a single backend process and development or early-stage workloads, but the following limitations are important for production planning:

- In-memory queues do not survive restarts.
- A full sandbox queue can drop events.
- Worker counts are process-local and do not scale across instances.
- There is no durable task status model for every background operation.
- There is no distributed coordination for the same project across multiple backend instances.
- AI edits are not committed, pushed, or converted into pull requests.
- There is no durable audit trail for every file change produced by the model.

These are current-state observations, not missing pieces that are silently assumed to exist.
