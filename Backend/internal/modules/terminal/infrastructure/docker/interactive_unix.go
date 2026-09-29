//go:build !windows

package docker

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"ai-agent/internal/modules/terminal/application"

	"github.com/creack/pty"
)

type unixInteractiveProcess struct {
	command   *exec.Cmd
	terminal  *os.File
	closeOnce sync.Once
}

func startInteractiveProcess(binary string, args []string, cols, rows int) (application.InteractiveProcess, error) {
	command := exec.Command(binary, args...)
	terminal, err := pty.StartWithSize(command, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
	if err != nil {
		return nil, fmt.Errorf("create Docker exec PTY: %w", err)
	}
	return &unixInteractiveProcess{command: command, terminal: terminal}, nil
}

func (process *unixInteractiveProcess) Read(data []byte) (int, error) {
	return process.terminal.Read(data)
}

func (process *unixInteractiveProcess) Write(data []byte) (int, error) {
	return process.terminal.Write(data)
}

func (process *unixInteractiveProcess) Resize(cols, rows int) error {
	return pty.Setsize(process.terminal, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
}

func (process *unixInteractiveProcess) Wait() (int, error) {
	err := process.command.Wait()
	if err == nil {
		return 0, nil
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode(), nil
	}
	return -1, err
}

func (process *unixInteractiveProcess) Close() error {
	var closeErr error
	process.closeOnce.Do(func() {
		_, _ = process.terminal.Write([]byte{3, 4})
		closeErr = process.terminal.Close()
		if process.command.Process != nil {
			_ = process.command.Process.Kill()
		}
	})
	return closeErr
}
