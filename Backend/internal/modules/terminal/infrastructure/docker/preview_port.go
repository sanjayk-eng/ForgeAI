package docker

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ai-agent/internal/modules/terminal/application"
)

type previewPortCache struct {
	port      int
	checkedAt time.Time
}

const previewPortCacheTTL = 2 * time.Second

const detectPreviewPortScript = `
const fs = require('fs');
const http = require('http');
const ports = new Set();
for (const file of ['/proc/net/tcp', '/proc/net/tcp6']) {
	try {
		for (const line of fs.readFileSync(file, 'utf8').trim().split('\n').slice(1)) {
			const fields = line.trim().split(/\s+/);
			if (fields[3] === '0A') {
				const port = Number.parseInt(fields[1].split(':').pop(), 16);
				if (Number.isInteger(port) && port > 0) ports.add(port);
			}
		}
	} catch {}
}
function responds(port) {
	return new Promise(resolve => {
		let settled = false;
		const finish = value => { if (!settled) { settled = true; resolve(value); } };
		const request = http.get({ host: '127.0.0.1', port, path: '/', timeout: 500 }, response => {
			response.resume();
			finish(true);
		});
		request.on('error', () => finish(false));
		request.on('timeout', () => { request.destroy(); finish(false); });
	});
}
(async () => {
	for (const port of [...ports].sort((left, right) => left - right)) {
		if (await responds(port)) { process.stdout.write(String(port)); return; }
	}
})();
`

func (runtime *DockerRuntime) forgetPreviewPort(containerID string) {
	runtime.previewMu.Lock()
	delete(runtime.previewPort, containerID)
	runtime.previewMu.Unlock()
}

func (runtime *DockerRuntime) DetectPreviewPort(ctx context.Context, containerID string) (int, error) {
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return 0, fmt.Errorf("container is required")
	}
	runtime.previewMu.Lock()
	defer runtime.previewMu.Unlock()
	if cached, ok := runtime.previewPort[containerID]; ok && time.Since(cached.checkedAt) < previewPortCacheTTL {
		return cached.port, nil
	}
	output, err := runtime.docker.Run(ctx, "exec", containerID, "node", "-e", detectPreviewPortScript)
	if err != nil {
		return 0, err
	}
	port, err := parsePreviewPort(output)
	if err != nil {
		return 0, err
	}
	if runtime.previewPort == nil {
		runtime.previewPort = make(map[string]previewPortCache)
	}
	runtime.previewPort[containerID] = previewPortCache{port: port, checkedAt: time.Now()}
	return port, nil
}

func parsePreviewPort(output string) (int, error) {
	value := strings.TrimSpace(output)
	if value == "" {
		return 0, application.ErrPreviewPortNotFound
	}
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("sandbox returned an invalid preview port")
	}
	return port, nil
}
