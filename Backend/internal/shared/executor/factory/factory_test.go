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
