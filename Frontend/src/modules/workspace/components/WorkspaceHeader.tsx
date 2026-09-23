import { ChevronDown, FolderKanban, Menu } from "lucide-react";
import { Link } from "react-router-dom";
import type { Workspace } from "../types/workspace.types";
import { NotificationDropdown } from "./NotificationDropdown";

type Props = {
  userName?: string;
  workspaces: Workspace[];
  selected?: Workspace;
  onSelect: (workspace: Workspace) => void;
  onMenu: () => void;
  accessToken?: string;
};

export function WorkspaceHeader({
  userName,
  workspaces,
  selected,
  onSelect,
  onMenu,
  accessToken,
}: Props) {
  return (
    <header className="flex h-[68px] shrink-0 items-center gap-4 border-b border-white/[0.08] bg-forge-bg/90 px-4 sm:gap-7 sm:px-7">
      <Link
        className="flex min-w-0 items-center gap-2.5 font-extrabold tracking-[-0.03em] text-forge-text no-underline sm:min-w-[184px]"
        to="/workspace"
        aria-label="ForgeAI workspace home"
      >
        <span className="grid size-7 shrink-0 place-items-center rounded-md bg-forge-accent font-mono text-sm font-extrabold text-forge-bg">
          F
        </span>
        <span className="hidden sm:block">
          Forge<span className="text-forge-accent">AI</span>
        </span>
      </Link>
      <div className="flex min-w-0 items-center gap-2 text-forge-soft">
        <FolderKanban size={15} />
        {workspaces.length === 0 ? (
          <span className="text-sm font-bold text-forge-muted">
            No workspace yet
          </span>
        ) : (
          <>
            <select
              className="max-w-[150px] truncate bg-transparent py-1 text-sm font-bold text-forge-text outline-none sm:max-w-[220px]"
              aria-label="Select workspace"
              value={selected?.id ?? ""}
              onChange={(event) => {
                const workspace = workspaces.find(
                  (item) => item.id === event.target.value,
                );
                if (workspace) onSelect(workspace);
              }}
            >
              {workspaces.map((workspace) => (
                <option key={workspace.id} value={workspace.id}>
                  {workspace.name}
                </option>
              ))}
            </select>
            <ChevronDown size={14} />
          </>
        )}
      </div>
      <div className="ml-auto flex items-center gap-2.5">
        {accessToken && <NotificationDropdown accessToken={accessToken} />}
        <div className="hidden items-center gap-2 text-sm font-bold text-forge-soft sm:flex">
          <span className="grid size-8 place-items-center rounded-full bg-forge-signal font-mono text-xs text-forge-bg">
            {userName?.slice(0, 1).toUpperCase() ?? "U"}
          </span>
          {userName ?? "User"}
        </div>
        <button
          className="grid size-9 place-items-center rounded-md border border-transparent text-forge-muted transition hover:bg-white/[0.06] hover:text-forge-text focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent md:hidden"
          aria-label="Open navigation"
          onClick={onMenu}
        >
          <Menu size={19} />
        </button>
      </div>
    </header>
  );
}
