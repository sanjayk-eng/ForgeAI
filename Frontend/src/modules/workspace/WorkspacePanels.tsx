import { useRef, useState } from "react";
import {
  ArrowUp,
  Check,
  CircleDot,
  Copy,
  ExternalLink,
  FileCode2,
  Layers3,
  Plus,
  RefreshCw,
  Send,
  Sparkles,
  X,
  Zap,
} from "lucide-react";
import {
  homeCode,
  iconButtonClass,
  projectFiles,
  smallButtonClass,
  type ChatMessage,
  type DeviceMode,
  type WorkspaceFile,
} from "./workspaceData";

const components = ["Navbar", "Hero", "FeatureGrid", "PricingCard", "Footer", "Modal"];

type Status = "idle" | "working" | "done";

export function PreviewPanel({ device, setDevice }: { device: DeviceMode; setDevice: (device: DeviceMode) => void }) {
  const [refreshing, setRefreshing] = useState(false);
  const previewRef = useRef<HTMLDivElement>(null);

  function refreshPreview() {
    setRefreshing(true);
    window.setTimeout(() => setRefreshing(false), 650);
  }

  function openPreview() {
    window.open("http://localhost:5173/preview", "_blank", "noopener,noreferrer");
  }

  function fullscreenPreview() {
    void previewRef.current?.requestFullscreen?.();
  }

  const widthClass = device === "desktop" ? "w-full" : device === "tablet" ? "w-[680px] max-w-full" : "w-[390px] max-w-full";

  return (
    <div className="flex min-h-0 flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <p className="m-0 font-mono text-[10px] uppercase tracking-[0.08em] text-forge-accent">Preview</p>
          <p className="m-0 mt-1 text-xs text-forge-muted">{refreshing ? "Refreshing preview..." : "Live application preview"}</p>
        </div>
        <div className="flex items-center gap-1 rounded-md border border-white/10 bg-forge-panel p-1">
          {(["desktop", "tablet", "mobile"] as DeviceMode[]).map((item) => (
            <button key={item} className={`rounded px-2 py-1 text-[11px] capitalize ${device === item ? "bg-white/10 text-[#f1f3f5]" : "text-forge-muted"}`} onClick={() => setDevice(item)}>{item}</button>
          ))}
          <button className={`${iconButtonClass} ${refreshing ? "animate-spin" : ""}`} title="Refresh preview" aria-label="Refresh preview" onClick={refreshPreview}><RefreshCw size={14} /></button>
          <button className={iconButtonClass} title="Open preview" aria-label="Open preview" onClick={openPreview}><ExternalLink size={14} /></button>
          <button className={iconButtonClass} title="Fullscreen preview" aria-label="Fullscreen preview" onClick={fullscreenPreview}><span className="text-xs">⛶</span></button>
        </div>
      </div>

      <div ref={previewRef} className="min-h-0 max-h-[calc(100vh-260px)] overflow-auto overscroll-contain rounded-lg border border-white/10 bg-[#0b0e12] p-5">
        <div className={`min-h-[560px] shrink-0 overflow-hidden rounded-lg border border-white/15 bg-[#181d23] shadow-2xl transition-all ${widthClass}`}>
          <div className="flex h-9 items-center gap-1.5 border-b border-white/10 bg-[#111419] px-3">
            <span className="size-2 rounded-full bg-red-300/60" /><span className="size-2 rounded-full bg-forge-signal/70" /><span className="size-2 rounded-full bg-forge-accent/70" />
            <span className="ml-3 flex-1 rounded bg-white/[0.05] px-2 py-1 font-mono text-[9px] text-forge-muted">localhost:5173 / saas-dashboard</span>
          </div>
          <div className="flex min-h-[calc(100%-36px)] flex-col items-center justify-center overflow-y-auto bg-gradient-to-br from-[#1c2930] via-[#17201f] to-[#20291b] p-8 text-center">
            <p className="font-mono text-[10px] uppercase tracking-[0.14em] text-forge-accent">Your next release</p>
            <h2 className="mt-4 max-w-lg text-4xl font-extrabold leading-none tracking-[-0.06em] text-[#f4f6ef]">Build a clearer way to work.</h2>
            <p className="mt-4 max-w-md text-sm leading-6 text-forge-soft">A focused workspace for your team, your product, and the ideas worth shipping.</p>
            <button className="mt-6 rounded-md bg-forge-accent px-4 py-2 text-xs font-bold text-[#111419]">Get started <ArrowUp size={14} className="ml-1 inline rotate-45" /></button>
          </div>
        </div>
      </div>
      <DeploymentCard />
    </div>
  );
}

function DeploymentCard() {
  const [deployStatus, setDeployStatus] = useState("Live");
  const [syncStatus, setSyncStatus] = useState("Connected");

  function deploy() {
    setDeployStatus("Deploying...");
    window.setTimeout(() => setDeployStatus("Live"), 1200);
  }

  function sync(action: "Pulling" | "Pushing") {
    setSyncStatus(`${action}...`);
    window.setTimeout(() => setSyncStatus("Connected"), 900);
  }

  return (
    <div className="grid gap-4 xl:grid-cols-2">
      <div className="rounded-lg border border-white/10 bg-forge-panel p-4">
        <div className="flex items-center justify-between"><p className="m-0 text-xs font-bold">Production</p><span className="flex items-center gap-1.5 text-[11px] text-forge-accent"><span className="size-1.5 rounded-full bg-forge-accent" />{deployStatus}</span></div>
        <p className="mt-4 font-mono text-xs text-forge-soft">https://saas-dashboard.forge.ai</p>
        <p className="mt-2 text-[11px] text-forge-muted">Latest deployment v1.8 · 2 minutes ago</p>
        <div className="mt-4 flex gap-2"><button className={smallButtonClass} onClick={() => window.open("http://localhost:5173/preview", "_blank", "noopener,noreferrer")}><ExternalLink size={13} />Open site</button><button className={smallButtonClass} onClick={deploy}>Deploy</button></div>
      </div>
      <div className="rounded-lg border border-white/10 bg-forge-panel p-4">
        <div className="flex items-center justify-between"><p className="m-0 text-xs font-bold">GitHub</p><span className="flex items-center gap-1.5 text-[11px] text-forge-accent"><Check size={13} />{syncStatus}</span></div>
        <p className="mt-4 text-xs text-forge-soft">my-company / saas-dashboard</p><p className="mt-2 text-[11px] text-forge-muted">main · Last synced 2 minutes ago</p>
        <div className="mt-4 flex gap-2"><button className={smallButtonClass} onClick={() => sync("Pulling")}>Pull changes</button><button className={smallButtonClass} onClick={() => sync("Pushing")}>Push changes</button></div>
      </div>
    </div>
  );
}

export function CodePanel() {
  const [selectedFile, setSelectedFile] = useState(projectFiles[0]);
  const [copied, setCopied] = useState(false);

  async function copyCode() {
    await navigator.clipboard?.writeText(homeCode);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 1200);
  }

  return (
    <div className="grid min-h-full gap-4 lg:grid-cols-[220px_minmax(0,1fr)]">
      <div className="rounded-lg border border-white/10 bg-forge-panel p-3"><div className="mb-3 flex items-center justify-between"><span className="font-mono text-[10px] uppercase text-forge-muted">Files</span><Plus size={14} className="text-forge-muted" /></div>{projectFiles.map((file) => <button key={file} className={`flex w-full items-center gap-2 rounded px-2 py-2 text-left text-xs ${selectedFile === file ? "bg-forge-accent/10 text-forge-accent" : "text-forge-soft hover:bg-white/[0.05]"}`} onClick={() => setSelectedFile(file)}><FileCode2 size={14} />{file}</button>)}</div>
      <div className="flex min-h-[420px] flex-col overflow-hidden rounded-lg border border-white/10 bg-[#0f1216]"><div className="flex items-center justify-between border-b border-white/10 px-3 py-2 text-xs"><span>{selectedFile}</span><button className="flex items-center gap-2 text-forge-muted hover:text-forge-soft" onClick={copyCode}><Copy size={14} />{copied ? "Copied" : "Copy"}</button></div><pre className="m-0 flex-1 overflow-auto p-4 font-mono text-xs leading-6 text-forge-soft">{homeCode}</pre></div>
    </div>
  );
}

export function ComponentsPanel() {
  const [selected, setSelected] = useState("");
  return <div className="grid gap-4 md:grid-cols-3">{components.map((item) => <button key={item} className={`rounded-lg border p-4 text-left transition ${selected === item ? "border-forge-accent/50 bg-forge-accent/10" : "border-white/10 bg-forge-panel hover:border-white/20"}`} onClick={() => setSelected(item)}><div className="mb-6 grid size-9 place-items-center rounded-md bg-forge-accent/10 text-forge-accent"><Layers3 size={17} /></div><p className="m-0 text-sm font-semibold">{item}</p><p className="mt-2 text-xs text-forge-muted">{selected === item ? "Selected for editing" : "Reusable project component"}</p></button>)}</div>;
}

export function ConsolePanel() {
  const [lines, setLines] = useState(["$ forge preview --watch", "Local server started on http://localhost:5173", "No errors found. Waiting for changes..."]);
  return <div className="rounded-lg border border-white/10 bg-[#0f1216] p-4 font-mono text-xs leading-6 text-forge-muted"><div className="mb-3 flex justify-end"><button className={smallButtonClass} onClick={() => setLines([])}>Clear</button></div>{lines.length ? lines.map((line) => <p key={line} className="m-0">{line}</p>) : <p className="m-0 text-forge-accent">Console cleared.</p>}</div>;
}

export function AssistantPanel({ chat, prompt, setPrompt, onSend, onClose }: { chat: ChatMessage[]; prompt: string; setPrompt: (value: string) => void; onSend: () => void; onClose: () => void }) {
  const [contextAdded, setContextAdded] = useState(false);
  const [status, setStatus] = useState<Status>("idle");

  function send() {
    setStatus("working");
    onSend();
    window.setTimeout(() => setStatus("done"), 700);
  }

  return <aside className="flex min-h-0 w-full shrink-0 flex-col border-t border-white/10 bg-forge-panel lg:w-[330px] lg:border-l lg:border-t-0"><div className="flex h-12 items-center justify-between border-b border-white/10 px-4"><div><div className="flex items-center gap-2 text-xs font-bold"><Sparkles size={15} className="text-forge-accent" />AI Agent</div><span className="ml-5 flex items-center gap-1 text-[10px] text-forge-accent"><CircleDot size={10} />{status === "working" ? "Working" : status === "done" ? "Updated" : "Ready"}</span></div><button className={iconButtonClass} title="Close assistant" aria-label="Close assistant" onClick={onClose}><X size={15} /></button></div><div className="border-b border-white/10 px-4 py-3"><div className="flex items-center justify-between"><span className="font-mono text-[10px] uppercase tracking-[0.08em] text-forge-muted">Context</span><button className="text-[10px] font-semibold text-forge-accent" onClick={() => setContextAdded(!contextAdded)}>{contextAdded ? "Remove context" : "+ Add context"}</button></div><div className="mt-3 grid grid-cols-2 gap-2 text-[10px] text-forge-muted"><span>Project <strong className="block text-forge-soft">SaaS Dashboard</strong></span><span>Branch <strong className="block text-forge-soft">main</strong></span><span>Files <strong className="block text-forge-soft">{contextAdded ? "Current file" : "42 files"}</strong></span><span>Environment <strong className="block text-forge-soft">Development</strong></span></div></div><div className="min-h-0 flex-1 space-y-3 overflow-y-auto p-4">{chat.length === 0 ? <div className="mt-8 text-center"><div className="mx-auto grid size-10 place-items-center rounded-lg bg-forge-accent/10 text-forge-accent"><Zap size={17} /></div><p className="mt-3 text-sm font-semibold">What are you building?</p><p className="mt-2 text-xs leading-5 text-forge-muted">Ask for a page, a component, or a change to your project.</p></div> : chat.map((item, index) => <div key={`${item.role}-${index}`} className={item.role === "user" ? "ml-5 rounded-lg bg-forge-accent/10 p-3 text-xs leading-5 text-[#e6efda]" : "rounded-lg border border-white/10 bg-white/[0.03] p-3 text-xs leading-5 text-forge-soft"}><p className="m-0 mb-1 font-mono text-[10px] uppercase tracking-[0.08em] text-forge-muted">{item.role === "user" ? "You" : "AI Agent"}</p>{item.text}</div>)}</div><div className="border-t border-white/10 p-3"><div className="rounded-lg border border-white/10 bg-black/10 p-2 focus-within:border-forge-accent/45"><textarea className="min-h-20 w-full resize-none bg-transparent text-xs leading-5 text-[#f1f3f5] outline-none placeholder:text-forge-muted" placeholder="Describe what you want to build..." value={prompt} onChange={(event) => setPrompt(event.target.value)} onKeyDown={(event) => { if (event.key === "Enter" && !event.shiftKey) { event.preventDefault(); send(); } }} /><div className="mt-2 flex items-center justify-between"><button className="flex items-center gap-1 text-[10px] text-forge-muted hover:text-forge-soft"><Plus size={13} />Add files</button><button className="grid size-8 place-items-center rounded-md bg-forge-accent text-[#111419] transition hover:bg-forge-signal" title="Send prompt" aria-label="Send prompt" onClick={send}><Send size={14} /></button></div></div></div></aside>;
}

export function FilePreview({ file }: { file: WorkspaceFile }) {
  return <pre className="m-0 overflow-auto p-4 font-mono text-xs leading-6 text-forge-soft">{file.code}</pre>;
}
