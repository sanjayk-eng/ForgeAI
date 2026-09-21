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
    return () => { active = false; };
  }, [acceptTokens, location.search, navigate, pushError]);

  return <div className="auth-content callback-state">
    {error ? <><p className="eyebrow">Could not finish sign in</p><h2>That connection did not complete.</h2><p className="form-error">{error}</p><Link className="text-link" to="/login">Return to sign in</Link></> : <><LoaderCircle className="spin" size={24} /><p>Securing your workspace...</p></>}
  </div>;
}
