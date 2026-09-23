import { Navigate, createBrowserRouter } from "react-router-dom";
import { AuthLayout } from "../modules/auth/AuthLayout";
import { LoginPage } from "../modules/auth/LoginPage";
import { OAuthCallbackPage } from "../modules/auth/OAuthCallbackPage";
import { RegisterPage } from "../modules/auth/RegisterPage";
import { ProtectedRoute } from "../modules/auth/ProtectedRoute";
import { WorkspaceLayout } from "../modules/workspace/WorkspaceLayout";
import { OverviewPage } from "../modules/workspace/features/overview/OverviewPage";
import { MembersPage } from "../modules/workspace/features/members/MembersPage";
import { SettingsPage } from "../modules/workspace/features/settings/SettingsPage";
import { AcceptInvitePage } from "../modules/workspace/features/invites/AcceptInvitePage";

export const router = createBrowserRouter([
  {
    element: <AuthLayout />,
    children: [
      { path: "/login", element: <LoginPage /> },
      { path: "/register", element: <RegisterPage /> },
      { path: "/auth/callback", element: <OAuthCallbackPage /> },
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
        path: "/workspace",
        element: <WorkspaceLayout />,
        children: [
          { index: true, element: <OverviewPage /> },
          { path: "members", element: <MembersPage /> },
          { path: "settings", element: <SettingsPage /> },
        ],
      },
    ],
  },
  { path: "/", element: <Navigate to="/login" replace /> },
  { path: "*", element: <Navigate to="/" replace /> },
]);
