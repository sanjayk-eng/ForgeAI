import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useRef } from "react";
import { getSandboxByProject } from "../../../api/sandbox.api";

export function useSandbox(accessToken: string | null, projectId: string | null, ensureMissing = false) {
  const queryClient = useQueryClient();
  const failedRecoveryStartedAt = useRef<number | null>(null);

  const sandboxQuery = useQuery({
    queryKey: ["sandbox", projectId],
    queryFn: () => {
      if (!accessToken || !projectId) {
        throw new Error("Missing access token or project ID");
      }
      return getSandboxByProject(accessToken, projectId, ensureMissing);
    },
    enabled: !!accessToken && !!projectId,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      if (status === "FAILED" && ensureMissing) {
        failedRecoveryStartedAt.current ??= Date.now();
        return Date.now() - failedRecoveryStartedAt.current < 30_000 ? 2000 : false;
      }
      failedRecoveryStartedAt.current = null;
      if (status === "CREATING" || status === "STARTING" || status === "STOPPING" || status === "RESTARTING") {
        return 2000;
      }
      return false;
    },
    retry: (failureCount, error: any) => {
      if (error?.response?.status === 404 || error?.status === 404) {
        return failureCount < 2;
      }
      return failureCount < 2;
    },
    retryDelay: (attemptIndex, error: any) => {
      if (error?.response?.status === 404 || error?.status === 404) {
        return 1000 * (attemptIndex + 1);
      }
      return Math.min(1000 * 2 ** attemptIndex, 10000);
    },
  });

  const refetchSandbox = () => {
    void queryClient.invalidateQueries({ queryKey: ["sandbox", projectId] });
  };

  return {
    sandbox: sandboxQuery.data,
    isLoading: sandboxQuery.isLoading,
    isFetching: sandboxQuery.isFetching,
    error: sandboxQuery.error,
    refetch: refetchSandbox,
  };
}
