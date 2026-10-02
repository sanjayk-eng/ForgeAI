import { useQuery } from "@tanstack/react-query";
import { Navigate, createBrowserRouter } from "react-router-dom";
import { AuthLayout } from "../modules/auth/AuthLayout";
import { LoginPage } from "../modules/auth/LoginPage";
import { OAuthCallbackPage } from "../modules/auth/OAuthCallbackPage";
import { RegisterPage } from "../modules/auth/RegisterPage";
import { VerifyEmailPage } from "../modules/auth/VerifyEmailPage";
import { ProtectedRoute } from "../modules/auth/ProtectedRoute";
import { OrganizationLayout } from "../modules/organization/OrganizationLayout";
import { OverviewPage } from "../modules/organization/features/overview/OverviewPage";
import { MembersPage } from "../modules/organization/features/members/MembersPage";
import { SettingsPage } from "../modules/organization/features/settings/SettingsPage";
import { AcceptInvitePage } from "../modules/organization/features/invites/AcceptInvitePage";
import { ProjectsPage } from "../modules/organization/features/projects/ProjectsPage";
import { useAuth } from "../modules/auth/useAuth";
import { getPostLoginRedirect } from "../modules/auth/postLoginRedirect";

export const router = createBrowserRouter([
  {
    element: <AuthLayout />,
    children: [
      { path: "/login", element: <LoginPage /> },
      { path: "/register", element: <RegisterPage /> },
      { path: "/auth/callback", element: <OAuthCallbackPage /> },
      { path: "/auth/verify", element: <VerifyEmailPage /> },
    ],
  },
  {
    path: "/accept-invite/:token",
    element: <AcceptInvitePage />,
  },
  {
    path: "/accept-invite/id/:inviteId",
    element: <AcceptInvitePage />,
  },
  {
    element: <ProtectedRoute />,
    children: [
      {
        path: "/organizations/:organizationId",
        element: <OrganizationLayout />,
        children: [
          { index: true, element: <OverviewPage /> },
          { path: "projects", element: <ProjectsPage /> },
          {
            path: "projects/:projectId",
            lazy: async () => {
              const { ProjectDetailPage } = await import(
                "../modules/organization/features/projects/ProjectDetailPage"
              );
              return { Component: ProjectDetailPage };
            },
          },
          { path: "members", element: <MembersPage /> },
          { path: "settings", element: <SettingsPage /> },
        ],
      },
    ],
  },
  { path: "/", element: <HomeRedirect /> },
  { path: "*", element: <Navigate to="/" replace /> },
]);

function HomeRedirect() {
  const { user, tokens, loading } = useAuth();
  const redirectQuery = useQuery({
    queryKey: ["post-login-redirect", tokens?.access_token],
    queryFn: () => getPostLoginRedirect(tokens!.access_token),
    enabled: !loading && Boolean(user && tokens),
    retry: false,
  });

  if (loading) return <div className="grid min-h-screen place-content-center">Loading ForgeAI...</div>;
  if (!user || !tokens) return <Navigate to="/login" replace />;
  if (redirectQuery.isError) {
    return <div className="grid min-h-screen place-content-center gap-3 text-center" role="alert">
      <p>Could not load your organization.</p>
      <button className="text-forge-accent" onClick={() => void redirectQuery.refetch()}>Try again</button>
    </div>;
  }
  if (!redirectQuery.data) return <div className="grid min-h-screen place-content-center">Loading your organization...</div>;
  return <Navigate to={redirectQuery.data} replace />;
}
