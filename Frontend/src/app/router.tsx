import { Navigate, createBrowserRouter } from "react-router-dom";
import { AuthLayout } from "../modules/auth/AuthLayout";
import { LoginPage } from "../modules/auth/LoginPage";
import { OAuthCallbackPage } from "../modules/auth/OAuthCallbackPage";
import { ProtectedRoute } from "../modules/auth/ProtectedRoute";
import { RegisterPage } from "../modules/auth/RegisterPage";
import { WorkspacePage } from "../modules/workspace/WorkspacePage";

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
    element: <ProtectedRoute />,
    children: [{ path: "/workspace", element: <WorkspacePage /> }],
  },
  { path: "/", element: <Navigate to="/workspace" replace /> },
  { path: "*", element: <Navigate to="/" replace /> },
]);
