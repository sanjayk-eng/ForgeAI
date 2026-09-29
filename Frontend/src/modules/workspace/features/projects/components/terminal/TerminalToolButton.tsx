import type { ReactNode } from "react";

export function TerminalToolButton({
  label,
  disabled,
  onClick,
  children,
}: {
  label: string;
  disabled: boolean;
  onClick: () => void;
  children: ReactNode;
}) {
  return (
    <button
      type="button"
      title={label}
      aria-label={label}
      disabled={disabled}
      onClick={onClick}
      className="grid size-8 place-items-center rounded text-forge-muted hover:bg-[var(--surface-hover)] hover:text-forge-text disabled:opacity-30"
    >
      {children}
    </button>
  );
}
