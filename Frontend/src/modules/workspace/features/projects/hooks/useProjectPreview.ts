import { useQuery } from "@tanstack/react-query";
import { getProjectPreview } from "../../../api/preview.api";

export function useProjectPreview(accessToken: string | null, projectId: string) {
  return useQuery({
    queryKey: ["project-preview", projectId],
    queryFn: () => {
      if (!accessToken) throw new Error("Authentication is required to load the preview");
      return getProjectPreview(accessToken, projectId);
    },
    enabled: Boolean(accessToken && projectId),
    refetchInterval: 5000,
    retry: false,
  });
}