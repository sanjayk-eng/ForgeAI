import { Terminal as Xterm } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";
import { Eraser, Plus, RotateCcw, Terminal as TerminalIcon, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { apiUrl } from "../../../../../shared/api/client";
import { useTheme } from "../../../../../shared/ui/themeContextStore";
import {
  closeTerminalSession,
  createTerminalSession,
  listTerminalShells,
  type TerminalSession,
  type TerminalShell,
} from "../../../api/terminal.api";

type ConnectionState = "Connected" | "Connecting" | "Disconnected";
type SessionActions = { clear: () => void; reconnect: () => void };

export function ManualTerminal({ projectId, accessToken }: { projectId: string; accessToken: string | null }) {
  const { resolvedTheme } = useTheme();
  const [shellResult, setShellResult] = useState<TerminalShell[] | null>(null);
  const [selectedShell, setSelectedShell] = useState("");
  const [sessions, setSessions] = useState<TerminalSession[]>([]);
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);
  const [states, setStates] = useState<Record<string, ConnectionState>>({});
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState("");
  const actions = useRef(new Map<string, SessionActions>());
  const sessionsRef = useRef(sessions);
  const accessTokenRef = useRef(accessToken);
  useEffect(() => {
    sessionsRef.current = sessions;
  }, [sessions]);
  useEffect(() => {
    accessTokenRef.current = accessToken;
  }, [accessToken]);

  const shells = shellResult ?? [];
  const loading = Boolean(accessToken && shellResult === null);
  const terminalError = accessToken ? error : "Sign in to open a terminal.";

  useEffect(() => {
    if (!accessToken) return;
    let active = true;
    listTerminalShells(accessToken, projectId)
      .then((availableShells) => {
        if (!active) return;
        const available = availableShells.filter((shell) => shell.available);
        setShellResult(available);
        setSelectedShell(available.find((shell) => shell.default)?.id ?? available[0]?.id ?? "");
        setError(available.length ? "" : "No supported shells are installed in this sandbox.");
      })
      .catch((cause: unknown) => {
        if (!active) return;
        setShellResult([]);
        setError(cause instanceof Error ? cause.message : "Could not load available shells.");
      });
    return () => {
      active = false;
    };
  }, [accessToken, projectId]);

  useEffect(() => () => {
    const token = accessTokenRef.current;
    if (!token) return;
    for (const session of sessionsRef.current) {
      void closeTerminalSession(token, projectId, session.session_id).catch(() => undefined);
    }
  }, [projectId]);

  async function startSession() {
    if (!accessToken || !selectedShell || creating) return;
    setCreating(true);
    setError("");
    try {
      const session = await createTerminalSession(accessToken, projectId, selectedShell);
      setSessions((current) => [...current, session]);
      setActiveSessionId(session.session_id);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Could not start the terminal.");
    } finally {
      setCreating(false);
    }
  }

  async function closeSession(session: TerminalSession) {
    if (!accessToken) return;
    actions.current.delete(session.session_id);
    setSessions((current) => current.filter((item) => item.session_id !== session.session_id));
    setStates((current) => {
      const next = { ...current };
      delete next[session.session_id];
      return next;
    });
    setActiveSessionId((current) => current === session.session_id
      ? sessions.find((item) => item.session_id !== session.session_id)?.session_id ?? null
      : current);
    try {
      await closeTerminalSession(accessToken, projectId, session.session_id);
    } catch {
      setError("The terminal disconnected before it could be closed on the server.");
    }
  }

  const activeState = activeSessionId ? states[activeSessionId] ?? "Connecting" : "Disconnected";
  const activeSession = sessions.find((session) => session.session_id === activeSessionId);

  return (
    <section className="flex h-full min-h-0 min-w-0 flex-1 flex-col bg-forge-bg text-forge-text">
      <header className="flex min-h-12 shrink-0 flex-wrap items-center gap-2 border-b border-[var(--border)] bg-forge-panel px-3 py-2">
        <span className="mr-1 inline-flex items-center gap-2 text-xs font-semibold text-forge-text"><TerminalIcon size={15} /> Terminal</span>
        <div className="mr-1 flex items-center gap-2 text-xs font-semibold text-forge-muted">
          <span className={`size-2 rounded-full ${activeState === "Connected" ? "bg-emerald-500" : activeState === "Connecting" ? "bg-amber-500" : "bg-forge-muted"}`} />
          {activeState}
        </div>
        <select
          aria-label="Terminal shell"
          value={selectedShell}
          onChange={(event) => setSelectedShell(event.target.value)}
          disabled={loading || shells.length === 0}
          className="h-8 max-w-40 rounded border border-[var(--border)] bg-[var(--input)] px-2 text-xs text-forge-text outline-none focus:border-forge-accent"
        >
          {shells.map((shell) => <option key={shell.id} value={shell.id}>{shell.name}</option>)}
        </select>
        <button
          type="button"
          onClick={() => void startSession()}
          disabled={loading || creating || !selectedShell}
          className="inline-flex h-8 items-center gap-1.5 rounded border border-forge-accent/40 px-2.5 text-xs font-medium text-forge-accent hover:bg-forge-accent/10 disabled:opacity-50"
        >
          <Plus size={14} /> New Terminal
        </button>
        {activeSession && <span className="min-w-0 truncate text-[11px] text-forge-muted">{activeSession.cwd}</span>}
        <div className="ml-auto flex items-center gap-1">
          <TerminalToolButton label="Reconnect" disabled={!activeSession} onClick={() => activeSessionId && actions.current.get(activeSessionId)?.reconnect()}>
            <RotateCcw size={15} />
          </TerminalToolButton>
          <TerminalToolButton label="Clear" disabled={!activeSession} onClick={() => activeSessionId && actions.current.get(activeSessionId)?.clear()}>
            <Eraser size={15} />
          </TerminalToolButton>
          <TerminalToolButton label="Close terminal" disabled={!activeSession} onClick={() => activeSession && void closeSession(activeSession)}>
            <X size={16} />
          </TerminalToolButton>
        </div>
      </header>

      {sessions.length > 0 && (
        <nav aria-label="Terminal sessions" className="flex h-9 shrink-0 items-stretch overflow-x-auto border-b border-[var(--border)] bg-[var(--surface-subtle)]">
          {sessions.map((session, index) => (
            <button
              type="button"
              key={session.session_id}
              onClick={() => setActiveSessionId(session.session_id)}
              className={`min-w-24 border-r border-[var(--border)] px-3 text-left text-[11px] ${activeSessionId === session.session_id ? "border-b-2 border-b-forge-accent bg-forge-bg text-forge-text" : "text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text"}`}
            >
              {session.shell} {index + 1}
            </button>
          ))}
        </nav>
      )}

      <div className="relative min-h-0 flex-1 overflow-hidden px-3 py-2">
        {sessions.map((session) => (
          <TerminalSessionView
            key={session.session_id}
            session={session}
            projectId={projectId}
            accessToken={accessToken}
            active={session.session_id === activeSessionId}
            resolvedTheme={resolvedTheme}
            registerActions={(sessionId, value) => {
              if (value) actions.current.set(sessionId, value);
              else actions.current.delete(sessionId);
            }}
            onState={(sessionId, state) => setStates((current) => ({ ...current, [sessionId]: state }))}
          />
        ))}
        {sessions.length === 0 && !loading && !error && (
          <div className="flex h-full items-center justify-center text-xs text-forge-muted">Select a shell and start a terminal.</div>
        )}
        {loading && <div className="p-3 text-xs text-forge-muted">Checking sandbox shells...</div>}
        {terminalError && <div role="alert" className="absolute inset-x-3 bottom-3 rounded border border-[var(--border)] bg-forge-panel px-3 py-2 text-xs text-forge-signal">{terminalError}</div>}
      </div>
    </section>
  );
}

function TerminalToolButton({
  label,
  disabled,
  onClick,
  children,
}: {
  label: string;
  disabled: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      title={label}
      aria-label={label}
      disabled={disabled}
      onClick={onClick}
      className="grid size-8 place-items-center rounded text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text disabled:opacity-30"
    >
      {children}
    </button>
  );
}

function TerminalSessionView({
  session,
  projectId,
  accessToken,
  active,
  resolvedTheme,
  registerActions,
  onState,
}: {
  session: TerminalSession;
  projectId: string;
  accessToken: string | null;
  active: boolean;
  resolvedTheme: "dark" | "light";
  registerActions: (sessionId: string, actions: SessionActions | null) => void;
  onState: (sessionId: string, state: ConnectionState) => void;
}) {
  const host = useRef<HTMLDivElement>(null);
  const terminal = useRef<Xterm | null>(null);
  const socket = useRef<WebSocket | null>(null);
  const connectRef = useRef<() => void>(() => undefined);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const retryCount = useRef(0);
  const disposed = useRef(false);
  const reconnectRequested = useRef(false);
  const resolvedThemeRef = useRef(resolvedTheme);
  const registerActionsRef = useRef(registerActions);
  const onStateRef = useRef(onState);
  useEffect(() => {
    resolvedThemeRef.current = resolvedTheme;
    if (terminal.current) terminal.current.options.theme = xtermTheme(resolvedTheme);
    registerActionsRef.current = registerActions;
    onStateRef.current = onState;
  }, [resolvedTheme, registerActions, onState]);

  useEffect(() => {
    if (!host.current || !accessToken) return;
    disposed.current = false;
    const emulator = new Xterm({
      cursorBlink: true,
      cursorStyle: "bar",
      cursorWidth: 2,
      fontSize: 13,
      fontFamily: "'Cascadia Code', 'SFMono-Regular', Consolas, monospace",
      scrollback: 5000,
      theme: xtermTheme(resolvedThemeRef.current),
    });
    const fit = new FitAddon();
    emulator.loadAddon(fit);
    emulator.open(host.current);
    terminal.current = emulator;

    function send(message: object) {
      if (socket.current?.readyState === WebSocket.OPEN) socket.current.send(JSON.stringify(message));
    }

    function fitTerminal() {
      if (!host.current || !host.current.clientWidth || !host.current.clientHeight) return;
      try {
        fit.fit();
        send({ type: "resize", cols: emulator.cols, rows: emulator.rows });
      } catch {
        onStateRef.current(session.session_id, "Disconnected");
      }
    }

    function scheduleReconnect() {
      if (disposed.current || reconnectTimer.current) return;
      const delay = Math.min(10_000, 500 * 2 ** Math.min(retryCount.current, 5));
      retryCount.current++;
      reconnectTimer.current = setTimeout(() => {
        reconnectTimer.current = undefined;
        connect();
      }, delay);
    }

    function connect() {
      if (disposed.current || !accessToken) return;
      onStateRef.current(session.session_id, "Connecting");
      try {
        const endpoint = new URL(apiUrl(`/projects/${encodeURIComponent(projectId)}/terminals/${encodeURIComponent(session.session_id)}/ws`));
        endpoint.protocol = endpoint.protocol === "https:" ? "wss:" : "ws:";
        const connection = new WebSocket(endpoint, ["forgeai", `forgeai-auth.${accessToken}`]);
        socket.current = connection;
        connection.onopen = () => {
          retryCount.current = 0;
          onStateRef.current(session.session_id, "Connected");
          fitTerminal();
        };
        connection.onmessage = (event) => {
          if (typeof event.data !== "string") return;
          try {
            const message = JSON.parse(event.data) as { type: string; data?: string; message?: string; code?: number };
            if (message.type === "output" && message.data) emulator.write(message.data);
            if (message.type === "error") emulator.write(`\r\n\x1b[31m${message.message ?? "Terminal error"}\x1b[0m\r\n`);
            if (message.type === "exit") {
              emulator.write(`\r\n\x1b[90m[terminal exited: ${message.code ?? 0}]\x1b[0m\r\n`);
              onStateRef.current(session.session_id, "Disconnected");
            }
          } catch {
            emulator.write("\r\n\x1b[31mInvalid terminal response.\x1b[0m\r\n");
          }
        };
        connection.onclose = () => {
          if (socket.current === connection) socket.current = null;
          if (disposed.current) return;
          onStateRef.current(session.session_id, "Disconnected");
          if (reconnectRequested.current) {
            reconnectRequested.current = false;
            connect();
          } else {
            scheduleReconnect();
          }
        };
        connection.onerror = () => connection.close();
      } catch {
        onStateRef.current(session.session_id, "Disconnected");
        scheduleReconnect();
      }
    }

    connectRef.current = connect;
    const input = emulator.onData((data) => send({ type: "input", data }));
    const resize = emulator.onResize(({ cols, rows }) => send({ type: "resize", cols, rows }));
    const observer = new ResizeObserver(fitTerminal);
    observer.observe(host.current);
    const initialFit = requestAnimationFrame(fitTerminal);

    registerActionsRef.current(session.session_id, {
      clear: () => emulator.clear(),
      reconnect: () => {
        if (socket.current?.readyState === WebSocket.OPEN || socket.current?.readyState === WebSocket.CONNECTING) {
          reconnectRequested.current = true;
          socket.current.close();
        } else {
          if (reconnectTimer.current) clearTimeout(reconnectTimer.current);
          reconnectTimer.current = undefined;
          connect();
        }
      },
    });

    return () => {
      disposed.current = true;
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current);
      observer.disconnect();
      cancelAnimationFrame(initialFit);
      input.dispose();
      resize.dispose();
      socket.current?.close();
      socket.current = null;
      emulator.dispose();
      terminal.current = null;
      registerActionsRef.current(session.session_id, null);
    };
  }, [accessToken, projectId, session.session_id]);

  useEffect(() => {
    if (active) requestAnimationFrame(() => terminal.current?.focus());
  }, [active]);

  return <div ref={host} className={`h-full min-h-0 w-full ${active ? "" : "hidden"}`} />;
}

function xtermTheme(theme: "dark" | "light") {
  return theme === "light"
    ? { background: "#EEF3FA", foreground: "#172234", cursor: "#2563EB", cursorAccent: "#EEF3FA", selectionBackground: "#2563EB33" }
    : { background: "#0C111A", foreground: "#EAF0FA", cursor: "#68A8FF", cursorAccent: "#0C111A", selectionBackground: "#418BE855" };
}
