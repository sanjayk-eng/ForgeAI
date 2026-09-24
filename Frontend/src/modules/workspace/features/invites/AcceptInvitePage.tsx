import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CheckCircle, XCircle, Clock, AlertCircle } from "lucide-react";
import { useAuth } from "../../../auth/useAuth";
import { useToast } from "../../../../shared/ui/useToast";
import { acceptInvite, getInviteByToken, rejectInvite } from "../../api/invites.api";

export function AcceptInvitePage() {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const { tokens, user, loading: authLoading } = useAuth();
  const toast = useToast();
  const queryClient = useQueryClient();
  const [rejecting, setRejecting] = useState(false);

  // Fetch invite details
  const inviteQuery = useQuery({
    queryKey: ["invite", token],
    queryFn: () => getInviteByToken(token!),
    enabled: Boolean(token),
    retry: false,
  });

  const invite = inviteQuery.data;
  const isExpired = invite ? new Date(invite.expires_at) < new Date() : false;
  const isAuthenticated = Boolean(tokens?.access_token);
  const userEmail = user?.email?.toLowerCase();
  const inviteEmail = invite?.email?.toLowerCase();
  const emailMatches = userEmail === inviteEmail;

  // Accept mutation
  const acceptMutation = useMutation({
    mutationFn: () => {
      if (!tokens?.access_token || !token || !user?.email) {
        throw new Error("Missing required data");
      }
      return acceptInvite(tokens.access_token, token, user.email);
    },
    onSuccess: (data) => {
      toast.pushSuccess("Invitation accepted! Redirecting to workspace...");
      void queryClient.invalidateQueries({ queryKey: ["my-pending-invites"] });
      navigate(`/workspace?workspace=${data.workspace_id}`, { replace: true });
    },
    onError: (error: Error) => {
      toast.pushError(error.message || "Failed to accept invitation");
    },
  });

  // Reject mutation
  const rejectMutation = useMutation({
    mutationFn: () => {
      if (!token || !invite?.email) {
        throw new Error("Missing required data");
      }
      return rejectInvite(token, invite.email);
    },
    onSuccess: () => {
      toast.pushSuccess("Invitation rejected");
      void queryClient.invalidateQueries({ queryKey: ["my-pending-invites"] });
      navigate("/", { replace: true });
    },
    onError: (error: Error) => {
      toast.pushError(error.message || "Failed to reject invitation");
    },
  });

  // Auto-redirect to login if not authenticated
  useEffect(() => {
    if (invite && !isAuthenticated && !rejecting) {
      const redirectUrl = `/accept-invite/${token}`;
      localStorage.setItem("redirect_after_login", redirectUrl);
    }
  }, [invite, isAuthenticated, token, rejecting]);

  if (authLoading || inviteQuery.isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-forge-bg">
        <div className="text-center">
          <Clock className="mx-auto mb-4 h-12 w-12 animate-pulse text-forge-accent" />
          <p className="text-forge-muted">Loading invitation...</p>
        </div>
      </div>
    );
  }

  if (inviteQuery.isError || !invite) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-forge-bg px-4">
        <div className="w-full max-w-md rounded-lg border border-red-500/20 bg-forge-panel p-8 text-center">
          <XCircle className="mx-auto mb-4 h-16 w-16 text-red-500" />
          <h1 className="mb-2 text-2xl font-bold text-forge-text">
            Invalid Invitation
          </h1>
          <p className="mb-6 text-forge-muted">
            This invitation link is invalid or has been removed.
          </p>
          <button
            onClick={() => navigate("/")}
            className="rounded-md bg-forge-accent px-6 py-2.5 font-bold text-[var(--primary-foreground)] transition hover:bg-forge-accent-strong"
          >
            Go to Home
          </button>
        </div>
      </div>
    );
  }

  // Already used
  if (invite.status !== "PENDING") {
    return (
      <div className="flex min-h-screen items-center justify-center bg-forge-bg px-4">
        <div className="w-full max-w-md rounded-lg border border-white/10 bg-forge-panel p-8 text-center">
          <AlertCircle className="mx-auto mb-4 h-16 w-16 text-yellow-500" />
          <h1 className="mb-2 text-2xl font-bold text-forge-text">
            Invitation {invite.status.toLowerCase()}
          </h1>
          <p className="mb-6 text-forge-muted">
            This invitation has already been {invite.status.toLowerCase()}.
          </p>
          <button
            onClick={() => navigate("/")}
            className="rounded-md bg-forge-accent px-6 py-2.5 font-bold text-[var(--primary-foreground)] transition hover:bg-forge-accent-strong"
          >
            Go to Home
          </button>
        </div>
      </div>
    );
  }

  // Expired
  if (isExpired) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-forge-bg px-4">
        <div className="w-full max-w-md rounded-lg border border-white/10 bg-forge-panel p-8 text-center">
          <Clock className="mx-auto mb-4 h-16 w-16 text-forge-muted" />
          <h1 className="mb-2 text-2xl font-bold text-forge-text">
            Invitation Expired
          </h1>
          <p className="mb-6 text-forge-muted">
            This invitation expired on{" "}
            {new Date(invite.expires_at).toLocaleDateString()}.
          </p>
          <button
            onClick={() => navigate("/")}
            className="rounded-md bg-forge-accent px-6 py-2.5 font-bold text-[var(--primary-foreground)] transition hover:bg-forge-accent-strong"
          >
            Go to Home
          </button>
        </div>
      </div>
    );
  }

  // Not authenticated
  if (!isAuthenticated) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-forge-bg px-4">
        <div className="w-full max-w-md rounded-lg border border-white/10 bg-forge-panel p-8">
          <h1 className="mb-2 text-2xl font-bold text-forge-text">
            Workspace Invitation
          </h1>
          <p className="mb-6 text-forge-muted">
            <span className="font-semibold text-forge-text">
              {invite.invited_by_user?.name || "Someone"}
            </span>{" "}
            invited you to join their workspace as a{" "}
            <span className="font-semibold text-forge-accent">
              {invite.role}
            </span>
            .
          </p>

          <div className="mb-6 rounded border border-white/10 bg-forge-bg p-4">
            <div className="text-sm text-forge-muted">Invited email</div>
            <div className="mt-1 font-mono text-sm text-forge-text">
              {invite.email}
            </div>
          </div>

          <div className="space-y-3">
            <button
              onClick={() => navigate("/login")}
              className="w-full rounded-md bg-forge-accent px-6 py-3 font-bold text-[var(--primary-foreground)] transition hover:bg-forge-accent-strong"
            >
              Sign in to accept
            </button>
            <button
              onClick={() => navigate("/register")}
              className="w-full rounded-md border border-white/20 bg-transparent px-6 py-3 font-bold text-forge-text transition hover:bg-white/5"
            >
              Create account
            </button>
            <button
              onClick={() => {
                setRejecting(true);
                rejectMutation.mutate();
              }}
              disabled={rejectMutation.isPending}
              className="w-full rounded-md bg-transparent px-6 py-2 text-sm font-medium text-forge-muted transition hover:text-forge-text disabled:opacity-50"
            >
              {rejectMutation.isPending ? "Rejecting..." : "Decline invitation"}
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Authenticated but email mismatch
  if (!emailMatches) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-forge-bg px-4">
        <div className="w-full max-w-md rounded-lg border border-red-500/20 bg-forge-panel p-8">
          <AlertCircle className="mx-auto mb-4 h-12 w-12 text-red-500" />
          <h1 className="mb-2 text-center text-2xl font-bold text-forge-text">
            Email Mismatch
          </h1>
          <p className="mb-4 text-center text-forge-muted">
            This invitation was sent to:
          </p>
          <div className="mb-4 rounded border border-white/10 bg-forge-bg p-3 text-center">
            <code className="text-sm text-forge-accent">{inviteEmail}</code>
          </div>
          <p className="mb-6 text-center text-forge-muted">
            But you're signed in as:
          </p>
          <div className="mb-6 rounded border border-white/10 bg-forge-bg p-3 text-center">
            <code className="text-sm text-forge-text">{userEmail}</code>
          </div>
          <p className="mb-6 text-center text-sm text-forge-muted">
            Please sign out and sign in with the invited email address.
          </p>
          <button
            onClick={() => navigate("/")}
            className="w-full rounded-md bg-forge-accent px-6 py-2.5 font-bold text-[var(--primary-foreground)] transition hover:bg-forge-accent-strong"
          >
            Go to Home
          </button>
        </div>
      </div>
    );
  }

  // Ready to accept
  return (
    <div className="flex min-h-screen items-center justify-center bg-forge-bg px-4">
      <div className="w-full max-w-md rounded-lg border border-white/10 bg-forge-panel p-8">
        <CheckCircle className="mx-auto mb-4 h-16 w-16 text-forge-accent" />
        <h1 className="mb-2 text-center text-2xl font-bold text-forge-text">
          Join Workspace
        </h1>
        <p className="mb-6 text-center text-forge-muted">
          <span className="font-semibold text-forge-text">
            {invite.invited_by_user?.name || "Someone"}
          </span>{" "}
          invited you to join as a{" "}
          <span className="font-semibold text-forge-accent">{invite.role}</span>
          .
        </p>

        <div className="mb-6 rounded border border-white/10 bg-forge-bg p-4">
          <div className="text-sm text-forge-muted">Your email</div>
          <div className="mt-1 font-mono text-sm text-forge-text">
            {userEmail}
          </div>
        </div>

        <div className="space-y-3">
          <button
            onClick={() => acceptMutation.mutate()}
            disabled={acceptMutation.isPending}
            className="w-full rounded-md bg-forge-accent px-6 py-3 font-bold text-[var(--primary-foreground)] transition hover:bg-forge-accent-strong disabled:opacity-50"
          >
            {acceptMutation.isPending ? "Accepting..." : "Accept invitation"}
          </button>
          <button
            onClick={() => rejectMutation.mutate()}
            disabled={rejectMutation.isPending}
            className="w-full rounded-md border border-white/20 bg-transparent px-6 py-2 text-sm font-medium text-forge-muted transition hover:text-forge-text disabled:opacity-50"
          >
            {rejectMutation.isPending ? "Declining..." : "Decline invitation"}
          </button>
        </div>
      </div>
    </div>
  );
}
