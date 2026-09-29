import { keepPreviousData, useQuery, useQueryClient } from "@tanstack/react-query";
import { listProjects } from "../../../api/projects.api";
import type { PageParams } from "../../../../../shared/api/pagination";

export const projectsQueryKey = (workspaceId: string) => ["projects", workspaceId];
const syncPollingTimeoutMs = 60_000;

export function useProjects(accessToken: string, workspaceId: string, params: PageParams = {}) {
  return useQuery({
    queryKey: [...projectsQueryKey(workspaceId), params.page ?? 1, params.search ?? ""],
    queryFn: () => listProjects(accessToken, workspaceId, params),
    enabled: Boolean(accessToken && workspaceId),
    placeholderData: keepPreviousData,
    refetchInterval: (query) => {
      const now = Date.now();
      const active = query.state.data?.items.some((project) => {
        const repository = project.repository;
        if (!repository || !["PENDING", "SYNCING"].includes(repository.sync_status)) {
          return false;
        }

        const statusUpdatedAt = Date.parse(repository.updated_at);
        return Number.isFinite(statusUpdatedAt) && now - statusUpdatedAt < syncPollingTimeoutMs;
      }
      );
      return active ? 1500 : false;
    },
  });
}

export function useRefreshProjects(workspaceId: string) {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: projectsQueryKey(workspaceId) });
}