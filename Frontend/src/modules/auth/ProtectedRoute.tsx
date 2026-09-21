import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "./useAuth";

export function ProtectedRoute() {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="grid min-h-screen place-content-center gap-3 font-mono text-xs text-forge-soft">
        <span className="mx-auto size-2.5 animate-pulse rounded-full bg-forge-accent" />
        Loading workspace
      </div>
    );
  }

  if (!user) return <Navigate to="/login" replace />;
  return <Outlet />;
}
