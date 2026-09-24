import { ChevronDown, Menu, Moon, Sun, Monitor } from "lucide-react";
import { useState } from "react";
import { Link } from "react-router-dom";
import type { Workspace } from "../types/workspace.types";
import { NotificationDropdown } from "./NotificationDropdown";
import { WorkspaceSelector } from "./WorkspaceSelector";
import { Dropdown } from "../../../shared/ui/Dropdown";
import { useTheme, type Theme } from "../../../shared/ui/themeContextStore";

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
  const [profileOpen, setProfileOpen] = useState(false);
  const { theme, resolvedTheme, setTheme } = useTheme();
  const themeOptions: { value: Theme; label: string; icon: typeof Moon }[] = [
    { value: "dark", label: "Dark", icon: Moon },
    { value: "light", label: "Light", icon: Sun },
    { value: "system", label: "System", icon: Monitor },
  ];
  const initial = userName?.slice(0, 1).toUpperCase() ?? "U";

  return (
    <header className="relative z-40 flex h-[72px] shrink-0 items-center gap-4 border-b border-[var(--border)] bg-forge-bg/90 px-4 backdrop-blur-sm sm:gap-7 sm:px-7">
      <Link
        className="flex min-w-0 items-center gap-2.5 font-extrabold tracking-[-0.03em] text-forge-text no-underline sm:min-w-[184px]"
        to="/workspace"
        aria-label="ForgeAI workspace home"
      >
        <span className="grid size-7 shrink-0 place-items-center rounded-md bg-forge-accent font-mono text-sm font-extrabold text-[var(--primary-foreground)]">
          F
        </span>
        <span className="hidden sm:block">
          Forge<span className="text-forge-accent">AI</span>
        </span>
      </Link>
      
      <WorkspaceSelector 
        workspaces={workspaces}
        selected={selected}
        onSelect={onSelect}
      />
      
      <div className="ml-auto flex items-center gap-2">
        {accessToken && <NotificationDropdown accessToken={accessToken} />}
        <div className="relative hidden sm:block">
          <button
            className="flex items-center gap-2 rounded-lg border border-transparent px-2 py-1.5 text-left transition hover:border-[var(--border)] hover:bg-[var(--surface-hover)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent"
            onClick={() => setProfileOpen(!profileOpen)}
            aria-expanded={profileOpen}
            aria-label="Open user menu"
          >
            <span className="grid size-8 place-items-center rounded-full bg-forge-signal font-mono text-xs font-bold text-[var(--primary-foreground)]">{initial}</span>
            <span className="max-w-[150px] truncate text-sm font-bold text-forge-soft">{userName ?? "User"}</span>
            <ChevronDown size={14} className="text-forge-muted" />
          </button>
          <Dropdown isOpen={profileOpen} onClose={() => setProfileOpen(false)} className="w-64 overflow-hidden" align="right">
            <div className="border-b border-[var(--border)] px-4 py-3">
              <p className="text-sm font-bold text-forge-text">Appearance</p>
              <p className="mt-1 text-xs text-forge-muted">System follows your device preference.</p>
            </div>
            <div className="p-2">
              {themeOptions.map(({ value, label, icon: Icon }) => (
                <button
                  key={value}
                  className={`flex w-full items-center gap-3 rounded-md px-3 py-2.5 text-left text-sm transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent ${theme === value ? "bg-forge-accent/[0.12] text-forge-text" : "text-forge-soft hover:bg-[var(--surface-hover)]"}`}
                  onClick={() => setTheme(value)}
                  aria-pressed={theme === value}
                >
                  <Icon size={16} />
                  <span className="flex-1">{label}</span>
                  {value === "system" && <span className="text-[10px] text-forge-muted">{resolvedTheme}</span>}
                  {theme === value && <span className="size-1.5 rounded-full bg-forge-accent" />}
                </button>
              ))}
            </div>
          </Dropdown>
        </div>
        <button
          className="grid size-9 place-items-center rounded-md border border-transparent text-forge-muted transition hover:bg-[var(--surface-hover)] hover:text-forge-text focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent md:hidden"
          aria-label="Open navigation"
          onClick={onMenu}
        >
          <Menu size={19} />
        </button>
      </div>
    </header>
  );
}
