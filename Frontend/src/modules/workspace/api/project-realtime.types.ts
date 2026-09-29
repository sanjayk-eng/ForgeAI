export type ProjectRealtimeEventName =
  | "file.created"
  | "file.changed"
  | "file.deleted"
  | "file.renamed"
  | "git.status.changed"
  | "sandbox.status.changed"
  | "agent.started"
  | "agent.progress"
  | "agent.completed"
  | "sync.required";

export interface ProjectRealtimeEvent {
  version: number;
  event: ProjectRealtimeEventName;
  workspace_id?: string;
  project_id: string;
  sandbox_id?: string;
  path?: string;
  old_path?: string;
  change_type?: string;
  is_directory?: boolean;
  status?: string;
  message?: string;
}

export interface SequencedProjectRealtimeEvent {
  sequence: number;
  event: ProjectRealtimeEvent;
}