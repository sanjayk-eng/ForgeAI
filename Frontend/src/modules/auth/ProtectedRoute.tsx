import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "./useAuth";

export function ProtectedRoute() {
  const { user, loading } = useAuth();

  if (loading) return <div className="loading-screen"><span className="loader-dot" /> Loading workspace</div>;
  if (!user) return <Navigate to="/login" replace />;
  return <Outlet />;
}
