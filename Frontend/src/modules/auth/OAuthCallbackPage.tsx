import { useEffect, useState } from "react";
import { LoaderCircle } from "lucide-react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { authenticateOAuth } from "./api";
import { useAuth } from "./useAuth";
import { useToast } from "../../shared/ui/useToast";

export function OAuthCallbackPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const { acceptTokens } = useAuth();
  const { pushError } = useToast();
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;

    async function finish() {
      const query = new URLSearchParams(location.search);
      const provider = query.get("type");
      const code = query.get("code");

      if (!provider || !code) {
        setError("The OAuth callback is missing provider or code information.");
        return;
      }

      try {
        await acceptTokens(await authenticateOAuth(provider, code));
        if (active) navigate("/workspace", { replace: true });
      } catch (reason) {
        const message = reason instanceof Error ? reason.message : "OAuth sign-in failed";
        if (active) {
          setError(message);
          pushError(message);
        }
      }
    }

    void finish();
    return () => {
      active = false;
    };
  }, [acceptTokens, location.search, navigate, pushError]);

  return (
    <div className="mx-auto w-full max-w-[420px] text-center">
      {error ? (
        <>
          <p className="m-0 font-mono text-[11px] uppercase tracking-[0.08em] text-forge-accent">
            Could not finish sign in
          </p>
          <h2 className="my-2 text-4xl font-extrabold tracking-[-0.05em]">
            That connection did not complete.
          </h2>
          <p className="m-0 text-xs leading-[1.4] text-red-300" role="alert">
            {error}
          </p>
          <Link className="mt-4 inline-flex items-center font-bold text-forge-accent no-underline" to="/login">
            Return to sign in
          </Link>
        </>
      ) : (
        <>
          <LoaderCircle className="mx-auto mb-4 animate-spin text-forge-accent" size={24} />
          <p className="text-forge-soft">Securing your workspace...</p>
        </>
      )}
    </div>
  );
}
