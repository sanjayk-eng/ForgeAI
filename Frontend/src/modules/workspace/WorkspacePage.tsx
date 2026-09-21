import { ArrowRight, BriefcaseBusiness, LogOut, Sparkles, Workflow } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { getCurrentUser } from "../auth/api";
import { useAuth } from "../auth/useAuth";

const cardClass =
  "min-h-[250px] rounded-[18px] border border-white/15 bg-forge-card/75 p-[22px] shadow-[0_24px_60px_rgba(0,0,0,0.28)] transition duration-300 hover:-translate-y-1 hover:border-forge-accent/35 hover:bg-forge-card";
const iconClass =
  "mb-[22px] grid size-[38px] place-items-center rounded-[10px] border border-white/15 bg-white/[0.02] text-forge-soft";
const eyebrowClass =
  "m-0 font-mono text-[11px] uppercase tracking-[0.08em] text-forge-accent";
const cardTitleClass =
  "my-2 max-w-[300px] text-2xl font-extrabold leading-[1.12] tracking-[-0.04em]";
const cardTextClass = "m-0 max-w-[300px] text-[13px] leading-[1.6] text-forge-soft";

export function WorkspacePage() {
  const { user, tokens, signOut } = useAuth();
  const { data: profile } = useQuery({
    queryKey: ["auth", "me", tokens?.access_token],
    queryFn: () => getCurrentUser(tokens!.access_token),
    enabled: Boolean(tokens?.access_token),
  });
  const currentUser = profile ?? user;
  const displayName = currentUser?.name?.trim() || "Workspace user";
  const firstName = currentUser?.name?.trim() ? displayName.split(/\s+/)[0] : "there";
  const email = currentUser?.email?.trim() || "No email available";
  const initials = displayName
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0])
    .join("")
    .toUpperCase();

  return (
    <main className="min-h-screen bg-forge-bg px-6 pb-12 text-[#f1f3f5] lg:px-[6vw] lg:pb-[72px]">
      <header className="flex flex-wrap items-center justify-between gap-5 border-b border-white/15 py-[18px] lg:h-[82px] lg:py-0">
        <div className="flex items-center gap-3 font-extrabold">
          <span className="grid size-8 place-items-center rounded-lg border border-forge-accent/50 bg-forge-accent/10 tracking-[-0.12em] text-forge-accent shadow-[0_0_20px_rgba(198,243,106,0.12)]">
            F/
          </span>
          <span>ForgeAI</span>
        </div>

        <nav className="order-3 flex w-full items-center justify-between gap-1 rounded-full border border-white/15 bg-forge-panel/80 p-1 shadow-[0_8px_24px_rgba(0,0,0,0.18)] lg:order-none lg:w-auto lg:justify-start" aria-label="Main navigation">
          <button type="button" className="rounded-full bg-forge-accent px-3 py-2 text-xs font-bold text-[#111419] shadow-[0_0_16px_rgba(198,243,106,0.16)]">
            Overview
          </button>
          <button type="button" className="rounded-full px-3 py-2 text-xs font-bold text-forge-soft">
            Agents
          </button>
          <button type="button" className="rounded-full px-3 py-2 text-xs font-bold text-forge-soft">
            Projects
          </button>
        </nav>

        <div className="flex items-center gap-3 text-xs text-forge-soft">
          <div className="grid size-9 place-items-center rounded-full border border-forge-accent/45 bg-forge-accent/10 font-mono text-[11px] font-bold text-forge-accent shadow-[0_0_18px_rgba(198,243,106,0.12)]" aria-hidden="true">
            {initials}
          </div>
          <div className="flex min-w-0 flex-col items-end leading-tight">
            <span className="max-w-[150px] truncate font-bold text-[#edf2ee]">{displayName}</span>
            <span className="max-w-[220px] truncate text-[11px] text-forge-muted max-[560px]:hidden">{email}</span>
          </div>
          <button
            className="grid place-items-center border-0 bg-transparent text-forge-soft transition hover:text-forge-accent"
            title="Sign out"
            aria-label="Sign out"
            onClick={signOut}
          >
            <LogOut size={18} />
          </button>
        </div>
      </header>

      <section className="border-b border-white/15 py-[60px] lg:py-[76px]">
        <p className={eyebrowClass}>Workspace overview</p>
        <h1 className="my-3 max-w-[720px] text-[clamp(42px,5.3vw,72px)] font-extrabold leading-[0.96] tracking-[-0.07em]">
          Good to see you, {firstName}.
        </h1>
        <p className="m-0 max-w-[600px] text-base leading-[1.6] text-forge-soft">
          Your workspace is ready for the next task, code review, or agent run.
        </p>
      </section>

      <section className="grid gap-4 pt-5 lg:grid-cols-[1.35fr_1.1fr_1fr]">
        <article className={`${cardClass} bg-gradient-to-br from-[#202c25] via-[#1a2020] to-[#191d23]`}>
          <div className={`${iconClass} text-forge-accent`}>
            <Sparkles size={18} />
          </div>
          <p className={eyebrowClass}>Agent studio</p>
          <h2 className={cardTitleClass}>Turn rough work into clear execution.</h2>
          <p className={cardTextClass}>
            Start with a goal, keep the context tight, and move quickly with a structured workflow.
          </p>
          <button type="button" className="mt-5 inline-flex items-center gap-2 border-0 bg-transparent p-0 font-bold text-forge-accent">
            Open studio <ArrowRight size={16} />
          </button>
        </article>

        <article className={cardClass}>
          <div className={iconClass}>
            <Workflow size={18} />
          </div>
          <p className={eyebrowClass}>Flow</p>
          <h2 className={cardTitleClass}>Simple, focused, and reliable.</h2>
          <p className={cardTextClass}>
            Keep the work visible and your decisions connected to the actual project context.
          </p>
        </article>

        <article className={cardClass}>
          <div className={iconClass}>
            <BriefcaseBusiness size={18} />
          </div>
          <p className={eyebrowClass}>Projects</p>
          <h2 className={cardTitleClass}>One place for active work.</h2>
          <p className={cardTextClass}>
            Connect a repo, track the next move, and keep the work moving without extra friction.
          </p>
        </article>
      </section>
    </main>
  );
}
