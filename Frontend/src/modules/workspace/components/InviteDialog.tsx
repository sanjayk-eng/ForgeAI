import { useEffect, useState, type FormEvent } from "react";
import { useMutation } from "@tanstack/react-query";
import { Mail, X } from "lucide-react";
import { useToast } from "../../../shared/ui/useToast";
import { createInvite } from "../api/invites.api";

export function InviteDialog({
  accessToken,
  workspaceId,
  onClose,
}: {
  accessToken: string;
  workspaceId: string;
  onClose: () => void;
}) {
  const toast = useToast();
  const [email, setEmail] = useState("");
  const [role, setRole] = useState("MEMBER");
  const validEmail = email.includes("@");
  const mutation = useMutation({
    mutationFn: () =>
      createInvite(accessToken, workspaceId, email.trim(), role),
    onSuccess: onClose,
    onError: (error) =>
      toast.pushError(
        error instanceof Error ? error.message : "Could not send invitation",
      ),
  });
  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !mutation.isPending) onClose();
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [mutation.isPending, onClose]);
  function submit(event: FormEvent) {
    event.preventDefault();
    if (validEmail) mutation.mutate();
  }
  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-[var(--overlay)] p-5 backdrop-blur-md" role="dialog" aria-modal="true" aria-labelledby="invite-title" onClick={() => !mutation.isPending && onClose()}>
      <form
        className="w-full max-w-[450px] rounded-xl border border-forge-accent/20 bg-forge-card p-6 shadow-2xl sm:p-7"
        onSubmit={submit}
        onClick={(event) => event.stopPropagation()}
      >
        <div className="flex justify-between">
          <div>
            <span className="mb-2 block font-mono text-[10px] uppercase tracking-[.12em] text-forge-accent">
              Team access
            </span>
            <h2 id="invite-title" className="m-0 text-[22px] font-bold text-forge-text">Invite a member</h2>
          </div>
          <button
            type="button"
            className="grid size-9 place-items-center rounded-lg text-forge-muted transition hover:bg-[var(--surface-hover)] hover:text-forge-text focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent"
            onClick={onClose}
            aria-label="Close dialog"
          >
            <X size={18} />
          </button>
        </div>
        <p className="mt-3 text-sm text-forge-muted">
          They will receive an invitation to join this workspace.
        </p>
        <label
          className="mt-6 block font-mono text-[11px] text-forge-soft"
          htmlFor="invite-email"
        >
          Email address
        </label>
        <div className="relative mt-2">
          <Mail className="absolute left-3 top-3 text-forge-muted" size={16} />
          <input
            id="invite-email"
            className="w-full rounded-lg border border-[var(--border)] bg-[var(--input)] py-3 pl-10 pr-3 text-sm text-forge-text outline-none transition focus:border-forge-accent focus:ring-2 focus:ring-forge-accent/15"
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            placeholder="name@company.com"
            autoFocus
          />
        </div>
        {email.length > 0 && !validEmail && (
          <p className="mt-2 text-xs text-[var(--destructive)]" role="alert">Enter a valid email address.</p>
        )}
        <label
          className="mt-5 block font-mono text-[11px] text-forge-soft"
          htmlFor="invite-role"
        >
          Workspace role
        </label>
        <select
          id="invite-role"
          className="mt-2 w-full rounded-lg border border-[var(--border)] bg-[var(--input)] p-3 text-sm text-forge-text outline-none focus:border-forge-accent focus:ring-2 focus:ring-forge-accent/15"
          value={role}
          onChange={(event) => setRole(event.target.value)}
        >
          <option value="MEMBER">Member · standard access</option>
          <option value="ADMIN">Admin · manage the workspace</option>
        </select>
        <div className="mt-7 flex justify-end gap-2">
          <button
            type="button"
            className="rounded-lg border border-[var(--border)] bg-[var(--surface-subtle)] px-4 py-2.5 text-xs font-bold text-forge-soft transition hover:bg-[var(--surface-hover)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent"
            onClick={onClose}
          >
            Cancel
          </button>
          <button
            className="rounded-md bg-forge-accent px-4 py-2.5 text-xs font-bold text-forge-bg disabled:opacity-45"
            disabled={mutation.isPending || !validEmail}
          >
            {mutation.isPending ? "Sending..." : "Send invitation"}
          </button>
        </div>
      </form>
    </div>
  );
}
