import { ArrowUpRight, ShieldCheck } from "lucide-react";
import { Link, Outlet } from "react-router-dom";

export function AuthLayout() {
  return (
    <main className="grid min-h-screen bg-forge-bg lg:grid-cols-[minmax(300px,0.92fr)_minmax(460px,1.08fr)]">
      <aside className="relative hidden min-h-screen flex-col justify-between overflow-hidden border-r border-white/15 bg-gradient-to-br from-[#191d23] via-[#111419] to-[#0e1014] p-9 text-[#f1f3f5] lg:flex lg:px-12" aria-label="ForgeAI workspace">
        <div className="relative z-10 flex items-center justify-between font-mono text-[11px] uppercase tracking-[0.08em] text-forge-soft">
          <span className="grid size-8 place-items-center rounded-lg border border-forge-accent/50 bg-forge-accent/5 font-sans font-extrabold tracking-[-0.12em] text-forge-accent">F/</span>
          <span>ForgeAI</span>
        </div>
        <div className="relative z-10 max-w-[440px]">
          <p className="m-0 font-mono text-[11px] uppercase tracking-[0.08em] text-forge-accent">ForgeAI / command studio</p>
          <h1 className="my-[18px] max-w-[500px] text-[clamp(42px,4.6vw,64px)] font-extrabold leading-[0.97] tracking-[-0.06em]">Make complex work feel simple.</h1>
          <p className="m-0 max-w-[360px] text-[15px] leading-[1.6] text-forge-soft">A clear workspace for agents, code, and execution.</p>
        </div>
        <div className="relative z-10 flex items-center justify-between font-mono text-[11px] uppercase tracking-[0.08em] text-forge-soft">
          <span className="flex items-center gap-2"><ShieldCheck size={16} /> Encrypted workspace</span>
          <span className="text-forge-signal">09 / 26</span>
        </div>
      </aside>
      <section className="flex min-h-screen w-full flex-col justify-center px-5 py-7 sm:px-8 lg:max-w-[680px] lg:px-[clamp(28px,7vw,120px)]">
        <div className="mb-10 flex items-center gap-2 font-extrabold lg:hidden">
          <span className="grid size-[30px] place-items-center rounded-lg border border-forge-accent/50 bg-forge-accent/5 tracking-[-0.12em] text-forge-accent">F/</span>
          <span>ForgeAI</span>
        </div>
        <Outlet />
        <p className="mx-auto mt-6 w-full max-w-[390px] text-center text-[10px] leading-[1.5] text-forge-muted">By continuing, you agree to use ForgeAI responsibly and keep your credentials private.</p>
      </section>
    </main>
  );
}

export function AuthSwitch({ prompt, label, to }: { prompt: string; label: string; to: string }) {
  return (
    <p className="mt-6 flex flex-wrap items-center justify-center gap-1.5 text-[13px] text-forge-soft">
      {prompt}
      <Link className="inline-flex items-center gap-0.5 font-bold text-forge-accent no-underline" to={to}>
        {label}
        <ArrowUpRight size={14} />
      </Link>
    </p>
  );
}
