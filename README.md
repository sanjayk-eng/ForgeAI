# ForgeAI - AI-Powered Development Workspace

## 🚀 Quick Links

- **[Complete Refactoring Guide](./ORGANIZATION_REFACTORING_COMPLETE_GUIDE.md)** - Full guide for workspace → organization migration
- **[Migration Script](./refactor.ps1)** - Automated refactoring script

---

## 📋 What's New

### Version 2.0 - Organization-Based Architecture

ForgeAI has been refactored from a workspace-based to an organization-based architecture:

**Key Changes**:
- ✅ Organizations replace Workspaces (better multi-tenant model)
- ✅ GitHub App integration at organization level
- ✅ Manual project creation (no auto-import)
- ✅ Optional repository linking
- ✅ Simplified sandbox model
- ❌ Removed: Background sync workers, repository import

**Benefits**:
- More scalable (enterprise-ready)
- Simpler codebase (11 vs 15 tables)
- Faster (no background workers)
- Clearer architecture

---

## 🏗️ Architecture Overview

```
User
 └─ Organization (multi-tenant boundary)
     ├─ Members (OWNER, ADMIN, MEMBER roles)
     ├─ Integrations (GitHub App per organization)
     ├─ Projects (manual creation)
     │   ├─ Repository (optional link to GitHub)
     │   └─ Sandbox (Docker-based isolated environment)
     └─ Settings
```

---

## 🚀 Getting Started

### Prerequisites
- Go 1.26+
- Node.js 20+
- PostgreSQL 16+
- Docker (for sandboxes)

### Quick Start

1. **Run automated refactoring**:
   ```powershell
   .\refactor.ps1
   ```

2. **Setup database**:
   ```powershell
   cd Backend
   goose -dir migrations postgres "connection-string" up
   ```

3. **Start backend**:
   ```powershell
   cd Backend
   go build -o ./bin/agent ./cmd/agent
   ./bin/agent
   ```

4. **Start frontend**:
   ```powershell
   cd Frontend
   npm install
   npm run dev
   ```

5. **Open browser**: http://localhost:5173

For detailed instructions, see [ORGANIZATION_REFACTORING_COMPLETE_GUIDE.md](./ORGANIZATION_REFACTORING_COMPLETE_GUIDE.md)

---

## 📁 Project Structure

```
ForgeAI/
├── Backend/
│   ├── cmd/agent/                 # Main entry point
│   ├── internal/
│   │   ├── config/                # Configuration
│   │   ├── middleware/            # HTTP middleware
│   │   ├── modules/
│   │   │   ├── organization/      # NEW: Organization module
│   │   │   ├── project/           # Simplified project module
│   │   │   ├── terminal/          # Sandbox management
│   │   │   ├── agent/             # AI agent
│   │   │   ├── git/               # Git operations
│   │   │   └── auth/              # Authentication
│   │   └── shared/                # Shared utilities
│   ├── migrations/                # Database migrations
│   └── go.mod
│
├── Frontend/
│   ├── src/
│   │   ├── app/                   # App initialization
│   │   ├── modules/
│   │   │   ├── organization/      # NEW: Organization features
│   │   │   └── auth/              # Authentication
│   │   └── shared/                # Shared components
│   └── package.json
│
├── ORGANIZATION_REFACTORING_COMPLETE_GUIDE.md  # Main guide
└── refactor.ps1                                 # Automated script
```

---

## 🔧 Technology Stack

### Backend
- **Language**: Go 1.26
- **Web Framework**: Gin
- **Database**: PostgreSQL 16 with sqlx
- **Migrations**: Goose
- **Authentication**: JWT (HS256)
- **OAuth**: GitHub & Google
- **Containers**: Docker
- **Logging**: Uber Zap

### Frontend
- **Framework**: React 19
- **Language**: TypeScript
- **Build Tool**: Vite
- **Routing**: React Router v7
- **State**: TanStack Query + Zustand
- **Styling**: TailwindCSS 4
- **Editor**: Monaco Editor
- **Terminal**: xterm.js

---

## 📖 Key Features

### 🏢 Organizations
- Multi-tenant workspace
- Role-based access (Owner, Admin, Member)
- Team collaboration
- Member invitations

### 🔗 GitHub Integration
- GitHub App installation (org-level)
- Repository linking
- OAuth authentication
- Secure git operations

### 📁 Projects
- Manual project creation
- Optional repository linking
- Project management
- Sandbox provisioning

### 🖥️ Sandboxes
- Docker-based isolation
- File operations (CRUD)
- Terminal access (WebSocket)
- Resource limits
- Network isolation

### 🔄 Git Operations
- Status tracking
- File staging
- Commits with author
- Push to remote (OAuth)
- Diff viewing
- Revert changes

### 🤖 AI Agent
- Code generation
- File modifications
- Context-aware
- Real-time progress
- LLM integration (OpenAI-compatible)

### 📡 Real-time Features
- WebSocket events
- File change notifications
- Git status updates
- Agent progress
- Terminal I/O

---

## 🔐 Security

- **Authentication**: JWT with refresh tokens
- **Authorization**: Role-based access control
- **Sandbox Isolation**: 
  - No network access
  - Read-only root filesystem
  - Resource limits (CPU, memory, PIDs)
  - Dropped capabilities
- **OAuth Tokens**: Stored encrypted, passed via stdin
- **Input Validation**: All endpoints validated
- **CORS**: Configurable origins

---

## 📊 Database Schema

### Core Tables
1. `tbl_user` - User accounts
2. `tbl_organization` - Organizations
3. `tbl_organization_member` - Members with roles
4. `tbl_organization_invite` - Pending invites
5. `tbl_organization_integration` - GitHub App installations
6. `tbl_project` - Projects
7. `tbl_project_repository` - Repository links
8. `tbl_sandbox` - Sandboxes
9. `tbl_sandbox_resource_limit` - Resource configs

See migration file: `Backend/migrations/20270101000000_initial_schema.sql`

---

## 🧪 Testing

### Backend
```powershell
cd Backend
go test ./internal/modules/...
go build -o ./bin/agent ./cmd/agent
```

### Frontend
```powershell
cd Frontend
npm run lint
npm run build
npm run dev
```

### API Testing
```powershell
# See API reference in ORGANIZATION_REFACTORING_COMPLETE_GUIDE.md
$TOKEN = "your-jwt-token"
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/organizations" `
  -Method POST `
  -Headers @{"Authorization"="Bearer $TOKEN"; "Content-Type"="application/json"} `
  -Body '{"name":"My Org"}'
```

---

## 🐛 Troubleshooting

See the [Troubleshooting section](./ORGANIZATION_REFACTORING_COMPLETE_GUIDE.md#troubleshooting) in the complete guide.

Common issues:
- Import errors → Run `go mod tidy` / `npm install`
- Database errors → Check schema matches migration
- 404 errors → Verify routes updated
- Access denied → Check organization membership
- Git errors → Test commands manually in sandbox

---

## 📚 Documentation

- **[Complete Refactoring Guide](./ORGANIZATION_REFACTORING_COMPLETE_GUIDE.md)** - Full implementation guide
  - Architecture changes
  - Database schema
  - Backend implementation
  - Frontend implementation
  - Migration scripts
  - API reference
  - Testing guide
  - Troubleshooting

- **[Architecture Summary](./ARCHITECTURE_SUMMARY.md)** - Original architecture (outdated)

---

## 🤝 Contributing

1. Read the refactoring guide
2. Follow existing code patterns
3. Write tests for new features
4. Update documentation
5. Submit pull request

---

## 📄 License

[Your License Here]

---

## 🆘 Support

For issues during refactoring:

1. Check [ORGANIZATION_REFACTORING_COMPLETE_GUIDE.md](./ORGANIZATION_REFACTORING_COMPLETE_GUIDE.md)
2. Review backend logs (set `LOG_LEVEL=debug`)
3. Check browser console
4. Test API endpoints individually
5. Verify database schema

---

## 🎯 Roadmap

### Completed
- ✅ Organization-based architecture
- ✅ Manual project creation
- ✅ Optional repository linking
- ✅ Simplified sandbox model
- ✅ GitHub App integration

### Planned
- [ ] Multiple sandboxes per project
- [ ] Sandbox templates (Node, Python, Rust, etc.)
- [ ] Branch switching UI
- [ ] File upload (drag-and-drop)
- [ ] Real-time collaboration
- [ ] Terminal multiplexing
- [ ] Preview service (port forwarding)
- [ ] Audit logging

---

**Version**: 2.0.0  
**Last Updated**: 2027-01-01
