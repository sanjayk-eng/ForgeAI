import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AlertCircle, Bot, Check, Loader2, Send } from "lucide-react";
import { useState, type FormEvent } from "react";
import { getAgentStatus, runAgentTask } from "../../../api/sandbox.api";

type AgentMessage = {
  role: "user" | "assistant";
  content: string;
  changedFiles?: string[];
};

export function AgentPanel({
  projectName,
  selectedFile,
  accessToken,
  sandboxId,
  sandboxStatus,
}: {
  projectName: string;
  selectedFile: string | null;
  accessToken: string | null;
  sandboxId: string | null;
  sandboxStatus: string | null;
}) {
  const [prompt, setPrompt] = useState("");
  const [messages, setMessages] = useState<AgentMessage[]>([]);
  const [taskError, setTaskError] = useState<string | null>(null);
  const queryClient = useQueryClient();
  const statusQuery = useQuery({
    queryKey: ["agent-status", sandboxId],
    queryFn: () => getAgentStatus(accessToken ?? "", sandboxId ?? ""),
    enabled: Boolean(accessToken && sandboxId && sandboxStatus === "RUNNING"),
    retry: false,
    staleTime: 30_000,
  });
  const taskMutation = useMutation({
    mutationFn: (task: string) => runAgentTask(accessToken ?? "", sandboxId ?? "", task),
    onSuccess: (result) => {
      setMessages((current) => [...current, {
        role: "assistant",
        content: result.message,
        changedFiles: result.changed_files,
      }]);
      setPrompt("");
      setTaskError(null);
      void queryClient.invalidateQueries({ queryKey: ["sandbox-files", sandboxId] });
      void queryClient.invalidateQueries({ queryKey: ["sandbox-file-content", sandboxId] });
    },
    onError: (error) => {
      setTaskError(error instanceof Error ? error.message : "Agent task failed");
    },
  });

  const configured = statusQuery.data?.configured === true;
  const canSend = configured && sandboxStatus === "RUNNING" && prompt.trim().length > 0 && !taskMutation.isPending;

  function submitTask(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const task = prompt.trim();
    if (!canSend || !task) return;
    setMessages((current) => [...current, { role: "user", content: task }]);
    setTaskError(null);
    taskMutation.mutate(task);
  }

  return (
    <div className="flex min-h-full flex-col">
      <div className="flex-1 space-y-4">
        <div>
          <div className="flex items-center gap-2 text-xs font-semibold text-forge-text">
            <Bot size={14} className="text-forge-accent" />
            Project agent
          </div>
          <p className="mt-2 text-xs leading-5 text-forge-muted">
            Ask for code changes or explanations in {projectName}.
          </p>
        </div>

        {sandboxStatus !== "RUNNING" && (
          <StatusNotice>Agent tasks need a running project sandbox.</StatusNotice>
        )}
        {sandboxStatus === "RUNNING" && statusQuery.isLoading && (
          <div className="flex items-center gap-2 text-xs text-forge-muted"><Loader2 size={13} className="animate-spin" /> Checking model configuration</div>
        )}
        {sandboxStatus === "RUNNING" && statusQuery.isError && (
          <StatusNotice>Could not reach the agent service. Check that the backend is running.</StatusNotice>
        )}
        {sandboxStatus === "RUNNING" && statusQuery.data && !configured && (
          <StatusNotice>Set AI_API_KEY on the backend to enable this agent.</StatusNotice>
        )}
        {configured && statusQuery.data?.model && (
          <div className="border border-[var(--border)] px-2.5 py-2 text-[11px] text-forge-muted">Model · {statusQuery.data.model}</div>
        )}
        <div className="border border-forge-signal/20 bg-forge-signal/[0.05] px-3 py-2 text-[10px] leading-4 text-forge-muted">
          Project source files are sent to the configured model provider. The agent can write files but cannot run shell commands.
        </div>
        {selectedFile && (
          <div className="truncate border border-[var(--border)] px-2.5 py-2 text-[11px] text-forge-muted" title={selectedFile}>
            Context file · {selectedFile}
          </div>
        )}
        {messages.map((message, index) => (
          <div key={`${message.role}-${index}`} className={`border p-3 ${message.role === "user" ? "border-forge-accent/20 bg-forge-accent/[0.05]" : "border-[var(--border)] bg-forge-bg"}`}>
            <p className="mb-1 text-[10px] font-bold uppercase text-forge-muted">{message.role === "user" ? "You" : "Agent"}</p>
            <p className="whitespace-pre-wrap text-xs leading-5 text-forge-text">{message.content}</p>
            {message.changedFiles && message.changedFiles.length > 0 && (
              <div className="mt-2 border-t border-[var(--border)] pt-2">
                <p className="mb-1 flex items-center gap-1 text-[10px] font-semibold text-forge-accent"><Check size={12} /> Updated files</p>
                {message.changedFiles.map((file) => <p key={file} className="truncate font-mono text-[10px] text-forge-muted" title={file}>{file}</p>)}
              </div>
            )}
          </div>
        ))}
        {taskMutation.isPending && (
          <div className="flex items-center gap-2 text-xs text-forge-muted"><Loader2 size={13} className="animate-spin" /> Reading project and applying changes</div>
        )}
        {taskError && <div role="alert" className="text-xs text-forge-signal">{taskError}</div>}
      </div>

      <form onSubmit={submitTask} className="mt-5 border border-[var(--border)] bg-forge-bg p-2">
        <textarea
          aria-label="Task for project agent"
          value={prompt}
          onChange={(event) => setPrompt(event.target.value)}
          disabled={!configured || sandboxStatus !== "RUNNING" || taskMutation.isPending}
          rows={3}
          maxLength={12000}
          placeholder={configured ? "Describe a code task..." : "Configure the model to send tasks"}
          className="w-full resize-y bg-transparent text-xs text-forge-text outline-none placeholder:text-forge-muted disabled:cursor-not-allowed"
        />
        <div className="flex items-center justify-between border-t border-[var(--border)] pt-2">
          <span className="text-[10px] text-forge-muted">Edits apply directly to this project</span>
          <button type="submit" disabled={!canSend} aria-label="Send task to agent" title="Send task" className="grid size-7 place-items-center bg-forge-accent text-[var(--primary-foreground)] transition hover:bg-forge-accent-strong disabled:cursor-not-allowed disabled:opacity-40">
            <Send size={12} />
          </button>
        </div>
      </form>
    </div>
  );
}

function StatusNotice({ children }: { children: string }) {
  return (
    <div className="border border-forge-signal/25 bg-forge-signal/[0.06] p-3">
      <div className="flex items-start gap-2 text-[11px] leading-5 text-forge-signal">
        <AlertCircle size={13} className="mt-0.5 shrink-0" />
        <span>{children}</span>
      </div>
    </div>
  );
}