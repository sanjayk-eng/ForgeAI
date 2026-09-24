import { useQuery, useQueryClient } from "@tanstack/react-query";
import { listProjects } from "../../../api/projects.api";
import type { PageParams } from "../../../../../shared/api/pagination";

export const projectsQueryKey = (workspaceId: string) => ["projects", workspaceId];

export function useProjects(accessToken: string, workspaceId: string, params: PageParams = {}) {
  return useQuery({
    queryKey: [...projectsQueryKey(workspaceId), params.page ?? 1, params.search ?? ""],
    queryFn: () => listProjects(accessToken, workspaceId, params),
    enabled: Boolean(accessToken && workspaceId),
    refetchInterval: (query) => {
      const active = query.state.data?.items.some((project) =>
        ["PENDING", "SYNCING"].includes(project.repository?.sync_status ?? ""),
      );
      return active ? 1500 : false;
    },
  });
}

export function useRefreshProjects(workspaceId: string) {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: projectsQueryKey(workspaceId) });
}