import { useState, type FormEvent, type ReactNode } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { X } from "lucide-react";
import { useToast } from "../../../../../shared/ui/useToast";
import { createProject } from "../../../api/projects.api";
import type { CreateProjectInput } from "../types/project.types";

export function CreateProjectDialog({ accessToken, organizationId, onClose }: { accessToken: string; organizationId: string; onClose: () => void }) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const toast = useToast();
  const queryClient = useQueryClient();
  const createMutation = useMutation({
    mutationFn: () => {
      const input: CreateProjectInput = {
        name: name.trim(),
        description: description.trim() || undefined,
      };
      return createProject(accessToken, organizationId, input);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["projects", organizationId] });
      toast.pushSuccess("Project created");
      onClose();
    },
    onError: (error) => toast.pushError(error instanceof Error ? error.message : "Could not create project"),
  });

  function submit(event: FormEvent) {
    event.preventDefault();
    if (name.trim()) createMutation.mutate();
  }

  const canCreate = Boolean(name.trim());

  return (
    <div className="fixed inset-0 z-30 grid place-items-center bg-[var(--overlay)] p-5 backdrop-blur-md">
      <form className="w-full max-w-[560px] rounded-xl border border-[var(--border)] bg-forge-panel p-7 shadow-[0_24px_80px_rgba(20,35,25,.2)]" onSubmit={submit}>
        <div className="flex items-start justify-between gap-5">
          <div><span className="mb-2 block font-mono text-[10px] uppercase tracking-[.12em] text-forge-accent">New project</span><h2 className="m-0 text-[22px] font-bold tracking-[-.03em]">Create a project</h2></div>
          <button type="button" className="grid size-9 place-items-center rounded-md text-forge-muted hover:bg-[var(--surface-hover)]" onClick={onClose} aria-label="Close dialog"><X size={18} /></button>
        </div>
        <div className="mt-6 grid gap-4">
          <Field label="Project name"><input className="input" value={name} onChange={(event) => setName(event.target.value)} placeholder="e.g. Agent Console" autoFocus /></Field>
          <Field label="Description"><textarea className="input min-h-20 resize-y" value={description} onChange={(event) => setDescription(event.target.value)} placeholder="What is this project for?" /></Field>
        </div>
        <div className="mt-7 flex justify-end gap-2"><button type="button" className="rounded-md bg-[var(--surface-subtle)] px-4 py-2.5 text-xs font-extrabold text-forge-soft hover:bg-[var(--surface-hover)]" onClick={onClose}>Cancel</button><button className="rounded-md bg-forge-accent px-4 py-2.5 text-xs font-extrabold text-[var(--primary-foreground)] hover:bg-forge-accent-strong disabled:cursor-not-allowed disabled:opacity-45" disabled={!canCreate || createMutation.isPending}>{createMutation.isPending ? "Creating..." : "Create project"}</button></div>
      </form>
    </div>
  );
}

function Field({ label, children }: { label: string; children: ReactNode }) { return <label className="block"><span className="label">{label}</span>{children}</label>; }
