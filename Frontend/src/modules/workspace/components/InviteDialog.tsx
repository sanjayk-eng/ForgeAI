import { useState, type FormEvent } from "react";
import { useMutation } from "@tanstack/react-query";
import { Mail, X } from "lucide-react";
import { useToast } from "../../../shared/ui/useToast";
import { createInvite } from "../api";

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
  const mutation = useMutation({
    mutationFn: () =>
      createInvite(accessToken, workspaceId, email.trim(), role),
    onSuccess: onClose,
    onError: (error) =>
      toast.pushError(
        error instanceof Error ? error.message : "Could not send invitation",
      ),
  });
  function submit(event: FormEvent) {
    event.preventDefault();
    if (email.includes("@")) mutation.mutate();
  }
  return (
    <div className="fixed inset-0 z-30 grid place-items-center bg-black/75 p-5 backdrop-blur-md">
      <form
        className="w-full max-w-[450px] border border-forge-accent/20 bg-[#171b20] p-7"
        onSubmit={submit}
      >
        <div className="flex justify-between">
          <div>
            <span className="mb-2 block font-mono text-[10px] uppercase tracking-[.12em] text-forge-accent">
              Team access
            </span>
            <h2 className="m-0 text-[22px] font-bold">Invite a member</h2>
          </div>
          <button
            type="button"
            className="text-forge-muted"
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
            className="w-full rounded-md border border-white/[0.13] bg-[#101217] py-3 pl-10 pr-3 text-sm text-forge-text outline-none focus:border-forge-accent"
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            placeholder="name@company.com"
            autoFocus
          />
        </div>
        <label
          className="mt-5 block font-mono text-[11px] text-forge-soft"
          htmlFor="invite-role"
        >
          Workspace role
        </label>
        <select
          id="invite-role"
          className="mt-2 w-full rounded-md border border-white/[0.13] bg-[#101217] p-3 text-sm text-forge-text"
          value={role}
          onChange={(event) => setRole(event.target.value)}
        >
          <option value="MEMBER">Member · standard access</option>
          <option value="ADMIN">Admin · manage the workspace</option>
        </select>
        <div className="mt-7 flex justify-end gap-2">
          <button
            type="button"
            className="rounded-md bg-white/[0.06] px-4 py-2.5 text-xs font-bold text-forge-soft"
            onClick={onClose}
          >
            Cancel
          </button>
          <button
            className="rounded-md bg-forge-accent px-4 py-2.5 text-xs font-bold text-forge-bg disabled:opacity-45"
            disabled={mutation.isPending || !email.includes("@")}
          >
            {mutation.isPending ? "Sending..." : "Send invitation"}
          </button>
        </div>
      </form>
    </div>
  );
}
