# ForgeAI: Automated Workspace → Organization Refactoring Script
# PowerShell version for Windows

param(
    [switch]$DryRun = $false,
    [switch]$BackupFirst = $true
)

$ErrorActionPreference = "Stop"

Write-Host "=============================================" -ForegroundColor Cyan
Write-Host "ForgeAI: Organization Refactoring Script" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host ""

if ($DryRun) {
    Write-Host "[DRY RUN MODE] No files will be modified" -ForegroundColor Yellow
    Write-Host ""
}

# ========================================
# 1. BACKUP
# ========================================

if ($BackupFirst -and -not $DryRun) {
    Write-Host "[1/7] Creating backup..." -ForegroundColor Green
    $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
    $backupDir = "backup_$timestamp"
    
    if (-not (Test-Path $backupDir)) {
        New-Item -ItemType Directory -Path $backupDir | Out-Null
    }
    
    Copy-Item -Path "Backend" -Destination "$backupDir/Backend" -Recurse
    Copy-Item -Path "Frontend" -Destination "$backupDir/Frontend" -Recurse
    
    Write-Host "  ✓ Backup created: $backupDir" -ForegroundColor Green
} else {
    Write-Host "[1/7] Skipping backup..." -ForegroundColor Yellow
}

Write-Host ""

# ========================================
# 2. BACKEND: Delete old modules
# ========================================

Write-Host "[2/7] Cleaning up old backend modules..." -ForegroundColor Green

$modulesToDelete = @(
    "Backend/internal/modules/workspaces",
    "Backend/internal/modules/project/sync",
    "Backend/internal/modules/project/github/catalog.go",
    "Backend/internal/modules/project/orchestrator/sync_orchestrator.go",
    "Backend/internal/modules/project/orchestrator/github_orchestrator.go"
)

foreach ($module in $modulesToDelete) {
    if (Test-Path $module) {
        if ($DryRun) {
            Write-Host "  [DRY RUN] Would delete: $module" -ForegroundColor Yellow
        } else {
            Remove-Item -Path $module -Recurse -Force
            Write-Host "  ✓ Deleted: $module" -ForegroundColor Green
        }
    } else {
        Write-Host "  - Not found: $module" -ForegroundColor DarkGray
    }
}

Write-Host ""

# ========================================
# 3. BACKEND: Find-Replace
# ========================================

Write-Host "[3/7] Updating backend code..." -ForegroundColor Green

$backendReplacements = @{
    'workspace_id' = 'organization_id'
    'workspaceID' = 'organizationID'
    'WorkspaceID' = 'OrganizationID'
    'tbl_workspace ' = 'tbl_organization '
    'tbl_workspace_member' = 'tbl_organization_member'
    'tbl_workspace_invite' = 'tbl_organization_invite'
    '/workspaces' = '/organizations'
    'workspaces/workspace' = 'organization/core'
    'workspaces/member' = 'organization/member'
    'workspaces/invite' = 'organization/invite'
}

$backendFiles = Get-ChildItem -Path "Backend/internal" -Filter "*.go" -Recurse
$totalFiles = $backendFiles.Count
$processed = 0

foreach ($file in $backendFiles) {
    $processed++
    $progress = [math]::Round(($processed / $totalFiles) * 100)
    Write-Progress -Activity "Processing backend files" -Status "$processed of $totalFiles" -PercentComplete $progress
    
    if ($DryRun) {
        continue
    }
    
    $content = Get-Content -Path $file.FullName -Raw
    $modified = $false
    
    foreach ($key in $backendReplacements.Keys) {
        if ($content -match [regex]::Escape($key)) {
            $content = $content -replace [regex]::Escape($key), $backendReplacements[$key]
            $modified = $true
        }
    }
    
    # Special replacements for type definitions
    if ($content -match 'type Workspace ') {
        $content = $content -replace 'type Workspace ', 'type Organization '
        $modified = $true
    }
    
    if ($content -match '\*Workspace') {
        $content = $content -replace '\*Workspace', '*Organization'
        $modified = $true
    }
    
    if ($content -match '\[\]Workspace') {
        $content = $content -replace '\[\]Workspace', '[]Organization'
        $modified = $true
    }
    
    if ($modified) {
        Set-Content -Path $file.FullName -Value $content -NoNewline
    }
}

Write-Progress -Activity "Processing backend files" -Completed
Write-Host "  ✓ Updated $totalFiles backend files" -ForegroundColor Green
Write-Host ""

# ========================================
# 4. FRONTEND: Rename modules
# ========================================

Write-Host "[4/7] Renaming frontend modules..." -ForegroundColor Green

if (Test-Path "Frontend/src/modules/workspace") {
    if ($DryRun) {
        Write-Host "  [DRY RUN] Would rename: workspace → organization" -ForegroundColor Yellow
    } else {
        Rename-Item -Path "Frontend/src/modules/workspace" -NewName "organization"
        Write-Host "  ✓ Renamed: workspace → organization" -ForegroundColor Green
    }
} else {
    Write-Host "  - Module not found: Frontend/src/modules/workspace" -ForegroundColor DarkGray
}

Write-Host ""

# ========================================
# 5. FRONTEND: Find-Replace
# ========================================

Write-Host "[5/7] Updating frontend code..." -ForegroundColor Green

$frontendReplacements = @{
    'workspaceId' = 'organizationId'
    'WorkspaceId' = 'OrganizationId'
    'workspace_id' = 'organization_id'
    '/workspaces' = '/organizations'
    '"workspace"' = '"organization"'
    'modules/workspace' = 'modules/organization'
}

$frontendFiles = Get-ChildItem -Path "Frontend/src" -Include "*.ts","*.tsx" -Recurse
$totalFiles = $frontendFiles.Count
$processed = 0

foreach ($file in $frontendFiles) {
    $processed++
    $progress = [math]::Round(($processed / $totalFiles) * 100)
    Write-Progress -Activity "Processing frontend files" -Status "$processed of $totalFiles" -PercentComplete $progress
    
    if ($DryRun) {
        continue
    }
    
    $content = Get-Content -Path $file.FullName -Raw
    $modified = $false
    
    foreach ($key in $frontendReplacements.Keys) {
        if ($content -match [regex]::Escape($key)) {
            $content = $content -replace [regex]::Escape($key), $frontendReplacements[$key]
            $modified = $true
        }
    }
    
    # Special replacements for types
    if ($content -match ': Workspace[^A-Za-z]') {
        $content = $content -replace ': Workspace([^A-Za-z])', ': Organization$1'
        $modified = $true
    }
    
    if ($content -match '<Workspace>') {
        $content = $content -replace '<Workspace>', '<Organization>'
        $modified = $true
    }
    
    if ($content -match 'interface Workspace ') {
        $content = $content -replace 'interface Workspace ', 'interface Organization '
        $modified = $true
    }
    
    if ($modified) {
        Set-Content -Path $file.FullName -Value $content -NoNewline
    }
}

Write-Progress -Activity "Processing frontend files" -Completed
Write-Host "  ✓ Updated $totalFiles frontend files" -ForegroundColor Green
Write-Host ""

# ========================================
# 6. FRONTEND: Delete import flows
# ========================================

Write-Host "[6/7] Removing repository import flows..." -ForegroundColor Green

$componentsToDelete = @(
    "Frontend/src/modules/organization/features/projects/components/ImportRepositoriesDialog.tsx",
    "Frontend/src/modules/organization/features/projects/components/GitHubRepositoryList.tsx",
    "Frontend/src/modules/organization/features/projects/components/SyncStatus.tsx",
    "Frontend/src/modules/organization/api/github.api.ts"
)

foreach ($component in $componentsToDelete) {
    if (Test-Path $component) {
        if ($DryRun) {
            Write-Host "  [DRY RUN] Would delete: $component" -ForegroundColor Yellow
        } else {
            Remove-Item -Path $component -Force
            Write-Host "  ✓ Deleted: $component" -ForegroundColor Green
        }
    } else {
        Write-Host "  - Not found: $component" -ForegroundColor DarkGray
    }
}

Write-Host ""

# ========================================
# 7. SUMMARY
# ========================================

Write-Host "[7/7] Refactoring complete!" -ForegroundColor Green
Write-Host ""
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host "Next Steps:" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "1. Review the changes:" -ForegroundColor White
Write-Host "   git diff" -ForegroundColor DarkGray
Write-Host ""
Write-Host "2. Create organization module files:" -ForegroundColor White
Write-Host "   - See REFACTORING_GUIDE.md for structure" -ForegroundColor DarkGray
Write-Host ""
Write-Host "3. Update server.go:" -ForegroundColor White
Write-Host "   - Load organization module" -ForegroundColor DarkGray
Write-Host "   - Remove workspace module" -ForegroundColor DarkGray
Write-Host ""
Write-Host "4. Reset database:" -ForegroundColor White
Write-Host "   cd Backend" -ForegroundColor DarkGray
Write-Host "   # Drop and recreate database" -ForegroundColor DarkGray
Write-Host "   goose postgres 'connection-string' up" -ForegroundColor DarkGray
Write-Host ""
Write-Host "5. Test backend:" -ForegroundColor White
Write-Host "   cd Backend" -ForegroundColor DarkGray
Write-Host "   go build -o ./bin/agent ./cmd/agent" -ForegroundColor DarkGray
Write-Host "   ./bin/agent" -ForegroundColor DarkGray
Write-Host ""
Write-Host "6. Test frontend:" -ForegroundColor White
Write-Host "   cd Frontend" -ForegroundColor DarkGray
Write-Host "   npm run build" -ForegroundColor DarkGray
Write-Host "   npm run dev" -ForegroundColor DarkGray
Write-Host ""
Write-Host "7. Verify all features work:" -ForegroundColor White
Write-Host "   - Organization CRUD" -ForegroundColor DarkGray
Write-Host "   - Member management" -ForegroundColor DarkGray
Write-Host "   - Project creation" -ForegroundColor DarkGray
Write-Host "   - Sandbox provisioning" -ForegroundColor DarkGray
Write-Host "   - Git operations" -ForegroundColor DarkGray
Write-Host "   - Agent tasks" -ForegroundColor DarkGray
Write-Host ""
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host ""

if ($DryRun) {
    Write-Host "NOTE: This was a DRY RUN. No files were modified." -ForegroundColor Yellow
    Write-Host "Run without -DryRun flag to apply changes." -ForegroundColor Yellow
    Write-Host ""
}
