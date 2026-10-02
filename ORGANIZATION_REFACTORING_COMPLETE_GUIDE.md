# ForgeAI: Complete Organization Refactoring Guide

## 📋 Table of Contents
1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Architecture Changes](#architecture-changes)
4. [Database Schema](#database-schema)
5. [Backend Implementation](#backend-implementation)
6. [Frontend Implementation](#frontend-implementation)
7. [Migration Scripts](#migration-scripts)
8. [API Reference](#api-reference)
9. [Testing Guide](#testing-guide)
10. [Troubleshooting](#troubleshooting)

---

## 🎯 Overview

### What's Changing?

**From**: Workspace-based architecture with GitHub sync
**To**: Organization-based architecture with manual project creation

### Key Changes
- ✅ **Organization** replaces Workspace (better multi-tenant model)
- ✅ **Organization Integration** (GitHub App at org level)
- ✅ **Manual Project Creation** (no auto-import/sync)
- ✅ **Optional Repository Linking** (link existing GitHub repos)
- ✅ **Simplified Sandbox** (combined container/volume tables)
- ❌ **Removed**: Repository sync service, GitHub catalog, import workflows

### Benefits
- **Scalable**: Organization-level GitHub App (enterprise ready)
- **Simple**: No background sync workers
- **Clear**: Organization → Project → Sandbox hierarchy
- **Fast**: Fewer database queries, no periodic API calls
- **Maintainable**: 11 tables instead of 15

---

## 🚀 Quick Start

### Prerequisites
- Go 1.26+
- Node.js 20+
- PostgreSQL 16+
- Docker (for sandboxes)

### 1. Run Automated Refactoring Script

```powershell
# Dry run first (see what will change)
.\refactor.ps1 -DryRun

# Apply changes
.\refactor.ps1
```

### 2. Reset Database

```powershell
# Drop old database
psql -U forgeai -d postgres -c "DROP DATABASE IF EXISTS forgeai;"
psql -U forgeai -d postgres -c "CREATE DATABASE forgeai;"

# Run migration
cd Backend
$env:GOOSE_DRIVER="postgres"
$env:GOOSE_DBSTRING="postgres://forgeai:password@localhost:5432/forgeai?sslmode=disable"
goose -dir migrations up
```

### 3. Build & Run Backend

```powershell
cd Backend
go mod tidy
go build -o ./bin/agent ./cmd/agent
./bin/agent
```

### 4. Build & Run Frontend

```powershell
cd Frontend
npm install
npm run dev
```

### 5. Test

Open browser: http://localhost:5173

---

## 🏗️ Architecture Changes

### Old Architecture (Workspace-Based)

```
User
 └─ Workspace (tenant boundary)
     ├─ Members (roles)
     ├─ Projects
     │   └─ Repository (GitHub, synced)
     └─ Sandbox (per project)

GitHub Integration: Per user OAuth
Background Workers: Sync metadata every 5 min
```

### New Architecture (Organization-Based)

```
User
 └─ Organization (tenant boundary)
     ├─ Members (roles)
     ├─ Integrations (GitHub App, per org)
     ├─ Projects (manual creation)
     │   └─ Repository (optional link)
     └─ Sandbox (per project)

GitHub Integration: Per org GitHub App
Background Workers: None (on-demand only)
```

### Terminology Mapping

| Old Term | New Term |
|----------|----------|
| Workspace | Organization |
| workspace_id | organization_id |
| WorkspaceID | OrganizationID |
| /workspaces | /organizations |
| WorkspaceMember | OrganizationMember |
| WorkspaceInvite | OrganizationInvite |

---

## 💾 Database Schema

### Tables Overview

```
1. tbl_user                         # User accounts
2. tbl_email_verification           # Email verification tokens
3. tbl_integration_provider         # Provider catalog (GitHub, GitLab, etc)
4. tbl_oauth_account                # User OAuth (GitHub/Google)
5. tbl_organization                 # Organizations
6. tbl_organization_member          # Members with roles
7. tbl_organization_invite          # Pending invites
8. tbl_organization_integration     # GitHub App installations
9. tbl_project                      # Projects (manual)
10. tbl_project_repository          # Optional repo links
11. tbl_sandbox                     # Sandboxes
12. tbl_sandbox_resource_limit      # Resource configs
```

### Key Relationships

```sql
-- Organization hierarchy
tbl_organization
  ├─ tbl_organization_member (user membership)
  ├─ tbl_organization_invite (pending invites)
  ├─ tbl_organization_integration (GitHub App)
  └─ tbl_project
       ├─ tbl_project_repository (optional)
       └─ tbl_sandbox
            └─ tbl_sandbox_resource_limit
```

### Access Control Flow

```
1. User authenticated → JWT with user_id
2. API request → Extract user_id from JWT
3. Validate access:
   - Organization: Check tbl_organization_member
   - Project: Get project → Check org membership
   - Sandbox: Get sandbox → Get project → Check org membership
```

---

## 🔧 Backend Implementation

### Module Structure

```
Backend/internal/modules/
├── organization/              # NEW MODULE
│   ├── core/
│   │   ├── model.go          # Organization entity
│   │   ├── repository.go     # CRUD operations
│   │   └── service.go        # Business logic
│   ├── member/
│   │   ├── model.go          # Member entity
│   │   ├── repository.go     # Member CRUD
│   │   └── service.go        # Member management
│   ├── invite/
│   │   ├── model.go          # Invite entity
│   │   ├── repository.go     # Invite CRUD
│   │   └── service.go        # Invite flow
│   ├── integration/
│   │   ├── model.go          # Integration entity
│   │   ├── repository.go     # Integration CRUD
│   │   ├── service.go        # Connect/disconnect
│   │   └── github_app.go     # GitHub App OAuth
│   ├── handler.go            # HTTP handlers
│   ├── route.go              # Route registration
│   └── module.go             # Module initialization
│
├── project/                   # SIMPLIFIED MODULE
│   ├── core/
│   │   ├── model.go          # Project entity
│   │   ├── repository.go     # Project CRUD
│   │   └── service.go        # Manual creation only
│   ├── repository/
│   │   ├── model.go          # ProjectRepository entity
│   │   ├── repository.go     # Link/unlink repo
│   │   └── service.go        # Repository linking
│   ├── handler.go
│   ├── route.go
│   └── module.go
│
├── terminal/                  # UPDATED MODULE
│   ├── application/
│   │   └── service.go        # Access control updated
│   ├── worker/
│   │   └── sandbox_worker.go # Removed sync events
│   ├── handler.go            # Organization-based access
│   └── module.go
│
└── [other modules unchanged]
```

### Key Models

**Organization**:
```go
type Organization struct {
    ID          string
    Name        string
    Slug        string
    Description string
    Settings    map[string]interface{}
    Status      string // ACTIVE, ARCHIVED
    CreatedBy   string
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   *time.Time
}
```

**OrganizationMember**:
```go
type OrganizationMember struct {
    ID             string
    OrganizationID string
    UserID         string
    Role           string // OWNER, ADMIN, MEMBER
    JoinedAt       time.Time
    CreatedAt      time.Time
    UpdatedAt      time.Time
    DeletedAt      *time.Time
}
```

**Project** (simplified):
```go
type Project struct {
    ID             string
    OrganizationID string
    Name           string
    Slug           string
    Description    string
    Status         string // ACTIVE, ARCHIVED
    CreatedBy      string
    CreatedAt      time.Time
    UpdatedAt      time.Time
    DeletedAt      *time.Time
    
    // Optional linked repository
    Repository *ProjectRepository
}
```

**ProjectRepository**:
```go
type ProjectRepository struct {
    ID                   string
    OrganizationID       string
    ProjectID            string
    IntegrationID        string
    ExternalRepositoryID string
    RepositoryName       string
    FullName             string
    DefaultBranch        string
    RepositoryURL        string
    IsPrivate            bool
    LinkedBy             string
    LinkedAt             time.Time
}
```

**Sandbox** (simplified):
```go
type Sandbox struct {
    ID            string
    ProjectID     string
    Status        string
    ContainerID   string
    ContainerName string
    VolumeName    string
    Image         string
    WorkspacePath string
    LastError     string
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DestroyedAt   *time.Time
    
    // Joined from tbl_sandbox_resource_limit
    CPUQuota       int
    MemoryLimitMB  int
    PIDsLimit      int
    StorageLimitMB int
}
```

### API Routes

**Old Routes (Removed)**:
```
POST   /workspaces/:id/projects/import
POST   /workspaces/:id/projects/sync
GET    /workspaces/:id/github/repositories
```

**New Routes**:
```
# Organizations
POST   /organizations
GET    /organizations
GET    /organizations/:id
PUT    /organizations/:id
DELETE /organizations/:id

# Members
GET    /organizations/:id/members
PUT    /organizations/:id/members/:member_id
DELETE /organizations/:id/members/:member_id

# Invites
POST   /organizations/:id/invites
GET    /organizations/:id/invites
DELETE /organizations/:id/invites/:invite_id
POST   /invites/accept

# Integrations (GitHub App)
GET    /organizations/:id/integrations
POST   /organizations/:id/integrations/github/connect
DELETE /organizations/:id/integrations/:integration_id
GET    /organizations/:id/integrations/:integration_id/repositories

# Projects (simplified)
POST   /organizations/:id/projects
GET    /organizations/:id/projects
GET    /projects/:id
PUT    /projects/:id
DELETE /projects/:id

# Repository Linking (NEW)
POST   /projects/:id/repository/link
DELETE /projects/:id/repository
PUT    /projects/:id/repository/branch

# Sandbox (unchanged paths, updated access control)
GET    /projects/:id/sandbox
GET    /sandboxes/:id/files
POST   /sandboxes/:id/files/save
GET    /sandboxes/:id/git/status
POST   /sandboxes/:id/git/commit
POST   /sandboxes/:id/git/push
```

### Server Wiring (cmd/agent/server.go)

**Remove**:
```go
// DELETE THIS
workspacecore.LoadModule(workspacecore.ModuleConfig{...})
```

**Add**:
```go
// Add organization module
organizationModule := organization.LoadModule(organization.ModuleConfig{
    Router:       protectedRouter,
    Database:     db,
    Logger:       appLogger,
    EmailService: emailModule.Service,
})

// Update project module (remove sync-related config)
projectModule := project.LoadModule(project.ModuleConfig{
    Router:        protectedRouter,
    Database:      db,
    Logger:        appLogger,
    // REMOVED: GitHubClient, GitHubAccount, SyncContext, SyncInterval
})

// Update terminal module (pass organization repo for access control)
terminalModule, err := terminal.LoadModule(terminal.ModuleConfig{
    Database:         db,
    Logger:           appLogger,
    DockerBinary:     "docker",
    PolicyPath:       settings.SandboxPolicyPath,
    ProjectRepo:      projectModule.CoreService,
    OrganizationRepo: organizationModule.CoreService, // NEW
    WebSocketOrigins: settings.CORSOrigins,
})
```

### Access Control Updates

**Old** (validate workspace membership):
```go
func ValidateWorkspaceAccess(ctx context.Context, userID, workspaceID string) error {
    // Check if user is member of workspace
}
```

**New** (validate organization membership):
```go
func ValidateOrganizationAccess(ctx context.Context, userID, organizationID string) error {
    member, err := memberRepo.FindByUserAndOrganization(ctx, userID, organizationID)
    if err != nil {
        return ErrAccessDenied
    }
    if member.DeletedAt != nil {
        return ErrAccessDenied
    }
    return nil
}

func ValidateProjectAccess(ctx context.Context, userID, projectID string) error {
    project, err := projectRepo.FindByID(ctx, projectID)
    if err != nil {
        return err
    }
    return ValidateOrganizationAccess(ctx, userID, project.OrganizationID)
}

func ValidateSandboxAccess(ctx context.Context, userID, sandboxID string) error {
    sandbox, err := sandboxRepo.FindByID(ctx, sandboxID)
    if err != nil {
        return err
    }
    return ValidateProjectAccess(ctx, userID, sandbox.ProjectID)
}
```

---

## 🎨 Frontend Implementation

### Directory Structure

**Before**:
```
Frontend/src/modules/workspace/
├── api/
│   ├── workspace.api.ts
│   ├── workspace-client.ts
│   └── github.api.ts         # DELETE
├── features/
│   ├── projects/
│   │   └── components/
│   │       ├── ImportDialog.tsx    # DELETE
│   │       └── SyncStatus.tsx      # DELETE
│   └── ...
└── hooks/
    └── useWorkspaceId.ts
```

**After**:
```
Frontend/src/modules/organization/
├── api/
│   ├── organization.api.ts
│   ├── organization-client.ts
│   └── integration.api.ts    # NEW (GitHub App)
├── features/
│   ├── projects/
│   │   └── components/
│   │       ├── CreateProjectDialog.tsx  # Simplified
│   │       └── LinkRepositoryDialog.tsx # NEW
│   └── ...
└── hooks/
    └── useOrganizationId.ts
```

### API Client Updates

**organization.api.ts**:
```typescript
export interface Organization {
  id: string;
  name: string;
  slug: string;
  description: string;
  status: 'ACTIVE' | 'ARCHIVED';
  created_by: string;
  created_at: string;
  updated_at: string;
}

export function createOrganization(
  accessToken: string,
  input: { name: string; description?: string }
) {
  return authenticatedRequest<Organization>(
    accessToken,
    '/organizations',
    { method: 'POST', body: JSON.stringify(input) }
  );
}

export function getOrganizations(accessToken: string) {
  return authenticatedRequest<Organization[]>(
    accessToken,
    '/organizations'
  );
}

export function getOrganization(accessToken: string, organizationId: string) {
  return authenticatedRequest<Organization>(
    accessToken,
    `/organizations/${organizationId}`
  );
}
```

### Routing Updates

**app/router.tsx**:
```typescript
// OLD
{
  path: "/workspace",
  element: <WorkspaceLayout />,
  children: [...]
}

// NEW
{
  path: "/organizations/:organizationId",
  element: <OrganizationLayout />,
  children: [
    { index: true, element: <OverviewPage /> },
    { path: "projects", element: <ProjectsPage /> },
    { path: "projects/:projectId", element: <ProjectDetailPage /> },
    { path: "members", element: <MembersPage /> },
    { path: "settings", element: <SettingsPage /> },
  ]
}
```

### Context Updates

**OrganizationContext.tsx**:
```typescript
interface OrganizationContextValue {
  organization: Organization | null;
  organizations: Organization[];
  isLoading: boolean;
  selectOrganization: (id: string) => void;
  refetch: () => void;
}

export function OrganizationProvider({ children }: { children: ReactNode }) {
  const [selectedId, setSelectedId] = useLocalStorage<string | null>(
    'forgeai_selected_organization',
    null
  );
  
  const { data: organizations, isLoading } = useQuery({
    queryKey: ['organizations'],
    queryFn: () => getOrganizations(accessToken),
  });
  
  const organization = organizations?.find(org => org.id === selectedId) ?? null;
  
  return (
    <OrganizationContext.Provider value={{
      organization,
      organizations: organizations ?? [],
      isLoading,
      selectOrganization: setSelectedId,
      refetch: () => queryClient.invalidateQueries(['organizations']),
    }}>
      {children}
    </OrganizationContext.Provider>
  );
}
```

### Component Updates

**CreateProjectDialog.tsx** (simplified):
```tsx
export function CreateProjectDialog({ organizationId, onClose }: Props) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [linkRepo, setLinkRepo] = useState(false);
  const [selectedRepo, setSelectedRepo] = useState<string | null>(null);
  
  // Fetch available repositories from integrations
  const { data: integrations } = useQuery({
    queryKey: ['integrations', organizationId],
    queryFn: () => getIntegrations(accessToken, organizationId),
    enabled: linkRepo,
  });
  
  const createMutation = useMutation({
    mutationFn: () => createProject(accessToken, organizationId, {
      name,
      description,
      ...(linkRepo && selectedRepo ? { repository_id: selectedRepo } : {})
    }),
    onSuccess: () => {
      queryClient.invalidateQueries(['projects', organizationId]);
      onClose();
    }
  });
  
  return (
    <Dialog>
      <input
        placeholder="Project name"
        value={name}
        onChange={e => setName(e.target.value)}
      />
      
      <textarea
        placeholder="Description (optional)"
        value={description}
        onChange={e => setDescription(e.target.value)}
      />
      
      <Checkbox
        checked={linkRepo}
        onChange={setLinkRepo}
        label="Link to GitHub repository"
      />
      
      {linkRepo && (
        <Select value={selectedRepo} onChange={setSelectedRepo}>
          <option value="">Select repository</option>
          {integrations?.flatMap(int => int.repositories).map(repo => (
            <option key={repo.id} value={repo.id}>
              {repo.full_name}
            </option>
          ))}
        </Select>
      )}
      
      <button onClick={() => createMutation.mutate()}>
        Create Project
      </button>
    </Dialog>
  );
}
```

### Improved Git Panel

**GitPanel.tsx**:
```tsx
export function GitPanel({ sandboxId, accessToken }: Props) {
  const [selectedFile, setSelectedFile] = useState<string | null>(null);
  const [commitMessage, setCommitMessage] = useState('');
  const [selectedFiles, setSelectedFiles] = useState<Set<string>>(new Set());
  
  const { data: status } = useQuery({
    queryKey: ['git-status', sandboxId],
    queryFn: () => getGitStatus(accessToken, sandboxId),
    refetchInterval: 5000,
  });
  
  const { data: diffs } = useQuery({
    queryKey: ['git-diffs', sandboxId],
    queryFn: () => getGitDiffs(accessToken, sandboxId),
    enabled: Boolean(status?.is_dirty),
  });
  
  const commitMutation = useMutation({
    mutationFn: () => gitCommit(accessToken, sandboxId, {
      message: commitMessage,
      files: Array.from(selectedFiles),
    }),
    onSuccess: () => {
      queryClient.invalidateQueries(['git-status']);
      setCommitMessage('');
      setSelectedFiles(new Set());
    },
  });
  
  const pushMutation = useMutation({
    mutationFn: () => gitPush(accessToken, sandboxId, {
      remote: 'origin',
      branch: status?.branch,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries(['git-status']);
    },
  });
  
  return (
    <div className="git-panel">
      {/* Branch Info */}
      <div className="branch-section">
        <BranchIcon />
        <span>{status?.branch}</span>
        {status?.ahead > 0 && <Badge>↑{status.ahead}</Badge>}
        {status?.behind > 0 && <Badge>↓{status.behind}</Badge>}
        <RefreshButton onClick={() => queryClient.invalidateQueries(['git-status'])} />
      </div>
      
      {/* File Groups */}
      <div className="files-section">
        {status?.staged.length > 0 && (
          <FileGroup
            label="Staged"
            files={status.staged}
            icon={<StagedIcon />}
            color="green"
            onFileClick={setSelectedFile}
            selectedFiles={selectedFiles}
            onToggleFile={(file) => {
              const next = new Set(selectedFiles);
              next.has(file) ? next.delete(file) : next.add(file);
              setSelectedFiles(next);
            }}
          />
        )}
        
        {status?.modified.length > 0 && (
          <FileGroup
            label="Modified"
            files={status.modified}
            icon={<ModifiedIcon />}
            color="orange"
            onFileClick={setSelectedFile}
            selectedFiles={selectedFiles}
            onToggleFile={(file) => {
              const next = new Set(selectedFiles);
              next.has(file) ? next.delete(file) : next.add(file);
              setSelectedFiles(next);
            }}
          />
        )}
        
        {status?.untracked.length > 0 && (
          <FileGroup
            label="Untracked"
            files={status.untracked}
            icon={<UntrackedIcon />}
            color="gray"
            onFileClick={setSelectedFile}
            selectedFiles={selectedFiles}
            onToggleFile={(file) => {
              const next = new Set(selectedFiles);
              next.has(file) ? next.delete(file) : next.add(file);
              setSelectedFiles(next);
            }}
          />
        )}
      </div>
      
      {/* Diff Viewer */}
      {selectedFile && diffs?.[selectedFile] && (
        <DiffViewer
          file={selectedFile}
          diff={diffs[selectedFile]}
          language={getLanguageFromPath(selectedFile)}
          onClose={() => setSelectedFile(null)}
          onStage={() => stageFile(selectedFile)}
          onUnstage={() => unstageFile(selectedFile)}
          onRevert={() => revertFile(selectedFile)}
        />
      )}
      
      {/* Commit Form */}
      {status?.is_dirty && (
        <div className="commit-section">
          <textarea
            placeholder="Commit message"
            value={commitMessage}
            onChange={e => setCommitMessage(e.target.value)}
            rows={3}
          />
          <div className="actions">
            <button
              onClick={() => {
                setSelectedFiles(new Set([
                  ...status.staged,
                  ...status.modified,
                  ...status.untracked,
                ]));
              }}
            >
              Select All
            </button>
            <button
              onClick={() => commitMutation.mutate()}
              disabled={!commitMessage.trim() || selectedFiles.size === 0}
            >
              Commit ({selectedFiles.size} files)
            </button>
          </div>
        </div>
      )}
      
      {/* Push Section */}
      {status?.ahead > 0 && (
        <div className="push-section">
          <p>{status.ahead} commit{status.ahead > 1 ? 's' : ''} ready to push</p>
          <button
            onClick={() => pushMutation.mutate()}
            disabled={pushMutation.isPending}
          >
            {pushMutation.isPending ? 'Pushing...' : `Push to ${status.branch}`}
          </button>
        </div>
      )}
    </div>
  );
}
```

---

## 🔄 Migration Scripts

### Automated PowerShell Script

Run the provided `refactor.ps1` script:

```powershell
# Preview changes
.\refactor.ps1 -DryRun

# Apply changes
.\refactor.ps1

# Skip backup (faster)
.\refactor.ps1 -BackupFirst:$false
```

### Manual Find-Replace Commands

**Backend**:
```powershell
cd Backend

# Replace variable names
Get-ChildItem -Path "./internal" -Filter "*.go" -Recurse | ForEach-Object {
    (Get-Content $_.FullName) `
        -replace 'workspace_id', 'organization_id' `
        -replace 'workspaceID', 'organizationID' `
        -replace 'WorkspaceID', 'OrganizationID' `
        -replace 'tbl_workspace ', 'tbl_organization ' `
        -replace 'tbl_workspace_member', 'tbl_organization_member' `
        -replace 'tbl_workspace_invite', 'tbl_organization_invite' `
        -replace '/workspaces', '/organizations' `
        | Set-Content $_.FullName
}
```

**Frontend**:
```powershell
cd Frontend

# Rename directory
Rename-Item -Path "src/modules/workspace" -NewName "organization"

# Replace imports and references
Get-ChildItem -Path "./src" -Include "*.ts","*.tsx" -Recurse | ForEach-Object {
    (Get-Content $_.FullName) `
        -replace 'workspaceId', 'organizationId' `
        -replace 'WorkspaceId', 'OrganizationId' `
        -replace 'workspace_id', 'organization_id' `
        -replace '/workspaces', '/organizations' `
        -replace '"workspace"', '"organization"' `
        -replace 'modules/workspace', 'modules/organization' `
        | Set-Content $_.FullName
}
```

---

## 📚 API Reference

### Organizations

**Create Organization**:
```http
POST /api/v1/organizations
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "My Company",
  "description": "Our organization"
}
```

**List Organizations**:
```http
GET /api/v1/organizations
Authorization: Bearer {token}
```

**Get Organization**:
```http
GET /api/v1/organizations/{id}
Authorization: Bearer {token}
```

### Projects

**Create Project** (manual, no repo):
```http
POST /api/v1/organizations/{org_id}/projects
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "My Project",
  "description": "A new project"
}
```

**Link Repository** (optional):
```http
POST /api/v1/projects/{project_id}/repository/link
Authorization: Bearer {token}
Content-Type: application/json

{
  "integration_id": "integration-uuid",
  "external_repository_id": "123456",
  "repository_name": "my-repo",
  "full_name": "username/my-repo",
  "default_branch": "main",
  "repository_url": "https://github.com/username/my-repo",
  "is_private": false
}
```

### Sandbox

**Get/Ensure Sandbox**:
```http
GET /api/v1/projects/{project_id}/sandbox?ensure=true
Authorization: Bearer {token}
```

**Git Status**:
```http
GET /api/v1/sandboxes/{sandbox_id}/git/status
Authorization: Bearer {token}
```

**Git Commit**:
```http
POST /api/v1/sandboxes/{sandbox_id}/git/commit
Authorization: Bearer {token}
Content-Type: application/json

{
  "message": "Fix bug",
  "files": ["src/main.ts", "src/utils.ts"]
}
```

---

## 🧪 Testing Guide

### Backend Tests

```powershell
cd Backend

# Unit tests
go test ./internal/modules/organization/...
go test ./internal/modules/project/...

# Integration tests
go test -tags=integration ./...

# Build
go build -o ./bin/agent ./cmd/agent
```

### Frontend Tests

```powershell
cd Frontend

# Type check
npm run tsc --noEmit

# Lint
npm run lint

# Build
npm run build
```

### Manual Testing Checklist

```
Backend:
[ ] Create organization
[ ] List organizations
[ ] Add member
[ ] Send invite
[ ] Accept invite
[ ] Create project
[ ] Link repository
[ ] Get sandbox (ensure=true)
[ ] List files
[ ] Save file
[ ] Git status
[ ] Git commit
[ ] Git push

Frontend:
[ ] Login works
[ ] Organizations list loads
[ ] Can create organization
[ ] Can switch organizations
[ ] Can create project
[ ] Project detail page loads
[ ] File explorer works
[ ] Monaco editor loads
[ ] Can edit and save files
[ ] Git panel shows status
[ ] Can commit changes
[ ] Can push to remote
[ ] Agent panel works
[ ] Terminal works
```

---

## 🔧 Troubleshooting

### Common Issues

**1. Import errors after refactoring**

```powershell
# Backend
cd Backend
go mod tidy
go clean -cache

# Frontend
cd Frontend
rm -rf node_modules package-lock.json
npm install
```

**2. Database foreign key violations**

```sql
-- Check constraints
SELECT * FROM information_schema.table_constraints
WHERE table_schema = 'public';

-- Fix data
-- Ensure all organization_id values exist in tbl_organization
```

**3. Frontend 404 errors**

- Check `app/router.tsx` paths
- Verify API URL in `.env` (VITE_API_URL)
- Check browser console for errors
- Use Network tab to see actual requests

**4. Access denied errors**

- Verify user is member of organization
- Check token is valid (not expired)
- Verify middleware is applied to routes
- Check access control logic

**5. Git panel shows wrong files**

- Verify git status parsing
- Check file path normalization
- Test git commands manually in sandbox:
  ```bash
  docker exec forgeai-cnt-{project-id} git status --short
  ```

### Debug Commands

**Backend logs**:
```powershell
# Set debug log level
$env:LOG_LEVEL="debug"
./bin/agent
```

**Database queries**:
```sql
-- Check organization membership
SELECT * FROM tbl_organization_member
WHERE user_id = 'your-user-id';

-- Check projects
SELECT p.*, pr.*
FROM tbl_project p
LEFT JOIN tbl_project_repository pr ON p.id = pr.project_id
WHERE p.organization_id = 'your-org-id';

-- Check sandboxes
SELECT * FROM tbl_sandbox
WHERE project_id = 'your-project-id';
```

**API testing**:
```powershell
# Test organization creation
$response = Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/organizations" `
  -Method POST `
  -Headers @{"Authorization"="Bearer $TOKEN"; "Content-Type"="application/json"} `
  -Body '{"name":"Test Org"}' `
  -Verbose

$response | ConvertTo-Json -Depth 10
```

---

## ✅ Success Criteria

Your refactoring is complete when:

1. ✅ New migration runs successfully
2. ✅ Backend builds without errors
3. ✅ Frontend builds without errors
4. ✅ All API endpoints work
5. ✅ Can create organization
6. ✅ Can create project
7. ✅ Can provision sandbox
8. ✅ Can edit files
9. ✅ Git operations work
10. ✅ Agent tasks work
11. ✅ No "workspace" terminology remains
12. ✅ Code is cleaner and more maintainable

---

## 📞 Support

If you encounter issues:

1. Check this guide's Troubleshooting section
2. Review backend logs (set LOG_LEVEL=debug)
3. Check browser console for frontend errors
4. Test API endpoints with curl/Postman
5. Verify database schema matches migration
6. Check Docker containers are running

---

## 🎉 Next Steps

After successful refactoring:

1. **Code Review**: Review all changes
2. **Git Commit**: Commit changes with clear message
3. **Documentation**: Update any additional docs
4. **Testing**: Run full test suite
5. **Staging Deploy**: Deploy to staging environment
6. **QA Testing**: Thorough QA testing
7. **Production Deploy**: Deploy to production
8. **Monitor**: Watch for errors and performance

---

**Last Updated**: 2027-01-01
**Version**: 2.0.0
