import { create } from "zustand";

type WorkspaceUiState = {
  mobileNavigationOpen: boolean;
  createWorkspaceOpen: boolean;
  inviteMemberOpen: boolean;
  openMobileNavigation: () => void;
  closeMobileNavigation: () => void;
  openCreateWorkspace: () => void;
  closeCreateWorkspace: () => void;
  openInviteMember: () => void;
  closeInviteMember: () => void;
};

export const useWorkspaceStore = create<WorkspaceUiState>((set) => ({
  mobileNavigationOpen: false,
  createWorkspaceOpen: false,
  inviteMemberOpen: false,
  openMobileNavigation: () => set({ mobileNavigationOpen: true }),
  closeMobileNavigation: () => set({ mobileNavigationOpen: false }),
  openCreateWorkspace: () => set({ createWorkspaceOpen: true }),
  closeCreateWorkspace: () => set({ createWorkspaceOpen: false }),
  openInviteMember: () => set({ inviteMemberOpen: true }),
  closeInviteMember: () => set({ inviteMemberOpen: false }),
}));
