package policy

import (
	"path/filepath"
	"testing"
)

func TestLoadSandboxPolicy(t *testing.T) {
	loaded, err := Load(filepath.Join("..", "..", "..", "..", "configs", "sandbox.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Image == "" || loaded.CommandTimeout <= 0 || loaded.Limits.MemoryBytes <= 0 {
		t.Fatalf("incomplete sandbox policy: %#v", loaded)
	}
}
