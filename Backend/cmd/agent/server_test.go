package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSandboxPolicyPathFromAgentDirectory(t *testing.T) {
	path, err := sandboxPolicyPath()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "sandbox.yaml" {
		t.Fatalf("unexpected sandbox policy path: %s", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("sandbox policy path is not readable: %v", err)
	}
}
