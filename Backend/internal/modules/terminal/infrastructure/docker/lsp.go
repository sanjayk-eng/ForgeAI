package docker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"ai-agent/internal/modules/terminal/application"
)

const maxLSPMessageBytes = 16 * 1024 * 1024

type lspClient struct {
	input  io.Writer
	output *bufio.Reader
	nextID int
}

type lspEnvelope struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type lspPosition struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type lspRange struct {
	Start lspPosition `json:"start"`
}

type lspLocation struct {
	URI                  string   `json:"uri"`
	Range                lspRange `json:"range"`
	TargetURI            string   `json:"targetUri"`
	TargetSelectionRange lspRange `json:"targetSelectionRange"`
}

func (runtime *DockerRuntime) ResolveGoDefinition(ctx context.Context, containerID, workspacePath, filePath string, line, column int) (application.GoDefinition, error) {
	if strings.TrimSpace(containerID) == "" || line < 1 || column < 1 || !isWithinWorkspace(workspacePath, filePath) {
		return application.GoDefinition{}, fmt.Errorf("container, Go file and valid position are required")
	}
	lock := runtime.definitionLock(containerID)
	lock.Lock()
	defer lock.Unlock()
	source, err := runtime.docker.Run(ctx, "exec", containerID, "cat", "--", filePath)
	if err != nil {
		return application.GoDefinition{}, fmt.Errorf("read Go source for definition: %w", err)
	}
	workspaceRootOutput, err := runtime.docker.Run(ctx, "exec", "-w", path.Dir(filePath), containerID, "go", "env", "-json", "GOWORK", "GOMOD")
	if err != nil {
		return application.GoDefinition{}, fmt.Errorf("resolve Go module for definition: %w", err)
	}
	workspaceRoot, err := parseGoWorkspaceRoot(workspaceRootOutput, workspacePath)
	if err != nil {
		return application.GoDefinition{}, err
	}
	if _, err := runtime.docker.Run(ctx, "exec", containerID, "gopls", "version"); err != nil {
		return application.GoDefinition{}, fmt.Errorf("gopls is not installed in this sandbox; recreate the project sandbox from forgeai-sandbox:latest: %w", err)
	}
	requestContext, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	command := exec.CommandContext(requestContext, runtime.docker.binary, "exec", "-i", "-w", workspaceRoot, containerID, "gopls", "serve")
	stdin, err := command.StdinPipe()
	if err != nil {
		return application.GoDefinition{}, err
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return application.GoDefinition{}, err
	}
	command.Stderr = io.Discard
	if err := command.Start(); err != nil {
		return application.GoDefinition{}, fmt.Errorf("start gopls in sandbox: %w", err)
	}
	waited := false
	defer func() {
		cancel()
		_ = stdin.Close()
		if !waited {
			_ = command.Wait()
		}
	}()
	client := &lspClient{input: stdin, output: bufio.NewReader(stdout)}
	rootURI := fileURI(workspaceRoot)
	var initialized json.RawMessage
	if err := client.request(requestContext, "initialize", map[string]any{
		"processId": nil,
		"rootUri":   rootURI,
		"capabilities": map[string]any{
			"workspace":    map[string]any{"workspaceFolders": true},
			"textDocument": map[string]any{"definition": map[string]any{"linkSupport": true}},
		},
		"workspaceFolders": []map[string]string{{"uri": rootURI, "name": path.Base(workspaceRoot)}},
	}, &initialized); err != nil {
		return application.GoDefinition{}, fmt.Errorf("initialize gopls: %w", err)
	}
	if err := client.notify("initialized", map[string]any{}); err != nil {
		return application.GoDefinition{}, err
	}
	if err := client.notify("textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{
			"uri": fileURI(filePath), "languageId": "go", "version": 1, "text": source,
		},
	}); err != nil {
		return application.GoDefinition{}, err
	}
	var result json.RawMessage
	if err := client.request(requestContext, "textDocument/definition", map[string]any{
		"textDocument": map[string]string{"uri": fileURI(filePath)},
		"position":     map[string]int{"line": line - 1, "character": column - 1},
	}, &result); err != nil {
		return application.GoDefinition{}, fmt.Errorf("request Go definition: %w", err)
	}
	definition, err := parseGoDefinition(result, workspacePath)
	if err != nil {
		return application.GoDefinition{}, err
	}
	_ = client.request(requestContext, "shutdown", nil, nil)
	_ = client.notify("exit", nil)
	_ = stdin.Close()
	if err := command.Wait(); err != nil {
		return application.GoDefinition{}, fmt.Errorf("stop gopls: %w", err)
	}
	waited = true
	return definition, nil
}

func (client *lspClient) notify(method string, params any) error {
	return client.write(map[string]any{"jsonrpc": "2.0", "method": method, "params": params})
}

func (client *lspClient) request(ctx context.Context, method string, params any, result any) error {
	client.nextID++
	id := client.nextID
	if err := client.write(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params}); err != nil {
		return err
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		payload, err := readLSPMessage(client.output)
		if err != nil {
			return err
		}
		var message lspEnvelope
		if err := json.Unmarshal(payload, &message); err != nil {
			return err
		}
		if message.Method != "" {
			if len(message.ID) > 0 && !bytes.Equal(message.ID, []byte("null")) {
				if err := client.write(map[string]any{"jsonrpc": "2.0", "id": message.ID, "result": nil}); err != nil {
					return err
				}
			}
			continue
		}
		if string(message.ID) != strconv.Itoa(id) {
			continue
		}
		if message.Error != nil {
			return fmt.Errorf("LSP %s failed (%d): %s", method, message.Error.Code, message.Error.Message)
		}
		if result == nil || len(message.Result) == 0 || bytes.Equal(message.Result, []byte("null")) {
			return nil
		}
		return json.Unmarshal(message.Result, result)
	}
}

func (client *lspClient) write(message any) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	if len(payload) > maxLSPMessageBytes {
		return fmt.Errorf("LSP message exceeds the size limit")
	}
	if _, err := fmt.Fprintf(client.input, "Content-Length: %d\r\n\r\n", len(payload)); err != nil {
		return err
	}
	_, err = client.input.Write(payload)
	return err
}

func readLSPMessage(reader *bufio.Reader) ([]byte, error) {
	contentLength := -1
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		name, value, found := strings.Cut(line, ":")
		if found && strings.EqualFold(name, "Content-Length") {
			contentLength, err = strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return nil, fmt.Errorf("invalid LSP content length")
			}
		}
	}
	if contentLength < 0 || contentLength > maxLSPMessageBytes {
		return nil, fmt.Errorf("invalid LSP message size")
	}
	payload := make([]byte, contentLength)
	_, err := io.ReadFull(reader, payload)
	return payload, err
}

func parseGoDefinition(payload json.RawMessage, workspacePath string) (application.GoDefinition, error) {
	if len(payload) == 0 || bytes.Equal(payload, []byte("null")) {
		return application.GoDefinition{}, application.ErrDefinitionNotFound
	}
	var locations []lspLocation
	if payload[0] == '[' {
		if err := json.Unmarshal(payload, &locations); err != nil {
			return application.GoDefinition{}, err
		}
	} else {
		var location lspLocation
		if err := json.Unmarshal(payload, &location); err != nil {
			return application.GoDefinition{}, err
		}
		locations = []lspLocation{location}
	}
	for _, location := range locations {
		uri := location.URI
		rangeValue := location.Range
		if location.TargetURI != "" {
			uri = location.TargetURI
			rangeValue = location.TargetSelectionRange
		}
		parsed, err := url.Parse(uri)
		if err != nil || parsed.Scheme != "file" {
			continue
		}
		filePath := path.Clean(parsed.Path)
		root := path.Clean(workspacePath)
		if filePath == root || !strings.HasPrefix(filePath, root+"/") {
			continue
		}
		return application.GoDefinition{Path: filePath, Line: rangeValue.Start.Line + 1, Column: rangeValue.Start.Character + 1}, nil
	}
	return application.GoDefinition{}, application.ErrDefinitionNotFound
}

func fileURI(filePath string) string {
	return (&url.URL{Scheme: "file", Path: filePath}).String()
}

func parseGoWorkspaceRoot(output, workspacePath string) (string, error) {
	var environment struct {
		Work string `json:"GOWORK"`
		Mod  string `json:"GOMOD"`
	}
	if err := json.Unmarshal([]byte(output), &environment); err != nil {
		return "", fmt.Errorf("decode Go workspace settings: %w", err)
	}
	root := path.Clean(workspacePath)
	if environment.Mod != "" && environment.Mod != "/dev/null" {
		root = path.Dir(environment.Mod)
	}
	if environment.Work != "" && environment.Work != "off" {
		root = path.Dir(environment.Work)
	}
	workspaceRoot := path.Clean(workspacePath)
	if root != workspaceRoot && !strings.HasPrefix(root, workspaceRoot+"/") {
		return "", fmt.Errorf("Go workspace root is outside the sandbox workspace")
	}
	return root, nil
}

func isWithinWorkspace(root, filePath string) bool {
	root = path.Clean(root)
	filePath = path.Clean(filePath)
	return filePath != root && strings.HasPrefix(filePath, root+"/")
}

var _ application.GoDefinitionResolver = (*DockerRuntime)(nil)
