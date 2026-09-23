import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Bell, Clock, X } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { getMyPendingInvites } from "../api/invites.api";

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
    refetchInterval: 30000, // Refetch every 30 seconds
  });

  const pendingCount = invites.length;

  return (
    <div className="relative">
      <button
        className="relative grid size-9 place-items-center rounded-md border border-transparent text-forge-muted transition hover:border-white/10 hover:bg-white/[0.06] hover:text-forge-text focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent"
        aria-label="Notifications"
        onClick={() => setIsOpen(!isOpen)}
      >
        <Bell size={17} />
        {pendingCount > 0 && (
          <span className="absolute right-0.5 top-0.5 flex h-4 min-w-[16px] items-center justify-center rounded-full bg-forge-signal px-1 font-mono text-[10px] font-bold text-forge-bg">
            {pendingCount}
          </span>
        )}
      </button>

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-40"
            onClick={() => setIsOpen(false)}
          />
          <div className="absolute right-0 top-full z-50 mt-2 w-[min(380px,calc(100vw-32px))] rounded-lg border border-white/10 bg-forge-panel shadow-[0_12px_40px_rgba(0,0,0,0.5)]">
            <div className="flex items-center justify-between border-b border-white/10 p-4">
              <h3 className="font-mono text-xs font-bold uppercase tracking-wider text-forge-text">
                Workspace Invites
              </h3>
              <button
                onClick={() => setIsOpen(false)}
                className="grid size-7 place-items-center rounded text-forge-muted transition hover:bg-white/10 hover:text-forge-text"
                aria-label="Close"
              >
                <X size={16} />
              </button>
            </div>

            <div className="max-h-[400px] overflow-y-auto">
              {pendingCount === 0 ? (
                <div className="p-8 text-center">
                  <Bell size={32} className="mx-auto mb-3 text-forge-muted opacity-50" />
                  <p className="text-sm text-forge-muted">
                    No pending invites
                  </p>
                </div>
              ) : (
                <div className="divide-y divide-white/5">
                  {invites.map((invite) => {
                    const isExpired = new Date(invite.expires_at) < new Date();
                    return (
                      <button
                        key={invite.id}
                        onClick={() => {
                          setIsOpen(false);
                          navigate(`/accept-invite/${invite.token}`);
                        }}
                        className="flex w-full items-start gap-3 p-4 text-left transition hover:bg-white/[0.03]"
                      >
                        <div className="mt-1 grid size-9 shrink-0 place-items-center rounded-lg border border-forge-accent/30 bg-forge-accent/10 text-forge-accent">
                          <Bell size={16} />
                        </div>
                        <div className="min-w-0 flex-1">
                          <p className="text-sm font-bold text-forge-text">
                            {invite.workspace_name || "Workspace"}
                          </p>
                          <p className="mt-1 text-xs text-forge-muted">
                            <span className="text-forge-soft">
                              {invite.invited_by_user?.name || "Someone"}
                            </span>{" "}
                            invited you as{" "}
                            <span className="font-semibold text-forge-accent">
                              {invite.role}
                            </span>
                          </p>
                          <p className="mt-2 flex items-center gap-1.5 text-xs text-forge-muted">
                            <Clock size={12} />
                            {isExpired ? (
                              <span className="text-red-400">Expired</span>
                            ) : (
                              `Expires ${new Date(invite.expires_at).toLocaleDateString()}`
                            )}
                          </p>
                        </div>
                      </button>
                    );
                  })}
                </div>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
