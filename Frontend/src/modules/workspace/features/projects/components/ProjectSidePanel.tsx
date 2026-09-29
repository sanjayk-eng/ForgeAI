import { X } from "lucide-react";
import type { ReactNode } from "react";

export function ProjectSidePanel({
  title,
  icon,
  onClose,
  children,
}: {
  title: string;
  icon: ReactNode;
  onClose: () => void;
  children: ReactNode;
}) {
  return (
    <aside className="flex w-[340px] shrink-0 flex-col border-l border-[var(--border)] bg-forge-panel shadow-[-12px_0_32px_rgba(0,0,0,.08)] max-lg:absolute max-lg:inset-y-0 max-lg:right-0 max-lg:z-20 max-lg:w-[min(360px,calc(100vw-40px))]">
      <div className="flex h-12 shrink-0 items-center justify-between border-b border-[var(--border)] px-4">
        <span className="flex items-center gap-2 text-xs font-semibold text-forge-text">
          {icon}
          {title}
        </span>
        <button
          type="button"
          onClick={onClose}
          aria-label={`Close ${title.toLowerCase()} panel`}
          className="grid size-7 place-items-center rounded text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text"
        >
          <X size={15} />
        </button>
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto p-4">{children}</div>
    </aside>
  );
}