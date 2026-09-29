//go:build windows

package docker

import (
	"context"
	"strings"
	"sync"

	"ai-agent/internal/modules/terminal/application"

	"github.com/UserExistsError/conpty"
)

type windowsInteractiveProcess struct {
	terminal  *conpty.ConPty
	closeOnce sync.Once
}

func startInteractiveProcess(binary string, args []string, cols, rows int) (application.InteractiveProcess, error) {
	commandLine := make([]string, 0, len(args)+1)
	commandLine = append(commandLine, quoteWindowsArgument(binary))
	for _, argument := range args {
		commandLine = append(commandLine, quoteWindowsArgument(argument))
	}
	terminal, err := conpty.Start(strings.Join(commandLine, " "), conpty.ConPtyDimensions(cols, rows))
	if err != nil {
		return nil, err
	}
	return &windowsInteractiveProcess{terminal: terminal}, nil
}

func quoteWindowsArgument(argument string) string {
	if argument != "" && !strings.ContainsAny(argument, " \t\n\v\"") {
		return argument
	}
	var quoted strings.Builder
	quoted.WriteByte('"')
	backslashes := 0
	for _, character := range argument {
		if character == '\\' {
			backslashes++
			continue
		}
		if character == '"' {
			quoted.WriteString(strings.Repeat("\\", backslashes*2+1))
			quoted.WriteRune(character)
			backslashes = 0
			continue
		}
		quoted.WriteString(strings.Repeat("\\", backslashes))
		quoted.WriteRune(character)
		backslashes = 0
	}
	quoted.WriteString(strings.Repeat("\\", backslashes*2))
	quoted.WriteByte('"')
	return quoted.String()
}

func (process *windowsInteractiveProcess) Read(data []byte) (int, error) {
	return process.terminal.Read(data)
}

func (process *windowsInteractiveProcess) Write(data []byte) (int, error) {
	return process.terminal.Write(data)
}

func (process *windowsInteractiveProcess) Resize(cols, rows int) error {
	return process.terminal.Resize(cols, rows)
}

func (process *windowsInteractiveProcess) Wait() (int, error) {
	exitCode, err := process.terminal.Wait(context.Background())
	return int(exitCode), err
}

func (process *windowsInteractiveProcess) Close() error {
	var closeErr error
	process.closeOnce.Do(func() {
		closeErr = process.terminal.Close()
	})
	return closeErr
}
