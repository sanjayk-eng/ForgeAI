import { QueryClientProvider } from "@tanstack/react-query";
import { Component, type ErrorInfo, type ReactNode } from "react";
import { RouterProvider } from "react-router-dom";
import { queryClient } from "./app/queryClient";
import { router } from "./app/router";
import { AuthProvider } from "./modules/auth/AuthContext";
import { ToastProvider } from "./shared/ui/ToastProvider";
import { ThemeProvider } from "./shared/ui/ThemeContext";

class AppErrorBoundary extends Component<
  { children: ReactNode },
  { hasError: boolean }
> {
  state = { hasError: false };

  static getDerivedStateFromError() {
    return { hasError: true };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error("AppErrorBoundary caught an error:", error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex h-screen items-center justify-center bg-[var(--bg)] p-6 text-center text-sm text-forge-text">
          <div className="max-w-md rounded-lg border border-[var(--border)] bg-[var(--panel)] p-6 shadow-sm">
            <h1 className="mb-2 text-lg font-semibold">Something went wrong</h1>
            <p className="text-forge-muted">The workspace UI could not load. Please refresh the page and try again.</p>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

function App() {
  return (
    <AppErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider>
          <ToastProvider>
            <AuthProvider>
              <RouterProvider router={router} />
            </AuthProvider>
          </ToastProvider>
        </ThemeProvider>
      </QueryClientProvider>
    </AppErrorBoundary>
  );
}

export default App;
