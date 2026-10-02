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
  { path: "/", element: <Navigate to="/login" replace /> },
  { path: "*", element: <Navigate to="/" replace /> },
]);
