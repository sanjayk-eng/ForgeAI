# Git Diff in Docker and Real-Time Display

This document explains how this project performs Git diff inside the Docker sandbox and how those changes are surfaced to the frontend in near real time.

## 1. Architecture overview

The project does not diff directly against the remote GitHub repo. Instead, it does the work inside a sandboxed workspace running in Docker.

The important flow is:

1. A project creates a sandbox container.
2. The project workspace is mounted into the container.
3. The repo is cloned into `/workspace` inside that container.
4. The AI agent or user edits files inside the sandbox.
5. Git status and git diff are computed inside that workspace.
6. The backend exposes those values through protected API routes.
7. The frontend refreshes the diff and file tree in real time.

This is the actual architecture used by the backend in the project.

---

## 2. Where the project does Git diff

The Git logic is handled by the backend module at:

- `Backend/internal/modules/git/`

The key functions are centered around:

- Git status
- Git diff
- Commit
- Push

The module does not directly modify the remote repo. It works against the sandbox workspace, then pushes to GitHub using the authenticated user token when requested.

---

## 3. Docker sandbox model

The configured sandbox image is:

```yaml
sandbox:
  image: alpine/git:latest
  git_image: alpine/git:latest
```

This is defined in:

- `Backend/configs/sandbox.yaml`

In practice:

- the main sandbox runs with `network_mode: none`
- the workspace is mounted on the host path and exposed inside the container at `/workspace`
- Git operations happen inside the sandbox filesystem
- a separate short-lived helper container may be used for networked GitHub operations such as push

This means the diff is computed locally inside the Docker workspace, not by calling the GitHub API.

---

## 4. How Git diff is computed in Docker

Inside the sandbox, the repo is located under the workspace directory, usually:

```bash
/workspace
```

Typical commands used in the project pattern are:

```bash
git -C /workspace status
git -C /workspace diff -- .
git -C /workspace diff HEAD
git -C /workspace diff --stat
```

The project’s Git module uses the same pattern: it reads the git state inside the sandbox and returns it to the caller.

### Example local Docker command

```bash
docker run --rm -v "$PWD":/workspace -w /workspace alpine/git:latest sh -lc "git init && echo 'hello' > test.txt && git add test.txt && git diff --cached" 
```

This shows the same concept: the repo is inside a mounted workspace, and the diff is computed inside the container.

---

## 5. Why this is done in Docker

This protects the project from direct host-level mutation and keeps the working copy isolated.

Benefits:

- each project gets its own working directory
- AI edits stay inside a controlled environment
- Git operations can be validated before pushing
- users can inspect file changes without editing the host repo directly

---

## 6. How real-time diff updates are shown

The project uses filesystem watchers and project-scoped realtime events.

### The real-time flow

1. The sandbox worker starts a filesystem watcher on the bound workspace.
2. File create, change, delete, and rename events are captured.
3. Paths are normalized to project-relative paths.
4. The project event hub sends updates to subscribed clients.
5. The frontend refreshes the relevant file tree and Git status/diff.

This is the critical fact: the real-time update is not a direct Git watch in the browser. It is a file-system event pipeline from the sandbox to the frontend.

---

## 7. The file-watching pattern

The watcher monitors working tree changes under the mounted host directory.

Typical design idea:

- watch `/workspace`
- filter out generated folders like `.git`, `node_modules`, `dist`, `build`
- normalize file paths
- coalesce rapid changes
- notify project subscribers

This is how the app shows file changes appearing in the UI without requiring a full page reload.

---

## 8. How the frontend gets updates

The frontend subscribes to project events and refreshes the relevant file and Git state through APIs.

Typical sequence:

1. browser opens a websocket for the project
2. backend emits file change or git-state update events
3. frontend updates file tree and selected file
4. frontend calls Git status/diff APIs if needed
5. diff panel refreshes with the latest content

This allows the UI to show near real-time repository changes as the sandbox updates.

---

## 9. Example of the runtime logic

A practical sequence looks like this:

```text
User edits file in sandbox
        ↓
fsnotify detects file change
        ↓
project event is published
        ↓
backend notifies websocket subscribers
        ↓
frontend refreshes file tree / diff panel
        ↓
user sees updated code changes immediately
```

---

## 10. Git diff display in the app

The app usually performs these operations when showing the diff:

```bash
git -C /workspace status
git -C /workspace diff HEAD
```

Then the backend serializes the result as a structured response for the frontend.

The UI can then show:

- changed files
- added lines
- removed lines
- commit-ready patch state

---

## 11. Important security boundary

The Docker sandbox is isolated and usually uses network restrictions.

This matters because:

- code edits happen inside the sandbox
- GitHub access is only used for authenticated push operations
- the diff is computed locally before any push is attempted

This prevents direct model-to-GitHub editing and keeps the user’s repo changes visible and reviewable before push.

---

## 12. Recommended flow for debugging diff issues

If you want to debug a Git diff problem in Docker:

```bash
docker ps
docker exec -it <container_name> sh
cd /workspace
git status
git diff
git diff HEAD
```

If you are debugging the host-side watch flow, inspect:

- sandbox worker lifecycle
- file watcher events
- project realtime hub
- frontend refresh calls

---

## 13. Summary

The project’s real workflow is:

- workspace files live in Docker sandbox
- Git diff is computed there
- filesystem watchers detect changes
- realtime project events notify the frontend
- the UI refreshes files and diff panels without reloading the whole app

So the answer is:

Git diff happens inside the Docker workspace, and the “real-time” experience is produced by file-watching + websocket event propagation, not by GitHub itself.

---

## 14. Practical command set

```bash
# inspect running containers
docker ps

# inspect sandbox workspace
docker exec -it <sandbox_container> sh
cd /workspace
git status
git diff
git diff --stat

# show latest tracked changes
git diff HEAD
```

This gives a quick way to confirm that the changes are being tracked in the Docker environment.
