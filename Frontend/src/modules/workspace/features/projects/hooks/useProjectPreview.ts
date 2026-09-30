import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { getProjectPreview } from "../../../api/preview.api";

const previewWaitTimeoutMs = 60_000;

export function useProjectPreview(accessToken: string | null, projectId: string) {
  const previewQuery = useQuery({
    queryKey: ["project-preview", projectId],
    queryFn: () => {
      if (!accessToken) throw new Error("Authentication is required to load the preview");
      return getProjectPreview(accessToken, projectId);
    },
    enabled: Boolean(accessToken && projectId),
    refetchInterval: 5000,
    retry: false,
  });
  const isWaitingForServer = previewQuery.data?.status === "starting" || previewQuery.data?.status === "application_unavailable";
  const [waitTimedOut, setWaitTimedOut] = useState(false);

  useEffect(() => {
    if (!isWaitingForServer) {
      setWaitTimedOut(false);
      return;
    }
    const timer = setTimeout(() => setWaitTimedOut(true), previewWaitTimeoutMs);
    return () => clearTimeout(timer);
  }, [isWaitingForServer, projectId]);

  return { ...previewQuery, waitTimedOut };
}