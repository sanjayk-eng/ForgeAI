import { useSearchParams } from "react-router-dom";

export function useWorkspaceId() {
  return useSearchParams()[0].get("workspace") ?? "";
}
