import { Terminal as Xterm } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { useEffect, useRef } from "react";
import { apiUrl } from "../../../../../../shared/api/client";
import type { TerminalSession } from "../../../../api/terminal.api";

export type ConnectionState = "Connected" | "Connecting" | "Disconnected";
export type SessionActions = { clear: () => void; reconnect: () => void };

export function TerminalSessionView({
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
      retryCount.current += 1;
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

export function xtermTheme(theme: "dark" | "light") {
  return theme === "light"
    ? { background: "#EEF3FA", foreground: "#172234", cursor: "#2563EB", cursorAccent: "#EEF3FA", selectionBackground: "#2563EB33" }
    : { background: "#0C111A", foreground: "#EAF0FA", cursor: "#68A8FF", cursorAccent: "#0C111A", selectionBackground: "#418BE855" };
}
