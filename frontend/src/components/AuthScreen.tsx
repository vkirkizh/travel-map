import {useState} from "react";

type User = {
  id: string;
  username: string;
  email: string;
  display_name: string;
  avatar_url: string | null;
};

type Props = {
  apiBaseUrl: string;
  onAuthenticated: (user: User) => void;
};

export function AuthScreen({ apiBaseUrl, onAuthenticated }: Props) {
  const [mode, setMode] = useState<"login" | "register">("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [username, setUsername] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [error, setError] = useState<string | null>(null);

  function getErrorMessage(error: string): string {
    switch (error) {
      case "invalid credentials":
        return "Incorrect email or password.";

      case "user already exists":
        return "A user with this email or username already exists.";

      case "internal server error":
        return "Something went wrong. Please try again.";

      case "invalid request":
        return "Please check all fields and try again.";

      case "validation failed":
        return "Please check the highlighted fields and try again.";

      default:
        return "Something went wrong. Please try again.";
    }
  }

  async function submit() {
    setError(null);

    try {
      const url =
        mode === "login"
          ? `${apiBaseUrl}/api/auth/login`
          : `${apiBaseUrl}/api/auth/register`;

      const body =
        mode === "login"
          ? { email, password }
          : { username, email, password, display_name: displayName };

      const response = await fetch(url, {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
      });

      const payload = await response.json();

      if (!response.ok) {
        setError(getErrorMessage(payload.error));
        return;
      }

      onAuthenticated(payload.user);
    } catch {
      setError("Unable to connect to Travel Map right now.");
    }
  }

  return (
    <div className="app-page">
      <form
        className="auth-card"
        onSubmit={(e) => {
          e.preventDefault();
          void submit();
        }}
      >
        <div className="landing-eyebrow">Travel Map App</div>
        <h1>{mode === "login" ? "Welcome back" : "Create your account"}</h1>

        <div className="auth-tabs">
          <button
            type="button"
            className={mode === "login" ? "active" : ""}
            onClick={() => setMode("login")}
          >
            Login
          </button>
          <button
            type="button"
            className={mode === "register" ? "active" : ""}
            onClick={() => setMode("register")}
          >
            Register
          </button>
        </div>

        {mode === "register" && (
          <>
            <label>
              Username
              <input value={username} onChange={(e) => setUsername(e.target.value)} />
            </label>

            <label>
              Display name
              <input
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
              />
            </label>
          </>
        )}

        <label>
          Email
          <input value={email} onChange={(e) => setEmail(e.target.value)} />
        </label>

        <label>
          Password
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </label>

        {error && <div className="form-error">{error}</div>}

        <button className="primary-button"  type="submit">
          {mode === "login" ? "Login" : "Register"}
        </button>

        <a className="secondary-link" href="/vkirkizh/">
          View public demo map
        </a>
      </form>
    </div>
  );
}

export type { User };
