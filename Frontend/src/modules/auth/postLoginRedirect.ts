import { createOrganization, listOrganizations } from "../organization/api/organization.api";

const redirectAfterLoginKey = "redirect_after_login";

export async function getPostLoginRedirect(accessToken: string): Promise<string> {
  const redirectAfterLogin = localStorage.getItem(redirectAfterLoginKey);
  if (redirectAfterLogin?.startsWith("/accept-invite/")) {
    localStorage.removeItem(redirectAfterLoginKey);
    return redirectAfterLogin;
  }

  const organizations = await listOrganizations(accessToken);
  const organization = organizations[0]
    ?? await createOrganization(accessToken, "My Organization");
  return `/organizations/${encodeURIComponent(organization.id)}`;
}