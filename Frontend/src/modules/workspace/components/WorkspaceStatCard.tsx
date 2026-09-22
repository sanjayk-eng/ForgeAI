import type { LucideIcon } from "lucide-react";

type Props = {
  icon: LucideIcon;
  label: string;
  value: string;
  detail: string;
  accent?: boolean;
};

export function WorkspaceStatCard({
  icon: Icon,
  label,
  value,
  detail,
  accent = false,
}: Props) {
  return (
    <article
      className={`group border p-5 transition duration-300 hover:-translate-y-0.5 hover:border-forge-accent/30 ${accent ? "border-forge-accent/25 bg-forge-accent/[0.07]" : "border-white/[0.08] bg-forge-panel/70"}`}
    >
      <div className="mb-7 flex items-start justify-between">
        <span className="grid size-9 place-items-center border border-forge-accent/25 text-forge-accent">
          <Icon size={17} strokeWidth={1.8} />
        </span>
        <span className="font-mono text-[10px] uppercase tracking-[.12em] text-forge-muted">
          Live
        </span>
      </div>
      <p className="m-0 text-xs text-forge-soft">{label}</p>
      <strong className="mt-2 block truncate text-xl font-extrabold tracking-[-.03em] text-forge-text">
        {value}
      </strong>
      <p className="mt-2 truncate font-mono text-[10px] text-forge-muted">
        {detail}
      </p>
    </article>
  );
}
