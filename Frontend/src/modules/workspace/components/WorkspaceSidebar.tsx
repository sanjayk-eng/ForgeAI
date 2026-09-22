import { LayoutDashboard, LogOut, Settings, Users, X } from "lucide-react";
import { NavLink } from "react-router-dom";

const items = [
  { label: "Overview", to: "/workspace", icon: LayoutDashboard, end: true },
  { label: "Members", to: "/workspace/members", icon: Users },
  { label: "Settings", to: "/workspace/settings", icon: Settings },
];

export function WorkspaceSidebar({
  open,
  hasWorkspace,
  workspaceId,
  onClose,
  onSignOut,
}: {
  open: boolean;
  hasWorkspace: boolean;
  workspaceId?: string;
  onClose: () => void;
  onSignOut: () => void;
}) {
  return (
    <>
      <aside
        className={`fixed inset-y-0 left-0 z-20 flex w-[248px] flex-col border-r border-white/[0.08] bg-[#111419] p-5 pt-8 transition-transform md:static md:z-auto md:translate-x-0 ${open ? "translate-x-0" : "-translate-x-full"}`}
      >
        <div className="mb-9 flex items-center justify-between md:hidden">
          <span className="font-bold text-forge-text">Workspace</span>
          <button
            className="grid size-9 place-items-center rounded-md text-forge-muted transition hover:bg-white/[0.06] focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent"
            onClick={onClose}
            aria-label="Close navigation"
          >
            <X size={18} />
          </button>
        </div>
        {hasWorkspace ? (
          <>
            <span className="px-3 pb-3 font-mono text-[10px] uppercase tracking-[.12em] text-forge-muted">
              Workspace
            </span>
            <nav className="grid gap-1" aria-label="Workspace navigation">
              {items.map(({ label, to, icon: Icon, end }) => (
                <NavLink
                  key={to}
                  to={`${to}?workspace=${workspaceId ?? ""}`}
                  end={end}
                  onClick={onClose}
                  className={({ isActive }) =>
                    `flex items-center gap-3 rounded-md border-l-2 px-3 py-2.5 text-[13px] font-bold no-underline transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent ${isActive ? "border-forge-accent bg-forge-accent/[0.09] text-forge-text" : "border-transparent text-forge-muted hover:bg-white/[0.05] hover:text-forge-text"}`
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
              <span className="flex items-center gap-3 px-3 py-2">
                <i className="size-1.5 rounded-full bg-forge-accent" />
                Agent workspace
              </span>
              <span className="flex items-center gap-3 px-3 py-2">
                <i className="size-1.5 rounded-full bg-forge-signal" />
                Repositories
              </span>
              <span className="flex items-center gap-3 px-3 py-2">
                <i className="size-1.5 rounded-full bg-[#8ed4ff]" />
                Conversations
              </span>
            </div>
          </>
        ) : (
          <div className="mt-2 border-l-2 border-forge-accent/40 px-3 py-3 text-xs leading-5 text-forge-muted">
            Create a workspace to unlock team navigation.
          </div>
        )}
        <button
          className="mt-auto flex items-center gap-3 border-t border-white/[0.07] px-3 pt-5 text-left text-[13px] font-bold text-forge-muted transition hover:text-forge-text focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent"
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
