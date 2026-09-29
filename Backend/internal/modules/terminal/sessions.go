package terminal

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"ai-agent/internal/modules/terminal/application"
)

var (
	errTerminalSessionNotFound = errors.New("terminal session not found")
	errTerminalSessionAttached = errors.New("terminal session is already connected")
)

type TerminalSessionInfo struct {
	SessionID string `json:"session_id"`
	ProjectID string `json:"project_id"`
	SandboxID string `json:"sandbox_id"`
	Shell     string `json:"shell"`
	CWD       string `json:"cwd"`
	Status    string `json:"status"`
}

type terminalSession struct {
	TerminalSessionInfo
	userID       string
	process      application.InteractiveProcess
	output       chan []byte
	detachNotify chan struct{}
	readerDone   chan struct{}
	done         chan struct{}
	closed       chan struct{}
	closeOnce    sync.Once
	mu           sync.Mutex
	attached     bool
	exitCode     int
	exitErr      error
	expires      *time.Timer
}

type SessionManager struct {
	service  *application.Service
	mu       sync.RWMutex
	sessions map[string]*terminalSession
}

func NewSessionManager(service *application.Service) *SessionManager {
	return &SessionManager{service: service, sessions: make(map[string]*terminalSession)}
}

func (manager *SessionManager) Create(ctx context.Context, userID, projectID, shell string, cols, rows int) (TerminalSessionInfo, error) {
	sandbox, process, err := manager.service.StartTerminal(ctx, userID, projectID, shell, cols, rows)
	if err != nil {
		return TerminalSessionInfo{}, err
	}
	sessionID, err := newTerminalSessionID()
	if err != nil {
		_ = process.Close()
		return TerminalSessionInfo{}, err
	}
	session := &terminalSession{
		TerminalSessionInfo: TerminalSessionInfo{
			SessionID: sessionID, ProjectID: projectID, SandboxID: sandbox.ID,
			Shell: shell, CWD: sandbox.WorkspacePath, Status: "running",
		},
		userID: userID, process: process, output: make(chan []byte, 64),
		detachNotify: make(chan struct{}, 1),
		readerDone:   make(chan struct{}), done: make(chan struct{}), closed: make(chan struct{}),
	}
	manager.mu.Lock()
	manager.sessions[sessionID] = session
	manager.mu.Unlock()
	go session.readOutput()
	go manager.waitForExit(session)
	return session.snapshot(), nil
}

func (manager *SessionManager) Attach(sessionID, userID, projectID string) (*terminalSession, error) {
	manager.mu.RLock()
	session := manager.sessions[sessionID]
	manager.mu.RUnlock()
	if session == nil {
		return nil, errTerminalSessionNotFound
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.userID != userID || session.ProjectID != projectID {
		return nil, errTerminalSessionNotFound
	}
	if session.attached {
		return nil, errTerminalSessionAttached
	}
	if session.expires != nil {
		session.expires.Stop()
		session.expires = nil
	}
	session.attached = true
	return session, nil
}

func (manager *SessionManager) Detach(session *terminalSession) {
	session.mu.Lock()
	if !session.attached {
		session.mu.Unlock()
		return
	}
	session.attached = false
	select {
	case session.detachNotify <- struct{}{}:
	default:
	}
	manager.scheduleExpirationLocked(session)
	session.mu.Unlock()
}

func (manager *SessionManager) Close(sessionID string) {
	manager.mu.Lock()
	session := manager.sessions[sessionID]
	delete(manager.sessions, sessionID)
	manager.mu.Unlock()
	if session != nil {
		session.close()
	}
}

func (manager *SessionManager) CloseOwned(sessionID, userID, projectID string) bool {
	manager.mu.Lock()
	session := manager.sessions[sessionID]
	if session == nil || session.userID != userID || session.ProjectID != projectID {
		manager.mu.Unlock()
		return false
	}
	delete(manager.sessions, sessionID)
	manager.mu.Unlock()
	session.close()
	return true
}

func (manager *SessionManager) CloseAll() {
	manager.mu.Lock()
	sessions := make([]*terminalSession, 0, len(manager.sessions))
	for sessionID, session := range manager.sessions {
		sessions = append(sessions, session)
		delete(manager.sessions, sessionID)
	}
	manager.mu.Unlock()
	for _, session := range sessions {
		session.close()
	}
}

func (manager *SessionManager) waitForExit(session *terminalSession) {
	exitCode, err := session.process.Wait()
	<-session.readerDone
	session.mu.Lock()
	session.exitCode = exitCode
	session.exitErr = err
	session.Status = "terminated"
	close(session.done)
	if !session.attached {
		manager.scheduleExpirationLocked(session)
	}
	session.mu.Unlock()
}

func (manager *SessionManager) scheduleExpirationLocked(session *terminalSession) {
	if session.expires != nil {
		session.expires.Stop()
	}
	session.expires = time.AfterFunc(2*time.Minute, func() {
		manager.expire(session)
	})
}

func (manager *SessionManager) expire(session *terminalSession) {
	manager.mu.Lock()
	if manager.sessions[session.SessionID] != session {
		manager.mu.Unlock()
		return
	}
	session.mu.Lock()
	if session.attached {
		session.expires = nil
		session.mu.Unlock()
		manager.mu.Unlock()
		return
	}
	delete(manager.sessions, session.SessionID)
	session.expires = nil
	session.mu.Unlock()
	manager.mu.Unlock()
	session.close()
}

func (session *terminalSession) readOutput() {
	defer close(session.readerDone)
	buffer := make([]byte, 4096)
	for {
		read, err := session.process.Read(buffer)
		if read > 0 {
			chunk := append([]byte(nil), buffer[:read]...)
			if !session.publishOutput(chunk) {
				return
			}
		}
		if err != nil {
			return
		}
	}
}

func (session *terminalSession) publishOutput(chunk []byte) bool {
	for {
		session.mu.Lock()
		attached := session.attached
		session.mu.Unlock()
		if !attached {
			select {
			case session.output <- chunk:
			default:
			}
			return true
		}
		select {
		case session.output <- chunk:
			return true
		case <-session.closed:
			return false
		case <-session.detachNotify:
		}
	}
}

func (session *terminalSession) write(data []byte) error {
	select {
	case <-session.done:
		return fmt.Errorf("terminal session has terminated")
	default:
	}
	_, err := session.process.Write(data)
	return err
}

func (session *terminalSession) resize(cols, rows int) error {
	select {
	case <-session.done:
		return fmt.Errorf("terminal session has terminated")
	default:
	}
	if cols < 2 || rows < 2 || cols > 500 || rows > 200 {
		return fmt.Errorf("terminal dimensions are out of range")
	}
	return session.process.Resize(cols, rows)
}

func (session *terminalSession) close() {
	session.closeOnce.Do(func() {
		close(session.closed)
		session.mu.Lock()
		if session.expires != nil {
			session.expires.Stop()
		}
		session.mu.Unlock()
		_ = session.process.Close()
	})
}

func (session *terminalSession) snapshot() TerminalSessionInfo {
	session.mu.Lock()
	defer session.mu.Unlock()
	return session.TerminalSessionInfo
}

func newTerminalSessionID() (string, error) {
	bytes := make([]byte, 18)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate terminal session ID: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
