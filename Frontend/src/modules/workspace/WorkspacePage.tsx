import { ArrowRight, BriefcaseBusiness, LogOut, Sparkles, Workflow } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { useAuth } from "../auth/useAuth";
import { getCurrentUser } from "../auth/api";

export function WorkspacePage() {
  const { user, tokens, signOut } = useAuth();
  const { data: profile } = useQuery({
    queryKey: ["auth", "me", tokens?.access_token],
    queryFn: () => getCurrentUser(tokens!.access_token),
    enabled: Boolean(tokens?.access_token),
  });

  const currentUser = profile ?? user;
  const firstName = currentUser?.name.split(" ")[0] ?? "there";

  return (
    <main className="workspace-shell">
      <header className="workspace-nav">
        <div className="brand-lockup">
          <span className="brand-mark">F/</span>
          <span>ForgeAI</span>
        </div>

        <nav className="workspace-tabs" aria-label="Main navigation">
          <button type="button" className="tab-button active">Overview</button>
          <button type="button" className="tab-button">Agents</button>
          <button type="button" className="tab-button">Projects</button>
        </nav>

        <div className="user-menu">
          <span>{currentUser?.email}</span>
          <button className="icon-button" title="Sign out" aria-label="Sign out" onClick={signOut}>
            <LogOut size={18} />
          </button>
        </div>
      </header>

      <section className="workspace-hero">
        <p className="eyebrow">Workspace overview</p>
        <h1>Good to see you, {firstName}.</h1>
        <p>Your workspace is ready for the next task, code review, or agent run.</p>
      </section>

      <section className="workspace-grid">
        <article className="workspace-card primary-card">
          <div className="card-icon"><Sparkles size={18} /></div>
          <p className="eyebrow">Agent studio</p>
          <h2>Turn rough work into clear execution.</h2>
          <p>Start with a goal, keep the context tight, and move quickly with a structured workflow.</p>
          <button type="button" className="card-action">
            Open studio <ArrowRight size={16} />
          </button>
        </article>

        <article className="workspace-card">
          <div className="card-icon muted"><Workflow size={18} /></div>
          <p className="eyebrow">Flow</p>
          <h2>Simple, focused, and reliable.</h2>
          <p>Keep the work visible and your decisions connected to the actual project context.</p>
        </article>

        <article className="workspace-card">
          <div className="card-icon muted"><BriefcaseBusiness size={18} /></div>
          <p className="eyebrow">Projects</p>
          <h2>One place for active work.</h2>
          <p>Connect a repo, track the next move, and keep the work moving without extra friction.</p>
        </article>
      </section>
    </main>
  );
}
