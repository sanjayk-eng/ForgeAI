import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useAuth } from "../../../auth/useAuth";
import { useToast } from "../../../../shared/ui/useToast";
import { listWorkspaces, updateWorkspace } from "../../api/workspace.api";
import { useWorkspaceId } from "../../hooks/useWorkspaceId";
import { Monitor, Moon, Sun } from "lucide-react";
import { useTheme, type Theme } from "../../../../shared/ui/themeContextStore";

export function SettingsPage() {
  const id = useWorkspaceId();
  const { tokens } = useAuth();
  const toast = useToast();
  const { theme, resolvedTheme, setTheme } = useTheme();
  const client = useQueryClient();
  const query = useQuery({
    queryKey: ["workspaces"],
    queryFn: () => listWorkspaces(tokens?.access_token ?? ""),
    enabled: Boolean(tokens?.access_token),
  });
  const workspace = query.data?.find((item) => item.id === id);
  const [name, setName] = useState("");
  const value = name || workspace?.name || "";
  const mutation = useMutation({
    mutationFn: () =>
      updateWorkspace(tokens?.access_token ?? "", id, value.trim()),
    onSuccess: (updated) =>
      client.setQueryData(["workspaces"], (current: typeof query.data) =>
        current?.map((item) => (item.id === updated.id ? updated : item)),
      ),
    onError: () => toast.pushError("Could not save workspace settings"),
  });
  return (
    <div className="animate-page-enter">
      <div className="mb-7 sm:mb-10">
        <span className="mb-2 block font-mono text-[10px] uppercase tracking-[.12em] text-forge-accent">
          Workspace / Settings
        </span>
        <h1 className="m-0 text-3xl font-extrabold tracking-[-.045em] sm:text-[42px]">
          Workspace settings
        </h1>
        <p className="mt-3 text-sm text-forge-muted">
          Keep the workspace identity clear for everyone on the team.
        </p>
      </div>
      <section className="max-w-[680px] border border-white/[0.09] bg-forge-panel/75 p-7">
        <h2 className="m-0 text-base">General</h2>
        <p className="mt-2 mb-7 text-sm text-forge-muted">
          Update the name used across ForgeAI.
        </p>
        <label
          className="block font-mono text-[11px] text-forge-soft"
          htmlFor="settings-name"
        >
          Workspace name
        </label>
        <input
          id="settings-name"
          className="mt-2 mb-5 w-full max-w-[470px] rounded-md border border-[var(--border)] bg-[var(--input)] p-3 text-forge-text outline-none transition focus:border-forge-accent focus:ring-2 focus:ring-forge-accent/15"
          value={value}
          onChange={(event) => setName(event.target.value)}
        />
        <br />
        <button
          className="rounded-md bg-forge-accent px-4 py-3 text-xs font-extrabold text-[var(--primary-foreground)] hover:bg-forge-accent-strong disabled:cursor-not-allowed disabled:opacity-45"
          disabled={
            mutation.isPending ||
            value.trim().length < 2 ||
            value === workspace?.name
          }
          onClick={() => mutation.mutate()}
        >
          {mutation.isPending ? "Saving..." : "Save changes"}
        </button>
      </section>
      <section className="mt-5 max-w-[680px] rounded-xl border border-[var(--border)] bg-forge-panel/80 p-6 shadow-[0_16px_40px_rgba(0,0,0,.08)] sm:p-7">
        <div className="border-b border-[var(--border)] pb-5">
          <h2 className="m-0 text-base font-bold text-forge-text">Appearance</h2>
          <p className="mt-2 text-sm text-forge-muted">Customize how the ForgeAI dashboard looks. System follows your device preference.</p>
        </div>
        <div className="pt-5">
          <p className="m-0 text-sm font-bold text-forge-soft">Theme</p>
          <div className="mt-3 grid gap-3 sm:grid-cols-3">
            {([
              ["dark", "Dark", "Deep charcoal surfaces", Moon],
              ["light", "Light", "Bright, focused workspace", Sun],
              ["system", "System", "Follow your device", Monitor],
            ] as const).map(([value, label, description, Icon]) => (
              <button
                key={value}
                className={`relative min-h-[142px] rounded-lg border p-4 text-left transition hover:-translate-y-0.5 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-forge-accent ${theme === value ? "border-forge-accent/70 bg-forge-accent/[0.11] shadow-[0_8px_24px_rgba(198,243,106,.08)]" : "border-[var(--border)] bg-[var(--surface-subtle)] hover:border-forge-accent/35 hover:bg-[var(--surface-hover)]"}`}
                onClick={() => setTheme(value as Theme)}
                aria-pressed={theme === value}
              >
                <span className={`grid size-9 place-items-center rounded-md ${theme === value ? "bg-forge-accent text-[var(--primary-foreground)]" : "bg-[var(--surface-strong)] text-forge-soft"}`}><Icon size={16} /></span>
                <span className="mt-3 block text-sm font-bold text-forge-text">{label}</span>
                <span className="mt-1 block text-[11px] leading-4 text-forge-muted">{value === "system" ? `${description}. Currently using ${resolvedTheme}.` : description}</span>
                <span className={`mt-3 block text-[10px] font-bold uppercase tracking-wide ${theme === value ? "text-forge-accent" : "text-transparent"}`}>{value === "system" && theme === value ? `Active · ${resolvedTheme}` : "Selected"}</span>
              </button>
            ))}
          </div>
        </div>
      </section>
    </div>
  );
}
