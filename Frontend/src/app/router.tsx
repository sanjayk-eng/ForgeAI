import { Navigate, createBrowserRouter } from "react-router-dom";
import { AuthLayout } from "../modules/auth/AuthLayout";
import { LoginPage } from "../modules/auth/LoginPage";
import { OAuthCallbackPage } from "../modules/auth/OAuthCallbackPage";
import { RegisterPage } from "../modules/auth/RegisterPage";
import { ProtectedRoute } from "../modules/auth/ProtectedRoute";
import { WorkspaceLayout } from "../modules/workspace/WorkspaceLayout";
import { WorkspaceOverview } from "../modules/workspace/pages/WorkspaceOverview";
import { WorkspaceMembers } from "../modules/workspace/pages/WorkspaceMembers";
import { WorkspaceSettings } from "../modules/workspace/pages/WorkspaceSettings";
import { AcceptInvitePage } from "../modules/workspace/pages/AcceptInvitePage";

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
          { index: true, element: <WorkspaceOverview /> },
          { path: "members", element: <WorkspaceMembers /> },
          { path: "settings", element: <WorkspaceSettings /> },
        ],
      },
    ],
  },
  { path: "/", element: <Navigate to="/login" replace /> },
  { path: "*", element: <Navigate to="/" replace /> },
]);
