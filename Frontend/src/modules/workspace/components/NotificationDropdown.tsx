import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Bell, CheckCircle2 } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { getMyPendingInvites } from "../api/invites.api";
import { Dropdown } from "../../../shared/ui/Dropdown";
import { EmptyState } from "../../../shared/ui/EmptyState";
import { NotificationBadge } from "./NotificationBadge";
import { InviteNotificationCard } from "./InviteNotificationCard";

type NotificationDropdownProps = {
  accessToken: string;
};

export function NotificationDropdown({ accessToken }: NotificationDropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const navigate = useNavigate();

  const { data: invites = [] } = useQuery({
    queryKey: ["my-pending-invites"],
    queryFn: () => getMyPendingInvites(accessToken),
    enabled: Boolean(accessToken),
    refetchInterval: 30000,
  });

  const pendingCount = invites.length;

  const handleInviteClick = (token: string) => {
    setIsOpen(false);
    navigate(`/accept-invite/${token}`);
  };

  return (
    <div className="relative">
      {/* Bell Button */}
      <button
        className="relative grid size-9 place-items-center rounded-md border border-transparent text-forge-muted transition hover:border-white/10 hover:bg-white/[0.06] hover:text-forge-text focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent"
        aria-label={`Notifications${pendingCount > 0 ? ` (${pendingCount} pending)` : ""}`}
        onClick={() => setIsOpen(!isOpen)}
      >
        <Bell size={17} />
        <NotificationBadge count={pendingCount} />
      </button>

      {/* Dropdown */}
      <Dropdown 
        isOpen={isOpen} 
        onClose={() => setIsOpen(false)}
        className="w-[min(380px,calc(100vw-32px))]"
        align="right"
      >
        {/* Header */}
        <div className="border-b border-white/10 bg-[#1f2329] px-4 py-3">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-sm font-semibold text-white">
                Invitations
              </h3>
              <p className="mt-0.5 text-xs text-gray-400">
                {pendingCount === 0 ? "No pending invites" : `${pendingCount} pending`}
              </p>
            </div>
            {pendingCount > 0 && (
              <div className="rounded-full bg-forge-accent/15 px-2 py-0.5 font-mono text-[10px] font-semibold text-forge-accent">
                {pendingCount}
              </div>
            )}
          </div>
        </div>

        {/* Content */}
        <div className="max-h-[420px] overflow-y-auto">
          {pendingCount === 0 ? (
            <EmptyState
              icon={CheckCircle2}
              title="All caught up"
              description="No pending workspace invitations."
            />
          ) : (
            <div className="divide-y divide-white/[0.06]">
              {invites.map((invite) => (
                <InviteNotificationCard
                  key={invite.id}
                  invite={invite}
                  onClick={() => handleInviteClick(invite.token!)}
                />
              ))}
            </div>
          )}
        </div>

        {/* Footer */}
        {pendingCount > 0 && (
          <div className="border-t border-white/10 bg-[#1f2329] px-4 py-2">
            <p className="text-center text-xs text-gray-400">
              Click to view details
            </p>
          </div>
        )}
      </Dropdown>
    </div>
  );
}
