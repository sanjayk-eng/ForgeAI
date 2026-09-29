import { useEffect, useRef, useState } from "react";
import { apiUrl } from "../../../../../shared/api/client";
import type { ProjectRealtimeEvent } from "../../../api/project-realtime.types";

export function useProjectRealtime(
  accessToken: string | null,
  projectId: string | null,
  onEvent: (event: ProjectRealtimeEvent) => void,
  onReconnect: () => void,
) {
  const [isConnected, setIsConnected] = useState(false);
  const callbacks = useRef({ onEvent, onReconnect });
  callbacks.current = { onEvent, onReconnect };

  useEffect(() => {
    if (!accessToken || !projectId) {
      setIsConnected(false);
      return;
    }
    const currentAccessToken = accessToken;
    const currentProjectId = projectId;

    let disposed = false;
    let retryTimer: ReturnType<typeof setTimeout> | undefined;
    let attempt = 0;
    let socket: WebSocket | undefined;

    function connect() {
      if (disposed) return;

      const endpoint = new URL(apiUrl(`/projects/${encodeURIComponent(currentProjectId)}/ws`));
      endpoint.protocol = endpoint.protocol === "https:" ? "wss:" : "ws:";

      try {
        socket = new WebSocket(endpoint, ["forgeai", `forgeai-auth.${currentAccessToken}`]);
      } catch {
        scheduleReconnect();
        return;
      }

      socket.onopen = () => {
        attempt = 0;
        setIsConnected(true);
        callbacks.current.onReconnect();
      };

      socket.onmessage = (message) => {
        if (typeof message.data !== "string") return;
        try {
          const event = JSON.parse(message.data) as ProjectRealtimeEvent;
          if (event.version === 1 && event.project_id === currentProjectId) {
            callbacks.current.onEvent(event);
          }
        } catch {
          // Ignore malformed events; reconnect resync restores the API-backed state.
        }
      };

      socket.onclose = () => {
        setIsConnected(false);
        scheduleReconnect();
      };

      socket.onerror = () => socket?.close();
    }

    function scheduleReconnect() {
      if (disposed || retryTimer) return;
      const delay = Math.min(10_000, 500 * 2 ** Math.min(attempt, 5));
      attempt += 1;
      retryTimer = setTimeout(() => {
        retryTimer = undefined;
        connect();
      }, delay);
    }

    connect();
    return () => {
      disposed = true;
      if (retryTimer) clearTimeout(retryTimer);
      socket?.close();
      setIsConnected(false);
    };
  }, [accessToken, projectId]);

  return isConnected;
}