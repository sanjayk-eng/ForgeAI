import { Bot, FolderGit2, LayoutDashboard, LogOut, MessagesSquare, Plus, Settings, Users, X } from "lucide-react";
import { NavLink, useLocation } from "react-router-dom";

const items = [
  { label: "Overview", path: "", icon: LayoutDashboard, end: true },
  { label: "Members", path: "/members", icon: Users },
  { label: "Settings", path: "/settings", icon: Settings },
];

const modules = [
  { label: "Agent organization", path: "/projects", query: "agent=1", icon: Bot },
  { label: "Projects", path: "/projects", icon: FolderGit2 },
  { label: "Conversations", icon: MessagesSquare },
];

export function OrganizationSidebar({
  open,
  hasOrganization,
  organizationId,
  onCreateOrganization,
  onClose,
  onSignOut,
}: {
  open: boolean;
  hasOrganization: boolean;
  organizationId?: string;
  onCreateOrganization: () => void;
  onClose: () => void;
  onSignOut: () => void;
}) {
  const location = useLocation();
  const agentMode = location.search.includes("agent=1") || location.search.includes("panel=agent");
  const organizationPath = organizationId
    ? `/organizations/${encodeURIComponent(organizationId)}`
    : "";

  return (
    <>
      <aside
        className={`fixed inset-y-0 left-0 z-20 flex h-screen w-[248px] shrink-0 flex-col overflow-hidden border-r border-[var(--sidebar-border)] bg-forge-panel p-5 pt-8 transition-transform md:sticky md:top-0 md:h-[calc(100vh-72px)] md:translate-x-0 ${open ? "translate-x-0" : "-translate-x-full"}`}
      >
        <div className="mb-9 flex items-center justify-between md:hidden">
          <span className="font-bold text-forge-text">organization</span>
          <button
            className="grid size-9 place-items-center rounded-md text-forge-muted transition hover:bg-[var(--surface-hover)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent"
            onClick={onClose}
            aria-label="Close navigation"
          >
            <X size={18} />
          </button>
        </div>
        {hasOrganization ? (
          <>
            <span className="px-3 pb-3 font-mono text-[10px] uppercase tracking-[.12em] text-forge-muted">
              organization
            </span>
            <nav className="grid gap-1" aria-label="organization navigation">
              {items.map(({ label, path, icon: Icon, end }) => (
                <NavLink
                  key={path}
                  to={`${organizationPath}${path}`}
                  end={end}
                  onClick={onClose}
                  className={({ isActive }) =>
                    `flex items-center gap-3 rounded-md border-l-2 px-3 py-2.5 text-[13px] font-bold no-underline transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent ${isActive ? "border-forge-accent bg-forge-accent/[0.09] text-forge-text" : "border-transparent text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text"}`
                  }
                >
                  <Icon size={17} />
                  <span>{label}</span>
                </NavLink>
              ))}
            </nav>
            <span className="mb-3 mt-9 px-3 font-mono text-[10px] uppercase tracking-[.12em] text-forge-muted">
              Modules
            </span>
            <div className="grid gap-1 text-[13px] font-semibold text-forge-muted">
              {modules.map(({ label, path, query, icon: Icon }) => path ? (
                <NavLink
                  key={label}
                  to={`${organizationPath}${path}${query ? `?${query}` : ""}`}
                  onClick={onClose}
                  className={({ isActive }) => {
                    const selected = isActive && (label === "Agent organization" ? agentMode : !agentMode);
                    return `flex items-center gap-3 rounded-md px-3 py-2 text-[13px] font-semibold no-underline transition ${selected ? "bg-forge-accent/[0.09] text-forge-text" : "text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text"}`;
                  }}
                >
                  <Icon size={17} />
                  {label}
                </NavLink>
              ) : (
                <span
                  key={label}
                  className="flex items-center gap-3 px-3 py-2"
                >
                  <Icon size={17} className="text-forge-muted" />
                  {label}
                </span>
              ))}
            </div>
          </>
        ) : (
          <div className="mt-2 border border-dashed border-forge-accent/30 bg-forge-accent/[0.05] p-4">
            <div className="mb-3 grid size-9 place-items-center rounded-md border border-forge-accent/30 bg-forge-accent/[0.1] text-forge-accent"><Plus size={17} /></div>
            <p className="text-sm font-bold text-forge-text">No organization yet</p>
            <p className="mt-1 text-xs leading-5 text-forge-muted">Create a organization to unlock team navigation.</p>
            <button type="button" className="mt-4 inline-flex w-full items-center justify-center gap-2 bg-forge-accent px-3 py-2.5 text-xs font-extrabold text-forge-bg transition hover:bg-forge-accent-strong" onClick={onCreateOrganization}><Plus size={14} /> Add organization</button>
          </div>
        )}
        <button
          className="mt-auto flex items-center gap-3 border-t border-[var(--border)] px-3 pt-5 text-left text-[13px] font-bold text-forge-muted transition hover:text-forge-text focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent"
          onClick={onSignOut}
        >
          <LogOut size={17} />
          Sign out
        </button>
      </aside>
      {open && (
        <button
          className="fixed inset-0 z-10 bg-black/60 md:hidden"
          aria-label="Close navigation"
          onClick={onClose}
        />
      )}
    </>
  );
}
