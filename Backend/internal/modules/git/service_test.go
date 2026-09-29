package git

import "testing"

func TestParseStatusOutput(t *testing.T) {
	output := "## main\n M README.md\n?? notes.txt\nA  app.go\n"

	status, err := parseStatusOutput(output)
	if err != nil {
		t.Fatalf("parseStatusOutput() returned error: %v", err)
	}
	if status.Branch != "main" {
		t.Fatalf("parseStatusOutput() branch = %q, want %q", status.Branch, "main")
	}
	if !status.IsDirty {
		t.Fatal("parseStatusOutput() IsDirty = false, want true")
	}
	if len(status.Modified) != 1 || status.Modified[0] != "README.md" {
		t.Fatalf("parseStatusOutput() modified = %#v, want [README.md]", status.Modified)
	}
	if len(status.Untracked) != 1 || status.Untracked[0] != "notes.txt" {
		t.Fatalf("parseStatusOutput() untracked = %#v, want [notes.txt]", status.Untracked)
	}
	if len(status.Staged) != 1 || status.Staged[0] != "app.go" {
		t.Fatalf("parseStatusOutput() staged = %#v, want [app.go]", status.Staged)
	}
}

func TestParseStatusOutputRequiresRepoState(t *testing.T) {
	if _, err := parseStatusOutput("fatal: not a git repository"); err == nil {
		t.Fatal("parseStatusOutput() should reject non-repository output")
	}
}

func TestParseStatusOutputTracksDivergenceAndRenames(t *testing.T) {
	status, err := parseStatusOutput("## feature...origin/feature [ahead 2, behind 1]\nR  old.txt -> new.txt\n")
	if err != nil {
		t.Fatalf("parseStatusOutput() returned error: %v", err)
	}
	if status.Branch != "feature" || status.Ahead != 2 || status.Behind != 1 {
		t.Fatalf("parseStatusOutput() tracking = %#v, want feature/ahead=2/behind=1", status)
	}
	if len(status.Staged) != 1 || status.Staged[0] != "new.txt" {
		t.Fatalf("parseStatusOutput() staged = %#v, want [new.txt]", status.Staged)
	}
}
