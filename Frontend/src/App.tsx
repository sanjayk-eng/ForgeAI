import { RouterProvider } from "react-router-dom";
import { QueryClientProvider } from "@tanstack/react-query";
import { router } from "./app/router";
import { queryClient } from "./app/queryClient";
import { AuthProvider } from "./modules/auth/AuthContext";
import { ToastProvider } from "./shared/ui/ToastProvider";

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ToastProvider><AuthProvider><RouterProvider router={router} /></AuthProvider></ToastProvider>
    </QueryClientProvider>
  );
}

export default App;
