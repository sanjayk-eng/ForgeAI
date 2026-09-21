import { ArrowRight, Box, LogOut, Radio, Sparkles } from "lucide-react";
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

  return <main className="workspace-shell">
    <header className="workspace-nav"><div className="brand-lockup"><span className="brand-mark">F/</span><span>ForgeAI</span></div><div className="user-menu"><span>{currentUser?.email}</span><button className="icon-button" title="Sign out" aria-label="Sign out" onClick={signOut}><LogOut size={18} /></button></div></header>
    <section className="workspace-hero"><p className="eyebrow">Workspace / overview</p><h1>Good to see you, {firstName}.</h1><p>Your ForgeAI control room is ready for the next useful thing.</p></section>
    <section className="workspace-grid">
      <article className="workspace-card primary-card"><div className="card-icon"><Sparkles size={19} /></div><p className="eyebrow">Agent studio</p><h2>Turn a rough thought into a working plan.</h2><p>Start a conversation with your coding agent and keep every decision close to the work.</p><button className="card-action">Open studio <ArrowRight size={16} /></button></article>
      <article className="workspace-card"><div className="card-icon muted"><Radio size={19} /></div><p className="eyebrow">Activity</p><h2>Quiet for now.</h2><p>Your latest runs, changes, and shared context will appear here.</p></article>
      <article className="workspace-card"><div className="card-icon muted"><Box size={19} /></div><p className="eyebrow">Projects</p><h2>One clear place to begin.</h2><p>Connect a repository when you are ready to give your agent something to work on.</p></article>
    </section>
  </main>;
}
