package docker

import (
	"errors"
	"strings"
	"testing"
)

func TestParseContainerID(t *testing.T) {
	validID := strings.Repeat("a", 64)
	tests := []struct {
		name    string
		output  string
		wantID  string
		wantErr bool
	}{
		{name: "trims Docker output", output: validID + "\n", wantID: validID},
		{name: "rejects empty output", wantErr: true},
		{name: "rejects multiple lines", output: validID + "\nprogress output", wantErr: true},
		{name: "rejects values exceeding the database column", output: strings.Repeat("a", 256), wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseContainerID(test.output)
			if (err != nil) != test.wantErr {
				t.Fatalf("parseContainerID() error = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.wantID {
				t.Fatalf("parseContainerID() = %q, want %q", got, test.wantID)
			}
		})
	}
}

func TestParseTerminalShells(t *testing.T) {
	shells := parseTerminalShells("sh\t/bin/sh\nbash\t/usr/bin/bash\npwsh\t/usr/bin/pwsh\npowershell\t\n")
	if len(shells) != 3 {
		t.Fatalf("detected %d shells, want 3: %+v", len(shells), shells)
	}
	if shells[0].ID != "sh" || !shells[0].Default || shells[0].Executable != "/bin/sh" {
		t.Fatalf("unexpected default shell: %+v", shells[0])
	}
	if shells[1].ID != "bash" || shells[1].Default {
		t.Fatalf("unexpected Bash shell: %+v", shells[1])
	}
	if shells[2].ID != "powershell" || shells[2].Executable != "/usr/bin/pwsh" {
		t.Fatalf("unexpected PowerShell shell: %+v", shells[2])
	}
	if shells := parseTerminalShells("bash\t\nzsh\t\n"); len(shells) != 0 {
		t.Fatalf("unavailable shells were reported: %+v", shells)
	}
}

func TestIsMissingDockerResource(t *testing.T) {
	if !isMissingDockerResource(errors.New("Error response from daemon: No such container: sandbox"), "container") {
		t.Fatal("expected missing container to be treated as already removed")
	}
	if isMissingDockerResource(errors.New("permission denied"), "container") {
		t.Fatal("permission errors must not be ignored")
	}
}

func TestHostWorkspacePath(t *testing.T) {
	path := HostWorkspacePath("forgeai-workspace-abc123")
	if path == "" {
		t.Fatal("host workspace path should not be empty")
	}
	if !strings.Contains(path, "forgeai-workspace-abc123") {
		t.Fatalf("expected host workspace path to include volume name, got %q", path)
	}
	if !strings.Contains(path, "forgeai-workspaces") {
		t.Fatalf("expected host workspace path to live under forgeai-workspaces, got %q", path)
	}
}

func TestParseListFilesIgnoresSummaryHeader(t *testing.T) {
	emptyEntries := parseListFiles("total 0\n", "/workspace")
	if len(emptyEntries) != 0 {
		t.Fatalf("expected empty directory listing to produce no entries, got %#v", emptyEntries)
	}

	entries := parseListFiles("total 2\n-rw-r--r--    1 root     root             0 2026-09-27 09:32:08 +0000 .forgeai-repository-cloned\ndrwxr-xr-x    1 root     root          4096 2026-09-27 09:32:05 +0000 .git/\n-rw-r--r--    1 root     root            32 2026-09-27 09:32:05 +0000 README.md\n", "/workspace")
	if len(entries) != 3 {
		t.Fatalf("expected 3 real entries, got %d: %#v", len(entries), entries)
	}
	if entries[0].Name != ".forgeai-repository-cloned" || entries[1].Name != ".git" || entries[2].Name != "README.md" {
		t.Fatalf("unexpected parsed entries: %#v", entries)
	}
	if !entries[0].IsDirectory && entries[1].IsDirectory != true && entries[2].IsDirectory {
		t.Fatalf("unexpected directory flags: %#v", entries)
	}
	if entries[0].Path != "/workspace/.forgeai-repository-cloned" || entries[1].Path != "/workspace/.git" || entries[2].Path != "/workspace/README.md" {
		t.Fatalf("unexpected entry paths: %#v", entries)
	}
}
