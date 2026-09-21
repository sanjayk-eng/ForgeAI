import {
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Folder,
  GitBranch,
  LifeBuoy,
  Menu,
  MoreHorizontal,
  PanelLeft,
  PanelRight,
  Rocket,
  Search,
  Sparkles,
  type LucideIcon,
} from "lucide-react";
import {
  iconButtonClass,
  smallButtonClass,
  workspaceSections,
  type WorkspaceTab,
} from "./workspaceData";

export function TopBar({ name, email, initials, onSignOut, onMenu, onAssistant }: { name: string; email: string; initials: string; onSignOut: () => void; onMenu: () => void; onAssistant: () => void }) {
  return (
    <header className="flex h-14 shrink-0 items-center justify-between border-b border-white/10 bg-forge-panel px-3 lg:px-4">
      <div className="flex items-center gap-2 lg:gap-3">
        <button className={`${iconButtonClass} lg:hidden`} title="Open navigation" aria-label="Open navigation" onClick={onMenu}><Menu size={17} /></button>
        <span className="grid size-7 place-items-center rounded-md border border-forge-accent/50 bg-forge-accent/10 font-mono text-[10px] font-bold tracking-[-0.12em] text-forge-accent">F/</span>
        <div className="hidden items-center gap-2 text-xs sm:flex"><span className="font-bold">ForgeAI</span><ChevronRight size={13} className="text-forge-muted" /><span className="text-forge-soft">Acme workspace</span><ChevronDown size={13} className="text-forge-muted" /><span className="text-forge-muted">SaaS Dashboard</span></div>
      </div>
      <div className="hidden items-center gap-2 text-xs text-forge-muted md:flex"><button className={`${smallButtonClass} w-52 justify-start text-forge-muted`}><Search size={14} /> Search files and actions <span className="ml-auto font-mono text-[10px]">⌘K</span></button><span className="flex items-center gap-1.5 px-2 text-[11px]"><GitBranch size={14} className="text-forge-accent" /> Connected</span><button className={smallButtonClass}>Share</button><button className="inline-flex items-center gap-2 rounded-md bg-forge-accent px-3 py-2 text-xs font-bold text-[#111419] hover:bg-forge-signal"><Rocket size={14} /> Deploy</button></div>
      <div className="flex items-center gap-2"><button className={`${iconButtonClass} lg:hidden`} title="Open agent" aria-label="Open agent" onClick={onAssistant}><Sparkles size={16} /></button><div className="hidden text-right sm:block"><p className="m-0 text-[11px] font-bold">{name}</p><p className="m-0 max-w-[140px] truncate text-[10px] text-forge-muted">{email}</p></div><button className="grid size-8 place-items-center rounded-full border border-forge-accent/45 bg-forge-accent/10 font-mono text-[10px] font-bold text-forge-accent" title="Sign out" onClick={onSignOut}>{initials}</button></div>
    </header>
  );
}

export function WorkspaceSidebar({ active, onChange, onCollapse }: { active: string; onChange: (label: string) => void; onCollapse: () => void }) {
  return (
    <aside className="hidden w-56 shrink-0 flex-col border-r border-white/10 bg-forge-panel md:flex">
      <div className="flex h-11 items-center justify-between border-b border-white/10 px-3"><span className="font-mono text-[10px] uppercase tracking-[0.08em] text-forge-muted">Workspace</span><button className={iconButtonClass} title="Collapse sidebar" aria-label="Collapse sidebar" onClick={onCollapse}><ChevronLeft size={15} /></button></div>
      <nav className="min-h-0 flex-1 overflow-y-auto p-2">{workspaceSections.map((section) => <div key={section.title} className="mb-5"><p className="px-2 py-2 font-mono text-[10px] uppercase tracking-[0.08em] text-forge-muted">{section.title}</p>{section.items.map(({ label, icon: Icon }) => <button key={label} className={`flex w-full items-center gap-2 rounded-md px-2 py-2 text-left text-xs transition ${active === label ? "bg-forge-accent/10 font-semibold text-forge-accent" : "text-forge-soft hover:bg-white/[0.05]"}`} onClick={() => onChange(label)}><Icon size={15} />{label}</button>)}</div>)}</nav>
      <div className="border-t border-white/10 p-3"><button className="flex items-center gap-2 text-xs text-forge-muted hover:text-forge-soft"><LifeBuoy size={15} /> Help center</button></div>
    </aside>
  );
}

export function WorkspaceHeader({ activeNav, onSearch }: { activeNav: string; onSearch: () => void }) {
  return <div className="flex h-12 shrink-0 items-center justify-between border-b border-white/10 px-4"><div><p className="m-0 text-[11px] font-semibold text-forge-muted">SaaS Dashboard / {activeNav}</p><h1 className="m-0 text-sm font-bold">{activeNav === "Overview" ? "Project overview" : activeNav}</h1></div><div className="flex items-center gap-2"><button className={smallButtonClass} onClick={onSearch}><Search size={14} /> Find</button><button className={iconButtonClass} title="More actions" aria-label="More actions"><MoreHorizontal size={16} /></button></div></div>;
}

export function WorkspaceTabs({ active, onChange }: { active: WorkspaceTab; onChange: (tab: WorkspaceTab) => void }) {
  return <div className="flex h-11 shrink-0 items-center gap-1 border-b border-white/10 px-4">{(["preview", "code", "components", "console"] as WorkspaceTab[]).map((tab) => <button key={tab} className={`px-3 py-2 text-xs font-semibold capitalize ${active === tab ? "border-b-2 border-forge-accent text-forge-accent" : "text-forge-muted hover:text-forge-soft"}`} onClick={() => onChange(tab)}>{tab}</button>)}</div>;
}

export function MobileNav({ onFiles, onAgent }: { onFiles: () => void; onAgent: () => void }) {
  return <div className="flex shrink-0 items-center justify-around border-t border-white/10 bg-forge-panel p-2 md:hidden"><button className="flex flex-col items-center gap-1 text-[10px] text-forge-accent"><PanelLeft size={16} />Home</button><button className="flex flex-col items-center gap-1 text-[10px] text-forge-muted" onClick={onFiles}><Folder size={16} />Files</button><button className="flex flex-col items-center gap-1 text-[10px] text-forge-muted" onClick={onAgent}><PanelRight size={16} />Agent</button></div>;
}

export type { LucideIcon };
