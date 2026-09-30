package docker

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

const previewTunnelScript = `
const net = require('net');
const socket = net.connect({ host: '127.0.0.1', port: Number(process.argv[1]) }, () => {
	process.stderr.write('READY\n');
	process.stdin.pipe(socket);
	socket.pipe(process.stdout);
});
socket.on('error', error => { process.stderr.write('ERROR ' + error.message + '\n'); process.exit(1); });
process.stdin.on('end', () => socket.end());
`

func (runtime *DockerRuntime) OpenPreviewTunnel(ctx context.Context, containerID string, containerPort int) (net.Conn, error) {
	if strings.TrimSpace(containerID) == "" || containerPort < 1 || containerPort > 65535 {
		return nil, fmt.Errorf("container and valid preview port are required")
	}
	tunnelContext, cancel := context.WithCancel(ctx)
	command := exec.CommandContext(tunnelContext, runtime.docker.binary, "exec", "-i", containerID, "node", "-e", previewTunnelScript, strconv.Itoa(containerPort))
	stdin, err := command.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	if err := command.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start Docker preview tunnel: %w", err)
	}
	ready := make(chan error, 1)
	go func() {
		line, readErr := bufio.NewReader(stderr).ReadString('\n')
		if readErr != nil {
			ready <- fmt.Errorf("Docker preview tunnel exited before connecting: %w", readErr)
			return
		}
		if strings.TrimSpace(line) != "READY" {
			ready <- fmt.Errorf("Docker preview tunnel failed: %s", strings.TrimSpace(line))
			return
		}
		ready <- nil
	}()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case err := <-ready:
		if err != nil {
			cancel()
			_ = command.Wait()
			return nil, err
		}
	case <-ctx.Done():
		cancel()
		_ = command.Wait()
		return nil, ctx.Err()
	case <-timer.C:
		cancel()
		_ = command.Wait()
		return nil, fmt.Errorf("timed out connecting to sandbox preview port")
	}
	proxyConn, tunnelConn := net.Pipe()
	go func() {
		inputDone := make(chan struct{})
		go func() {
			_, _ = io.Copy(stdin, tunnelConn)
			_ = stdin.Close()
			close(inputDone)
		}()
		_, _ = io.Copy(tunnelConn, stdout)
		_ = tunnelConn.Close()
		cancel()
		<-inputDone
		_ = command.Wait()
	}()
	return &previewTunnelConn{Conn: proxyConn, cancel: cancel}, nil
}

type previewTunnelConn struct {
	net.Conn
	cancel context.CancelFunc
	close  sync.Once
}

func (connection *previewTunnelConn) Close() error {
	var err error
	connection.close.Do(func() {
		err = connection.Conn.Close()
		connection.cancel()
	})
	return err
}
