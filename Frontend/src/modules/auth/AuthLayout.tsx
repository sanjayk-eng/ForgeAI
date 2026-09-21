import { ArrowUpRight, ShieldCheck } from "lucide-react";
import { Link, Outlet } from "react-router-dom";

export function AuthLayout() {
  return (
    <main className="auth-shell">
      <section className="auth-poster" aria-label="ForgeAI workspace">
        <div className="poster-topline">
          <span className="brand-mark">F/</span>
          <span>ForgeAI access</span>
        </div>
        <div className="poster-copy">
          <p className="eyebrow">Agent workspace / 01</p>
          <h1>Build with a clear line of sight.</h1>
          <p className="poster-note">One secure home for your agents, code, and conversations.</p>
        </div>
        <div className="poster-foot">
          <span><ShieldCheck size={16} /> Protected by your workspace identity</span>
          <span>2026</span>
        </div>
      </section>
      <section className="auth-panel">
        <div className="mobile-brand"><span className="brand-mark">F/</span><span>ForgeAI</span></div>
        <Outlet />
        <p className="auth-legal">By continuing, you agree to use ForgeAI responsibly and keep your credentials private.</p>
      </section>
    </main>
  );
}

export function AuthSwitch({ prompt, label, to }: { prompt: string; label: string; to: string }) {
  return <p className="auth-switch">{prompt} <Link to={to}>{label} <ArrowUpRight size={14} /></Link></p>;
}
