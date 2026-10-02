import { useSearchParams } from "react-router-dom";

export function useOrganizationId() {
  return useSearchParams()[0].get("organization") ?? "";
}
