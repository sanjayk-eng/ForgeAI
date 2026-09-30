package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"ai-agent/internal/modules/terminal/application"
)

func TestParseGoDefinitionLocation(t *testing.T) {
	payload := json.RawMessage(`{"uri":"file:///workspace/Backend/cmd/agent/server.go","range":{"start":{"line":8,"character":3}}}`)
	definition, err := parseGoDefinition(payload, "/workspace")
	if err != nil {
		t.Fatalf("parseGoDefinition() error = %v", err)
	}
	if definition.Path != "/workspace/Backend/cmd/agent/server.go" || definition.Line != 9 || definition.Column != 4 {
		t.Fatalf("unexpected definition: %+v", definition)
	}
}

func TestParseGoDefinitionLocationLink(t *testing.T) {
	payload := json.RawMessage(`{"targetUri":"file:///workspace/pkg/run.go","targetSelectionRange":{"start":{"line":2,"character":0}}}`)
	definition, err := parseGoDefinition(payload, "/workspace")
	if err != nil {
		t.Fatalf("parseGoDefinition() error = %v", err)
	}
	if definition.Path != "/workspace/pkg/run.go" || definition.Line != 3 || definition.Column != 1 {
		t.Fatalf("unexpected definition link: %+v", definition)
	}
}

func TestParseGoDefinitionRejectsOutsideWorkspace(t *testing.T) {
	payload := json.RawMessage(`{"uri":"file:///usr/local/go/src/fmt/print.go","range":{"start":{"line":1,"character":0}}}`)
	if _, err := parseGoDefinition(payload, "/workspace"); !errors.Is(err, application.ErrDefinitionNotFound) {
		t.Fatalf("parseGoDefinition() error = %v, want ErrDefinitionNotFound", err)
	}
}

func TestParseGoWorkspaceRoot(t *testing.T) {
	root, err := parseGoWorkspaceRoot(`{"GOWORK":"","GOMOD":"/workspace/Backend/go.mod"}`, "/workspace")
	if err != nil || root != "/workspace/Backend" {
		t.Fatalf("parseGoWorkspaceRoot() = %q, %v; want module directory", root, err)
	}
	if _, err := parseGoWorkspaceRoot(`{"GOWORK":"/outside/go.work","GOMOD":"/outside/go.mod"}`, "/workspace"); err == nil {
		t.Fatal("workspace root outside /workspace unexpectedly accepted")
	}
}

func TestMissingGoplsMessage(t *testing.T) {
	err := fmt.Errorf("docker exec: /bin/sh: gopls: not found")
	if got := strings.Contains(err.Error(), "gopls"); !got {
		t.Fatal("expected gopls-alignment in error message")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatal("expected missing-tool wording in error message")
	}
}

func TestDefinitionLockIsSharedPerContainer(t *testing.T) {
	runtime := NewDockerRuntime("docker")
	first := runtime.definitionLock("container-a")
	second := runtime.definitionLock("container-a")
	third := runtime.definitionLock("container-b")
	if first != second {
		t.Fatal("expected same container to reuse the same lock")
	}
	if first == third {
		t.Fatal("expected different containers to use different locks")
	}
}
