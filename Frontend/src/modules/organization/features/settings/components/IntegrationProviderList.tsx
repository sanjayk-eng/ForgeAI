import { GitBranch, PanelsTopLeft } from "lucide-react";

export function IntegrationProviderList() {
  return (
    <nav className="grid content-start gap-1 border-r border-[var(--border)] p-3" aria-label="Integration providers">
      <button
        type="button"
        className="flex min-h-12 items-center gap-3 border border-forge-accent/20 bg-forge-accent/[0.08] px-3 text-left text-sm font-bold text-forge-text"
        aria-current="page"
      >
        <GitBranch size={19} /> GitHub
      </button>
      <div className="flex min-h-12 items-center gap-3 px-3 text-sm text-forge-muted">
        <GitBranch size={19} /> GitLab
        <span className="ml-auto font-mono text-[9px] uppercase text-forge-soft">Coming soon</span>
      </div>
      <div className="flex min-h-12 items-center gap-3 px-3 text-sm text-forge-muted">
        <PanelsTopLeft size={19} /> Bitbucket
        <span className="ml-auto font-mono text-[9px] uppercase text-forge-soft">Coming soon</span>
      </div>
    </nav>
  );
}