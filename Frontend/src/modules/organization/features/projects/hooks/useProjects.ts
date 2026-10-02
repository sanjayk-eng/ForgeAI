import { keepPreviousData, useQuery, useQueryClient } from "@tanstack/react-query";
import { listProjects } from "../../../api/projects.api";
import type { PageParams } from "../../../../../shared/api/pagination";

export const projectsQueryKey = (organizationId: string) => ["projects", organizationId];
const syncPollingTimeoutMs = 60_000;

export function useProjects(accessToken: string, organizationId: string, params: PageParams = {}) {
  return useQuery({
    queryKey: [...projectsQueryKey(organizationId), params.page ?? 1, params.search ?? ""],
    queryFn: () => listProjects(accessToken, organizationId, params),
    enabled: Boolean(accessToken && organizationId),
    placeholderData: keepPreviousData,
    refetchInterval: (query) => {
      const now = Date.now();
      const projects = query.state.data?.items;
      const active = Array.isArray(projects) && projects.some((project) => {
        const repository = project.repository;
        if (!repository || !["PENDING", "SYNCING"].includes(repository.sync_status)) {
          return false;
        }

        const statusUpdatedAt = Date.parse(repository.updated_at);
        return Number.isFinite(statusUpdatedAt) && now - statusUpdatedAt < syncPollingTimeoutMs;
      });
      return active ? 1500 : false;
    },
  });
}

export function useRefreshProjects(organizationId: string) {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: projectsQueryKey(organizationId) });
}