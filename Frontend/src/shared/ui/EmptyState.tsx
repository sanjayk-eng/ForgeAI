import type { ReactNode } from "react";
import type { LucideIcon } from "lucide-react";

type EmptyStateProps = {
  icon: LucideIcon;
  title: string;
  description?: string;
  action?: ReactNode;
};

export function EmptyState({ icon: Icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center px-6 py-10">
      <div className="mb-3 grid size-12 place-items-center rounded-xl border border-white/15 bg-white/5">
        <Icon size={22} className="text-gray-400" />
      </div>
      <p className="mb-1 text-sm font-semibold text-white">{title}</p>
      {description && (
        <p className="text-center text-xs leading-5 text-gray-400">
          {description}
        </p>
      )}
      {action && <div className="mt-4">{action}</div>}
    </div>
  );
}
