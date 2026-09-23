import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { listMembers, removeMember, updateMemberRole } from "../../api/members.api";
import { useToast } from "../../../../shared/ui/useToast";

export function useMembersQuery(accessToken: string, workspaceId: string) {
  const toast = useToast();
  const queryClient = useQueryClient();

  const membersQuery = useQuery({
    queryKey: ["members", workspaceId],
    queryFn: () => listMembers(accessToken, workspaceId),
    enabled: Boolean(accessToken && workspaceId),
  });

  const roleMutation = useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: string }) =>
      updateMemberRole(accessToken, workspaceId, userId, role),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["members", workspaceId] }),
    onError: () => toast.pushError("Could not update member role"),
  });

  const removeMutation = useMutation({
    mutationFn: (userId: string) =>
      removeMember(accessToken, workspaceId, userId),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["members", workspaceId] }),
    onError: () => toast.pushError("Could not remove member"),
  });

  return {
    members: membersQuery.data ?? [],
    isLoading: membersQuery.isLoading,
    isError: membersQuery.isError,
    refetch: membersQuery.refetch,
    updateRole: (userId: string, role: string) => roleMutation.mutate({ userId, role }),
    removeMember: (userId: string) => removeMutation.mutate(userId),
    isUpdating: roleMutation.isPending || removeMutation.isPending,
  };
}
