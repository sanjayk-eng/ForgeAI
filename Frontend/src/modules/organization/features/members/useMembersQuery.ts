import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { listMembers, removeMember, updateMemberRole } from "../../api/members.api";
import { useToast } from "../../../../shared/ui/useToast";
import { useState } from "react";
import { useDebouncedValue } from "../../../../shared/hooks/useDebouncedValue";

export function useMembersQuery(accessToken: string, workspaceId: string) {
  const toast = useToast();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const debouncedSearch = useDebouncedValue(search);

  const membersQuery = useQuery({
    queryKey: ["members", workspaceId, page, debouncedSearch],
    queryFn: () => listMembers(accessToken, workspaceId, { page, perPage: 10, search: debouncedSearch }),
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
    members: membersQuery.data?.items ?? [],
    total: membersQuery.data?.total ?? 0,
    page,
    totalPages: membersQuery.data?.total_pages ?? 0,
    search,
    isSearching: search !== debouncedSearch || membersQuery.isFetching,
    setPage,
    setSearch: (value: string) => { setSearch(value); setPage(1); },
    isLoading: membersQuery.isLoading,
    isError: membersQuery.isError,
    refetch: membersQuery.refetch,
    updateRole: (userId: string, role: string) => roleMutation.mutate({ userId, role }),
    removeMember: (userId: string) => removeMutation.mutate(userId),
    isUpdating: roleMutation.isPending || removeMutation.isPending,
  };
}
