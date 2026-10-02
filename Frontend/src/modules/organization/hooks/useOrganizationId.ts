import { useParams } from "react-router-dom";

export function useOrganizationId() {
  return useParams<{ organizationId: string }>().organizationId ?? "";
}
