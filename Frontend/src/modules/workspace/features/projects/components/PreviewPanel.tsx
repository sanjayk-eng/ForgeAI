import { AlertCircle, ExternalLink, Loader2, RefreshCw } from "lucide-react";
import { useEffect, useState } from "react";
import { getProjectPreview, type PreviewInfo } from "../../../api/preview.api";

type PreviewPanelProps = {
  projectId: string;
  accessToken: string | null;
};

export function PreviewPanel({ projectId, accessToken }: PreviewPanelProps) {
  const [preview, setPreview] = useState<PreviewInfo | null>(null);
  const [requestFailed, setRequestFailed] = useState(false);
  const [requestVersion, setRequestVersion] = useState(0);
  const [frameVersion, setFrameVersion] = useState(0);
  const [loadedFrame, setLoadedFrame] = useState("");

  useEffect(() => {
    let active = true;
    let timer: ReturnType<typeof setTimeout> | undefined;

    async function refresh() {
      if (!accessToken) {
        setRequestFailed(true);
        return;
      }
      try {
        const result = await getProjectPreview(accessToken, projectId);
        if (!active) return;
        setPreview(result);
        setRequestFailed(false);
      } catch {
        if (active) setRequestFailed(true);
      } finally {
        if (active) timer = setTimeout(() => void refresh(), 5000);
      }
    }

    void refresh();
    return () => {
      active = false;
      if (timer) clearTimeout(timer);
    };
  }, [accessToken, projectId, requestVersion]);

  const status = requestFailed ? "gateway_error" : preview?.status ?? "starting";
  const message = status === "running"
    ? "Preview running"
    : status === "stopped"
      ? "No running application"
      : status === "sandbox_unavailable"
        ? "Sandbox unavailable"
        : status === "application_unavailable"
          ? "Application is not running"
          : status === "gateway_error"
            ? "Unable to load preview"
            : "Starting preview...";
  const previewUrl = status === "running" ? preview?.url ?? null : null;
  const frameKey = previewUrl ? `${previewUrl}:${frameVersion}` : "";

  function retry() {
    setRequestVersion((current) => current + 1);
  }

  function reloadFrame() {
    setLoadedFrame("");
    setFrameVersion((current) => current + 1);
  }

  return (
    <section className="flex h-full min-h-0 min-w-0 flex-1 flex-col bg-forge-bg text-forge-text">
      <header className="flex h-11 shrink-0 items-center gap-2 border-b border-[var(--border)] bg-forge-panel px-3">
        <span className="mr-auto text-xs font-semibold">Preview</span>
        <span className="flex items-center gap-1.5 text-[11px] text-forge-muted" aria-live="polite">
          {status === "starting" && <Loader2 size={13} className="animate-spin" />}
          {status !== "starting" && status !== "running" && <AlertCircle size={13} />}
          {message}
        </span>
        <button
          type="button"
          title={previewUrl ? "Reload preview" : "Retry preview status"}
          aria-label={previewUrl ? "Reload preview" : "Retry preview status"}
          onClick={previewUrl ? reloadFrame : retry}
          className="grid size-8 place-items-center rounded text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text disabled:opacity-30"
        >
          <RefreshCw size={14} />
        </button>
        <button
          type="button"
          title="Open preview in a new tab"
          aria-label="Open preview in a new tab"
          disabled={!previewUrl}
          onClick={() => previewUrl && window.open(previewUrl, "_blank", "noopener,noreferrer")}
          className="grid size-8 place-items-center rounded text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text disabled:opacity-30"
        >
          <ExternalLink size={14} />
        </button>
      </header>

      <div className="relative min-h-0 min-w-0 flex-1 bg-white">
        {previewUrl ? (
          <>
            <iframe
              key={frameKey}
              src={previewUrl}
              title="Project Preview"
              onLoad={() => setLoadedFrame(frameKey)}
              onError={() => setRequestFailed(true)}
              className="size-full border-0 bg-white"
            />
            {loadedFrame !== frameKey && (
              <div className="absolute inset-0 grid place-items-center bg-forge-bg text-xs text-forge-muted">
                <span className="inline-flex items-center gap-2"><Loader2 size={14} className="animate-spin" /> Starting preview...</span>
              </div>
            )}
          </>
        ) : (
          <div className="flex h-full items-center justify-center px-4 text-center text-sm text-forge-muted">
            {status === "gateway_error" && (
              <span className="inline-flex items-center gap-2">
                <AlertCircle size={16} className="text-forge-signal" />
                {message}
              </span>
            )}
            {status !== "gateway_error" && message}
          </div>
        )}
      </div>
    </section>
  );
}