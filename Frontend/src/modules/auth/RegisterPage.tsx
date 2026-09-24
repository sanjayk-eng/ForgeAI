import { useState } from "react";
import { GitBranch, Globe } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { AuthForm, PasswordField } from "./AuthForm";
import { AuthSwitch } from "./AuthLayout";
import { oauthUrl } from "./api";
import { useAuth } from "./useAuth";
import { useToast } from "../../shared/ui/useToast";

const inputClass =
  "h-12 w-full rounded-[10px] border border-[var(--border)] bg-[var(--input)] px-3.5 text-sm text-forge-text outline-none transition placeholder:text-forge-muted focus:border-forge-accent focus:ring-4 focus:ring-forge-accent/10";

const fieldClass = "grid gap-2 text-xs font-bold text-forge-soft";
const oauthButtonClass =
  "flex h-11 items-center justify-center gap-2 rounded-[10px] border border-[var(--border)] bg-[var(--surface-subtle)] text-[13px] font-bold text-forge-text no-underline transition hover:border-forge-accent/45 hover:bg-forge-accent/10 focus-visible:outline focus-visible:outline-2 focus-visible:outline-forge-accent";

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

  return (
    <div className="mx-auto w-full max-w-[420px]">
      <div className="mb-7 lg:mb-5">
        <p className="m-0 font-mono text-[11px] uppercase tracking-[0.08em] text-forge-accent">
          Start here
        </p>
        <h2 className="my-2 text-[clamp(30px,3.2vw,40px)] font-extrabold leading-[1.04] tracking-[-0.05em]">
          Make room for better work.
        </h2>
        <p className="m-0 leading-[1.6] text-forge-soft">
          Create your workspace identity in less than a minute.
        </p>
      </div>

      <AuthForm onSubmit={submit} submitLabel="Create workspace" error={error}>
        <label className={fieldClass}>
          <span>Your name</span>
          <input
            className={inputClass}
            required
            maxLength={150}
            autoComplete="name"
            value={name}
            onChange={(event) => setName(event.target.value)}
          />
        </label>

        <label className={fieldClass}>
          <span>Email address</span>
          <input
            className={inputClass}
            required
            type="email"
            autoComplete="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
          />
        </label>

        <PasswordField value={password} onChange={setPassword} />
      </AuthForm>

      <div className="my-6 flex items-center gap-3 font-mono text-[10px] uppercase text-forge-muted before:h-px before:flex-1 before:bg-[var(--border)] after:h-px after:flex-1 after:bg-[var(--border)] lg:my-4">
        <span>or register with</span>
      </div>

      <div className="grid grid-cols-2 gap-2 max-[470px]:grid-cols-1">
        <a className={oauthButtonClass} href={oauthUrl("github")}>
          <GitBranch size={17} />
          GitHub
        </a>
        <a className={oauthButtonClass} href={oauthUrl("google")}>
          <Globe size={17} />
          Google
        </a>
      </div>

      <AuthSwitch prompt="Already have access?" label="Sign in" to="/login" />
    </div>
  );
}
