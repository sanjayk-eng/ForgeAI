package factory

import "testing"

func TestCreateSupportedShells(t *testing.T) {
	for _, selected := range []Shell{ShellPowerShell, ShellCMD, ShellBash, ShellZsh} {
		created, err := Create(selected)
		if err != nil || created == nil {
			t.Fatalf("Create(%q) returned executor=%v, err=%v", selected, created, err)
		}
	}
}

func TestCreateRejectsUnsupportedShell(t *testing.T) {
	if _, err := Create(Shell("unknown")); err == nil {
		t.Fatal("expected unsupported shell error")
	}
}

func TestResolveUsesExplicitShell(t *testing.T) {
	created, selected, err := Resolve(ShellCMD)
	if err != nil || created == nil {
		t.Fatalf("Resolve(%q) returned executor=%v, shell=%q, err=%v", ShellCMD, created, selected, err)
	}
	if selected != ShellCMD {
		t.Fatalf("Resolve(%q) selected shell=%q", ShellCMD, selected)
	}
}

func TestDetectShellReturnsSupportedShell(t *testing.T) {
	switch DetectShell() {
	case ShellPowerShell, ShellCMD, ShellBash, ShellZsh:
	default:
		t.Fatalf("DetectShell() returned unsupported shell %q", DetectShell())
	}
}
