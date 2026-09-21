import { useState } from "react";
import { GitBranch, Globe } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { AuthForm, PasswordField } from "./AuthForm";
import { AuthSwitch } from "./AuthLayout";
import { oauthUrl } from "./api";
import { useAuth } from "./useAuth";
import { useToast } from "../../shared/ui/useToast";

export function RegisterPage() {
  const navigate = useNavigate();
  const { signUp } = useAuth();
  const { pushError } = useToast();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function submit() {
    try {
      setError("");
      await signUp({ name, email, password });
      navigate("/workspace", { replace: true });
    } catch (reason) {
      const message = reason instanceof Error ? reason.message : "Unable to create account";
      setError(message);
      pushError(message);
    }
  }

  return <div className="auth-content">
    <div className="auth-heading"><p className="eyebrow">Start here</p><h2>Make room for better work.</h2><p>Create your workspace identity in less than a minute.</p></div>
    <AuthForm onSubmit={submit} submitLabel="Create workspace" error={error}>
      <label className="field"><span>Your name</span><input required maxLength={150} autoComplete="name" value={name} onChange={(event) => setName(event.target.value)} /></label>
      <label className="field"><span>Email address</span><input required type="email" autoComplete="email" value={email} onChange={(event) => setEmail(event.target.value)} /></label>
      <PasswordField value={password} onChange={setPassword} />
    </AuthForm>
    <div className="or-divider"><span>or register with</span></div>
    <div className="oauth-row"><a className="oauth-button" href={oauthUrl("github")}><GitBranch size={17} /> GitHub</a><a className="oauth-button" href={oauthUrl("google")}><Globe size={17} /> Google</a></div>
    <AuthSwitch prompt="Already have access?" label="Sign in" to="/login" />
  </div>;
}
