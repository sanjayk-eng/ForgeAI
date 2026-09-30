package policy

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLoadSandboxPolicy(t *testing.T) {
	loaded, err := Load(filepath.Join("..", "..", "..", "..", "configs", "sandbox.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Image != "forgeai-sandbox:latest" || loaded.CommandTimeout <= 0 || loaded.Limits.MemoryBytes <= 0 ||
		loaded.PreviewPort < MinPreviewPort || loaded.PreviewPort > MaxPreviewPort {
		t.Fatalf("incomplete sandbox policy: %#v", loaded)
	}
}

func TestValidatePreviewPortDefaultsAndBounds(t *testing.T) {
	validPolicy := func(previewPort int) Sandbox {
		return Sandbox{
			Image: "sandbox", GitImage: "git", WorkspacePath: "/workspace", NetworkMode: "bridge",
			PreviewPort: previewPort, CommandTimeoutRaw: time.Minute.String(),
			Limits: Limits{CPUShares: 1, MemoryBytes: 1, PidsLimit: 1},
		}
	}

	defaultPort := validPolicy(0)
	if err := defaultPort.validate(); err != nil {
		t.Fatal(err)
	}
	if defaultPort.PreviewPort != DefaultPreviewPort {
		t.Fatalf("default preview port = %d, want %d", defaultPort.PreviewPort, DefaultPreviewPort)
	}

	for _, previewPort := range []int{MinPreviewPort - 1, MaxPreviewPort + 1} {
		candidate := validPolicy(previewPort)
		if err := candidate.validate(); err == nil {
			t.Errorf("preview port %d unexpectedly validated", previewPort)
		}
	}

	customPort := validPolicy(4000)
	if err := customPort.validate(); err != nil {
		t.Fatalf("custom preview port failed validation: %v", err)
	}
}
