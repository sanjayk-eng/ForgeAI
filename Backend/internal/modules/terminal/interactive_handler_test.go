package terminal

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"ai-agent/internal/middleware"
	"ai-agent/internal/modules/terminal/application"
	"ai-agent/internal/modules/terminal/domain"
	"ai-agent/internal/modules/terminal/policy"
	appjwt "ai-agent/pkg/jwt"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

func TestTerminalRoutesEnforceProjectAndSessionOwnership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtManager, err := appjwt.NewManager(strings.Repeat("t", 32), time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	ownerTokens, err := jwtManager.GeneratePair("member-1")
	if err != nil {
		t.Fatal(err)
	}
	otherTokens, err := jwtManager.GeneratePair("member-2")
	if err != nil {
		t.Fatal(err)
	}
	runtime := &interactiveTestRuntime{}
	service := application.NewService(interactiveTestStore{}, runtime, nil, nil, policy.Sandbox{})
	module := &Module{Service: service, Sessions: NewSessionManager(service)}
	router := gin.New()
	RegisterRoutes(middleware.ProtectedGroup(router, jwtManager, nil), NewHandler(module))
	server := httptest.NewServer(router)
	defer server.Close()

	request, err := http.NewRequest(http.MethodGet, server.URL+"/projects/project-1/terminals/shells", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+ownerTokens.AccessToken)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	var shellPayload struct {
		Data struct {
			Shells []application.TerminalShell `json:"shells"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&shellPayload); err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK || len(shellPayload.Data.Shells) != 1 || shellPayload.Data.Shells[0].ID != "sh" {
		t.Fatalf("shell list status/payload = %d/%+v", response.StatusCode, shellPayload)
	}

	request, err = http.NewRequest(http.MethodPost, server.URL+"/projects/project-1/terminals", strings.NewReader(`{"shell":"/bin/bash -c evil"}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+ownerTokens.AccessToken)
	request.Header.Set("Content-Type", "application/json")
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("arbitrary shell status = %d, want %d", response.StatusCode, http.StatusBadRequest)
	}
	if runtime.startedShell != "" {
		t.Fatalf("unavailable shell reached runtime: %q", runtime.startedShell)
	}

	request, err = http.NewRequest(http.MethodPost, server.URL+"/projects/project-1/terminals", strings.NewReader(`{"shell":"sh"}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+ownerTokens.AccessToken)
	request.Header.Set("Content-Type", "application/json")
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	var sessionPayload struct {
		Data TerminalSessionInfo `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&sessionPayload); err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusCreated || sessionPayload.Data.SessionID == "" || sessionPayload.Data.CWD != "/workspace" {
		t.Fatalf("terminal create status/payload = %d/%+v", response.StatusCode, sessionPayload)
	}

	closeURL := server.URL + "/projects/project-1/terminals/" + sessionPayload.Data.SessionID
	webSocketURL := "ws" + strings.TrimPrefix(closeURL, "http") + "/ws"
	connection, _, err := websocket.Dial(context.Background(), webSocketURL, &websocket.DialOptions{
		Subprotocols: []string{"forgeai", "forgeai-auth." + ownerTokens.AccessToken},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close(websocket.StatusNormalClosure, "test complete")
	readTerminalFrame(t, connection)
	if err := connection.Write(context.Background(), websocket.MessageText, []byte(`{"type":"input","data":"cd src\n"}`)); err != nil {
		t.Fatal(err)
	}
	select {
	case input := <-runtime.process.inputs:
		if string(input) != "cd src\n" {
			t.Fatalf("terminal input = %q", input)
		}
	case <-time.After(time.Second):
		t.Fatal("terminal input was not forwarded to the PTY")
	}
	if err := connection.Write(context.Background(), websocket.MessageText, []byte(`{"type":"resize","cols":120,"rows":30}`)); err != nil {
		t.Fatal(err)
	}
	select {
	case size := <-runtime.process.resizes:
		if size != [2]int{120, 30} {
			t.Fatalf("terminal resize = %v", size)
		}
	case <-time.After(time.Second):
		t.Fatal("terminal resize was not forwarded to the PTY")
	}
	if _, err := runtime.process.writer.Write([]byte("streamed output")); err != nil {
		t.Fatal(err)
	}
	output := readTerminalFrame(t, connection)
	if output.Type != "output" || output.Data != "streamed output" {
		t.Fatalf("terminal output frame = %+v", output)
	}

	request, err = http.NewRequest(http.MethodDelete, closeURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+otherTokens.AccessToken)
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("unauthorized project close status = %d, want %d", response.StatusCode, http.StatusForbidden)
	}
	if runtime.process.wasClosed() {
		t.Fatal("unauthorized request closed the terminal process")
	}

	request, err = http.NewRequest(http.MethodDelete, closeURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+ownerTokens.AccessToken)
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK || !runtime.process.wasClosed() {
		t.Fatalf("authorized close status/process closed = %d/%t", response.StatusCode, runtime.process.wasClosed())
	}
	exit := readTerminalFrame(t, connection)
	if exit.Type != "exit" || exit.Code != 0 {
		t.Fatalf("terminal exit frame = %+v", exit)
	}
}

func readTerminalFrame(t *testing.T, connection *websocket.Conn) terminalServerMessage {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, payload, err := connection.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var message terminalServerMessage
	if err := json.Unmarshal(payload, &message); err != nil {
		t.Fatal(err)
	}
	return message
}

type interactiveTestStore struct{ application.Store }

func (interactiveTestStore) FindProjectWorkspace(context.Context, string) (string, error) {
	return "workspace-1", nil
}

func (interactiveTestStore) CanAccessWorkspace(_ context.Context, _ string, userID string) (bool, error) {
	return userID == "member-1", nil
}

func (interactiveTestStore) FindActiveByProject(context.Context, string) (domain.Sandbox, error) {
	return domain.Sandbox{
		ID: "sandbox-1", ProjectID: "project-1", organizationID: "workspace-1",
		ContainerID: "container-1", WorkspacePath: "/workspace", Status: domain.StatusRunning,
	}, nil
}

type interactiveTestRuntime struct {
	application.Runtime
	startedShell string
	process      *interactiveTestProcess
}

func (runtime *interactiveTestRuntime) ListTerminalShells(context.Context, string) ([]application.TerminalShell, error) {
	return []application.TerminalShell{{ID: "sh", Name: "Shell", Available: true, Default: true}}, nil
}

func (runtime *interactiveTestRuntime) StartTerminal(_ context.Context, _, _, shell string, _, _ int) (application.InteractiveProcess, error) {
	runtime.startedShell = shell
	runtime.process = newInteractiveTestProcess()
	return runtime.process, nil
}

type interactiveTestProcess struct {
	reader    *io.PipeReader
	writer    *io.PipeWriter
	closed    chan struct{}
	inputs    chan []byte
	resizes   chan [2]int
	closeOnce sync.Once
}

func newInteractiveTestProcess() *interactiveTestProcess {
	reader, writer := io.Pipe()
	return &interactiveTestProcess{
		reader: reader, writer: writer, closed: make(chan struct{}),
		inputs: make(chan []byte, 4), resizes: make(chan [2]int, 4),
	}
}

func (process *interactiveTestProcess) Read(data []byte) (int, error) {
	return process.reader.Read(data)
}
func (process *interactiveTestProcess) Write(data []byte) (int, error) {
	process.inputs <- append([]byte(nil), data...)
	return len(data), nil
}
func (process *interactiveTestProcess) Resize(cols, rows int) error {
	process.resizes <- [2]int{cols, rows}
	return nil
}
func (process *interactiveTestProcess) Wait() (int, error) {
	<-process.closed
	return 0, nil
}
func (process *interactiveTestProcess) Close() error {
	process.closeOnce.Do(func() {
		close(process.closed)
		_ = process.reader.Close()
		_ = process.writer.Close()
	})
	return nil
}
func (process *interactiveTestProcess) wasClosed() bool {
	select {
	case <-process.closed:
		return true
	default:
		return false
	}
}
