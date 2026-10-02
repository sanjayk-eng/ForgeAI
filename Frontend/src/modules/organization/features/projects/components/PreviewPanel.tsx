import { AlertCircle, ExternalLink, Loader2, RefreshCw } from "lucide-react";
import { useEffect, useState } from "react";
import { useProjectPreview } from "../hooks/useProjectPreview";

type PreviewPanelProps = {
  projectId: string;
  accessToken: string | null;
};

export function PreviewPanel({ projectId, accessToken }: PreviewPanelProps) {
  const { data: preview, isError, refetch, waitTimedOut } = useProjectPreview(accessToken, projectId);
  const [frameFailed, setFrameFailed] = useState(false);
  const [frameTimedOut, setFrameTimedOut] = useState(false);
  const [frameVersion, setFrameVersion] = useState(0);
  const [loadedFrame, setLoadedFrame] = useState("");

  const status = isError || frameFailed || !accessToken ? "gateway_error" : preview?.status ?? "starting";
  const message = status === "running"
    ? "Preview running"
    : status === "stopped"
      ? "Sandbox stopped"
      : status === "sandbox_unavailable"
        ? "Sandbox unavailable"
        : status === "application_unavailable"
          ? waitTimedOut ? "No project server detected. Start it in Terminal." : "Detecting project server..."
          : status === "starting" && waitTimedOut
            ? "Sandbox is taking longer than expected."
          : status === "gateway_error"
            ? "Unable to load preview"
            : "Starting sandbox...";
  const previewUrl = status === "running" ? preview?.url ?? null : null;
  const containerPort = status === "running" ? preview?.container_port ?? null : null;
  const frameKey = previewUrl ? `${previewUrl}:${frameVersion}` : "";

  useEffect(() => {
    setFrameTimedOut(false);
    if (!previewUrl || loadedFrame === frameKey) return;
    const timer = setTimeout(() => setFrameTimedOut(true), 15_000);
    return () => clearTimeout(timer);
  }, [frameKey, loadedFrame, previewUrl]);

  function retry() {
    setFrameFailed(false);
    void refetch();
  }

  function reloadFrame() {
    setLoadedFrame("");
    setFrameTimedOut(false);
    setFrameVersion((current) => current + 1);
  }

  return (
    <section className="flex h-full min-h-0 min-w-0 flex-1 flex-col bg-forge-bg text-forge-text">
      <header className="flex h-11 shrink-0 items-center gap-2 border-b border-[var(--border)] bg-forge-panel px-3">
        <span className="mr-auto text-xs font-semibold">Preview</span>
        {containerPort !== null && (
          <span className="text-[11px] text-forge-muted" title="Detected port inside the sandbox">
            Container port {containerPort}
          </span>
        )}
        <span className="flex items-center gap-1.5 text-[11px] text-forge-muted" aria-live="polite">
          {(status === "starting" || status === "application_unavailable") && !waitTimedOut && <Loader2 size={13} className="animate-spin" />}
          {status !== "starting" && status !== "running" && status !== "application_unavailable" && <AlertCircle size={13} />}
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
              title="ForgeAI Preview"
              onLoad={() => setLoadedFrame(frameKey)}
              onError={() => setFrameFailed(true)}
              className="size-full border-0 bg-white"
            />
            {loadedFrame !== frameKey && (
              <div className="absolute inset-0 grid place-items-center bg-forge-bg text-xs text-forge-muted">
                <span className="inline-flex items-center gap-2">
                  {frameTimedOut ? <AlertCircle size={14} className="text-forge-signal" /> : <Loader2 size={14} className="animate-spin" />}
                  {frameTimedOut ? "Preview is taking too long. Retry or reload it." : "Connecting to preview..."}
                </span>
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