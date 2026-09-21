import { useState } from "react";
import { GitBranch, Globe } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { AuthForm, PasswordField } from "./AuthForm";
import { AuthSwitch } from "./AuthLayout";
import { oauthUrl } from "./api";
import { useAuth } from "./useAuth";
import { useToast } from "../../shared/ui/useToast";

export function LoginPage() {
  const navigate = useNavigate();
  const { signIn } = useAuth();
  const { pushError } = useToast();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function submit() {
    try {
      setError("");
      await signIn({ email, password });
      navigate("/workspace", { replace: true });
    } catch (reason) {
      const message = reason instanceof Error ? reason.message : "Unable to sign in";
      setError(message);
      pushError(message);
    }
  }

  return <div className="auth-content">
    <div className="auth-heading"><p className="eyebrow">Welcome back</p><h2>Pick up where you left off.</h2><p>Sign in to return to your ForgeAI workspace.</p></div>
    <AuthForm onSubmit={submit} submitLabel="Enter workspace" error={error}>
      <label className="field"><span>Email address</span><input required type="email" autoComplete="email" value={email} onChange={(event) => setEmail(event.target.value)} /></label>
      <PasswordField value={password} onChange={setPassword} />
    </AuthForm>
    <div className="or-divider"><span>or continue with</span></div>
    <div className="oauth-row">
      <a className="oauth-button" href={oauthUrl("github")}><GitBranch size={17} /> GitHub</a>
      <a className="oauth-button" href={oauthUrl("google")}><Globe size={17} /> Google</a>
    </div>
    <AuthSwitch prompt="New to ForgeAI?" label="Create an account" to="/register" />
  </div>;
}
