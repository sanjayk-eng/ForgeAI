import { Bell, Clock, ArrowRight } from "lucide-react";
import type { WorkspaceInvite } from "../types/workspace.types";

type InviteNotificationCardProps = {
  invite: WorkspaceInvite;
  onClick: () => void;
};

export function InviteNotificationCard({ invite, onClick }: InviteNotificationCardProps) {
  const isExpired = new Date(invite.expires_at) < new Date();
  const daysLeft = Math.ceil(
    (new Date(invite.expires_at).getTime() - new Date().getTime()) / (1000 * 60 * 60 * 24)
  );

  return (
    <button
      onClick={onClick}
      className="group flex w-full items-start gap-3 px-4 py-3 text-left transition hover:bg-white/5"
    >
      {/* Icon */}
      <div className="mt-0.5 grid size-9 shrink-0 place-items-center rounded-lg border border-white/15 bg-white/5 text-gray-400 transition group-hover:border-forge-accent/40 group-hover:bg-forge-accent/10 group-hover:text-forge-accent">
        <Bell size={16} />
      </div>
      
      {/* Content */}
      <div className="min-w-0 flex-1">
        <p className="text-sm font-semibold text-white">
          {invite.workspace_name || "Workspace"}
        </p>
        <p className="mt-1 text-xs text-gray-400">
          Invited by <span className="text-gray-300">{invite.invited_by_user?.name || "Someone"}</span> as <span className="font-medium text-white">{invite.role}</span>
        </p>
        
        {/* Footer */}
        <div className="mt-2 flex items-center justify-between text-xs">
          <div className="flex items-center gap-1.5 text-gray-400">
            <Clock size={11} />
            {isExpired ? (
              <span className="text-red-400">Expired</span>
            ) : daysLeft <= 1 ? (
              <span className="text-yellow-400">Expires today</span>
            ) : (
              <span>Expires in {daysLeft}d</span>
            )}
          </div>
          <div className="flex items-center gap-1 font-medium text-forge-accent opacity-0 transition group-hover:opacity-100">
            View
            <ArrowRight size={12} />
          </div>
        </div>
      </div>
    </button>
  );
}
