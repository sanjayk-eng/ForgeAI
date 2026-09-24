import { useQuery, useQueryClient } from "@tanstack/react-query";
import { listProjects } from "../../../api/projects.api";

export const projectsQueryKey = (workspaceId: string) => ["projects", workspaceId];

export function useProjects(accessToken: string, workspaceId: string) {
  return useQuery({
    queryKey: projectsQueryKey(workspaceId),
    queryFn: () => listProjects(accessToken, workspaceId),
    enabled: Boolean(accessToken && workspaceId),
    refetchInterval: (query) => {
      const active = query.state.data?.some((project) =>
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