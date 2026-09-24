import { useQuery } from "@tanstack/react-query";
import {
  ArrowUpRight,
  Bot,
  Check,
  Clock3,
  FolderGit2,
  Settings2,
  Users,
  Workflow,
} from "lucide-react";
import { Link } from "react-router-dom";
import { useAuth } from "../../../auth/useAuth";
import { listMembers } from "../../api/members.api";
import { listWorkspaces } from "../../api/workspace.api";
import { WorkspaceStatCard } from "../../components/WorkspaceStatCard";
import { useWorkspaceId } from "../../hooks/useWorkspaceId";

const modules = [
  {
    icon: Bot,
    name: "Agent workspace",
    detail: "Run and organize agent workflows",
    state: "Ready",
  },
  {
    icon: FolderGit2,
    name: "Repositories",
    detail: "Connect code and project context",
    state: "Coming next",
  },
  {
    icon: Workflow,
    name: "Conversations",
    detail: "Keep team decisions in one place",
    state: "Coming next",
  },
];

export function OverviewPage() {
  const id = useWorkspaceId();
  const { user, tokens } = useAuth();
  const accessToken = tokens?.access_token ?? "";
  const workspacesQuery = useQuery({
    queryKey: ["workspaces"],
    queryFn: () => listWorkspaces(accessToken),
    enabled: Boolean(accessToken),
  });
  const membersQuery = useQuery({
    queryKey: ["members", id],
    queryFn: () => listMembers(accessToken, id),
    enabled: Boolean(accessToken && id),
  });
  const workspace = workspacesQuery.data?.find((item) => item.id === id);
  const memberCount = membersQuery.data?.total ?? 0;
  const firstName = user?.name?.split(" ")[0] ?? "there";

  return (
    <div className="animate-page-enter space-y-5">
      <header className="flex flex-col justify-between gap-5 border-b border-white/[0.08] pb-6 lg:flex-row lg:items-end">
        <div>
          <span className="mb-3 block font-mono text-[10px] uppercase tracking-[.14em] text-forge-accent">
            Workspace overview
          </span>
          <h1 className="m-0 text-3xl font-extrabold tracking-[-.05em] text-forge-text sm:text-[42px]">
            Good to see you, {firstName}.
          </h1>
          <p className="mt-3 max-w-xl text-sm leading-6 text-forge-muted">
            Your team workspace is ready. Keep the important work close and the
            next action obvious.
          </p>
        </div>
        <div className="flex items-center gap-2 self-start border border-forge-accent/20 px-3 py-2 font-mono text-[10px] uppercase tracking-[.1em] text-[#bde986] lg:self-auto">
          <span className="size-1.5 bg-forge-accent shadow-[0_0_9px_#c6f36a]" />
          Operational
        </div>
      </header>

      <section className="grid gap-4 lg:grid-cols-[minmax(0,1.7fr)_minmax(260px,.8fr)]">
        <div className="relative min-h-[245px] overflow-hidden border border-white/[0.09] bg-forge-card p-6 sm:p-8">
          <div className="absolute right-0 top-0 h-full w-1/2 bg-[linear-gradient(115deg,transparent_15%,rgba(198,243,106,.07)_15%,transparent_16%,transparent_35%,rgba(142,212,255,.04)_35%,transparent_36%)] opacity-70" />
          <div className="relative flex h-full flex-col justify-between">
            <div>
              <span className="font-mono text-[10px] uppercase tracking-[.13em] text-forge-muted">
                Current workspace
              </span>
              <h2 className="mt-3 max-w-lg truncate text-2xl font-extrabold tracking-[-.04em] text-forge-text sm:text-3xl">
                {workspace?.name ?? "Workspace"}
              </h2>
              <p className="mt-2 font-mono text-[11px] text-forge-muted">
                /{workspace?.slug ?? "workspace"}
              </p>
            </div>
            <div className="flex flex-wrap items-end justify-between gap-5">
              <div>
                <p className="m-0 text-xs text-forge-soft">Workspace status</p>
                <p className="mt-1 flex items-center gap-2 text-sm font-bold text-forge-text">
                  <Check size={15} className="text-forge-accent" /> All systems
                  ready
                </p>
              </div>
              <p className="m-0 flex items-center gap-2 font-mono text-[10px] text-forge-muted">
                <Clock3 size={13} /> Created{" "}
                {workspace
                  ? new Date(workspace.created_at).toLocaleDateString()
                  : "today"}
              </p>
            </div>
          </div>
        </div>

        <aside className="border border-white/[0.09] bg-forge-panel/75 p-6">
          <div className="flex items-center justify-between border-b border-white/[0.08] pb-4">
            <div>
              <span className="font-mono text-[10px] uppercase tracking-[.13em] text-forge-muted">
                Shortcuts
              </span>
              <h2 className="mt-1 text-base font-bold text-forge-text">
                Move the work forward
              </h2>
            </div>
            <ArrowUpRight size={17} className="text-forge-accent" />
          </div>
          <div className="mt-4 grid gap-2">
            <Link
              className="group flex items-center justify-between border border-white/[0.07] px-3 py-3 text-xs font-bold text-forge-soft no-underline transition hover:border-forge-accent/30 hover:bg-white/[0.03] hover:text-forge-text"
              to={`/workspace/members?workspace=${id}`}
            >
              Invite a team member{" "}
              <ArrowUpRight
                size={14}
                className="text-forge-muted transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5"
              />
            </Link>
            <Link
              className="group flex items-center justify-between border border-white/[0.07] px-3 py-3 text-xs font-bold text-forge-soft no-underline transition hover:border-forge-accent/30 hover:bg-white/[0.03] hover:text-forge-text"
              to={`/workspace/settings?workspace=${id}`}
            >
              Configure workspace{" "}
              <Settings2
                size={14}
                className="text-forge-muted transition-transform group-hover:rotate-45"
              />
            </Link>
          </div>
        </aside>
      </section>

      <section className="grid gap-3 sm:grid-cols-3">
        <WorkspaceStatCard
          icon={Users}
          label="Team members"
          value={membersQuery.isLoading ? "--" : String(memberCount)}
          detail="People with workspace access"
          accent
        />
        <WorkspaceStatCard
          icon={Clock3}
          label="Created"
          value={
            workspace
              ? new Date(workspace.created_at).toLocaleDateString()
              : "--"
          }
          detail="Workspace foundation"
        />
        <WorkspaceStatCard
          icon={FolderGit2}
          label="Workspace key"
          value={workspace?.slug ?? "--"}
          detail="Used across ForgeAI"
        />
      </section>

      <section className="grid gap-4 lg:grid-cols-[minmax(0,1.25fr)_minmax(280px,.75fr)]">
        <div className="border border-white/[0.09] bg-forge-panel/70 p-6">
          <div className="mb-5 flex items-end justify-between border-b border-white/[0.08] pb-4">
            <div>
              <span className="font-mono text-[10px] uppercase tracking-[.13em] text-forge-muted">
                Workspace map
              </span>
              <h2 className="mt-1 text-lg font-bold text-forge-text">
                Your ForgeAI modules
              </h2>
            </div>
            <span className="font-mono text-[10px] text-forge-muted">
              {modules.filter((module) => module.state === "Ready").length}/
              {modules.length} active
            </span>
          </div>
          <div className="grid gap-1">
            {modules.map(({ icon: Icon, name, detail, state }) => (
              <div
                key={name}
                className="flex items-center gap-3 border-b border-white/[0.06] py-4 last:border-0"
              >
                <span className="grid size-9 shrink-0 place-items-center border border-white/[0.1] text-forge-soft">
                  <Icon size={17} />
                </span>
                <div className="min-w-0 flex-1">
                  <p className="m-0 truncate text-sm font-bold text-forge-text">
                    {name}
                  </p>
                  <p className="mt-1 truncate text-xs text-forge-muted">
                    {detail}
                  </p>
                </div>
                <span
                  className={`shrink-0 font-mono text-[10px] ${state === "Ready" ? "text-[#bde986]" : "text-forge-muted"}`}
                >
                  {state}
                </span>
              </div>
            ))}
          </div>
        </div>
        <div className="border border-white/[0.09] bg-forge-panel/70 p-6">
          <span className="font-mono text-[10px] uppercase tracking-[.13em] text-forge-muted">
            Getting started
          </span>
          <h2 className="mt-1 text-lg font-bold text-forge-text">
            Build your foundation
          </h2>
          <div className="mt-6 h-1 bg-white/[0.08]">
            <div className="h-1 w-1/3 bg-forge-accent" />
          </div>
          <p className="mt-3 text-xs leading-5 text-forge-muted">
            One of three setup steps complete. Invite your team to keep momentum
            moving.
          </p>
          <Link
            className="mt-6 inline-flex items-center gap-2 text-xs font-bold text-forge-accent no-underline"
            to={`/workspace/members?workspace=${id}`}
          >
            Continue setup <ArrowUpRight size={14} />
          </Link>
        </div>
      </section>
    </div>
  );
}
