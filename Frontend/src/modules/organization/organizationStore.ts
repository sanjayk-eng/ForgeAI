import { create } from "zustand";

type OrganizationUiState = {
  mobileNavigationOpen: boolean;
  createOrganizationOpen: boolean;
  inviteMemberOpen: boolean;
  openMobileNavigation: () => void;
  closeMobileNavigation: () => void;
  openCreateOrganization: () => void;
  closeCreateOrganization: () => void;
  openInviteMember: () => void;
  closeInviteMember: () => void;
};

export const useOrganizationStore = create<OrganizationUiState>((set) => ({
  mobileNavigationOpen: false,
  createOrganizationOpen: false,
  inviteMemberOpen: false,
  openMobileNavigation: () => set({ mobileNavigationOpen: true }),
  closeMobileNavigation: () => set({ mobileNavigationOpen: false }),
  openCreateOrganization: () => set({ createOrganizationOpen: true }),
  closeCreateOrganization: () => set({ createOrganizationOpen: false }),
  openInviteMember: () => set({ inviteMemberOpen: true }),
  closeInviteMember: () => set({ inviteMemberOpen: false }),
}));
