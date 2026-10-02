import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { listInvites, revokeInvite } from "../../api/invites.api";
import { useToast } from "../../../../shared/ui/useToast";

export function useInvitesQuery(accessToken: string, organizationId: string) {
  const toast = useToast();
  const queryClient = useQueryClient();

  const invitesQuery = useQuery({
    queryKey: ["invites", organizationId],
    queryFn: () => listInvites(accessToken, organizationId),
    enabled: Boolean(accessToken && organizationId),
  });

  const revokeMutation = useMutation({
    mutationFn: (inviteId: string) =>
      revokeInvite(accessToken, organizationId, inviteId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["invites", organizationId] });
      toast.pushSuccess("Invitation revoked");
    },
    onError: () => toast.pushError("Could not revoke invitation"),
  });

  return {
    invites: invitesQuery.data ?? [],
    isLoading: invitesQuery.isLoading,
    isError: invitesQuery.isError,
    refetch: invitesQuery.refetch,
    revokeInvite: (inviteId: string) => revokeMutation.mutate(inviteId),
    isRevoking: revokeMutation.isPending,
  };
}
