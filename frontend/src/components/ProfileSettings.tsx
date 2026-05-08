import {useState} from "react";
import type {User} from "./AuthScreen";

type Props = {
  apiBaseUrl: string;
  user: User;
  onUserUpdated: (user: User) => void;
};

function getSettingsErrorMessage(error: string): string {
  switch (error) {
    case "user already exists":
      return "This email is already used by another account.";
    case "validation failed":
      return "Please check the form and try again.";
    default:
      return "Something went wrong. Please try again.";
  }
}

export function ProfileSettings({ apiBaseUrl, user, onUserUpdated }: Props) {
  const [displayName, setDisplayName] = useState(user.display_name);
  const [email, setEmail] = useState(user.email);
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function save(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    setError(null);
    setSuccess(null);
    setIsSubmitting(true);

    try {
      const response = await fetch(`${apiBaseUrl}/api/me`, {
        method: "PATCH",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          display_name: displayName,
          email,
          current_password: currentPassword || null,
          new_password: newPassword || null,
        }),
      });

      const payload = await response.json();

      if (!response.ok) {
        setError(getSettingsErrorMessage(payload.error));
        return;
      }

      onUserUpdated(payload.user);
      setCurrentPassword("");
      setNewPassword("");
      setSuccess("Profile updated.");
    } catch {
      setError("Unable to update profile.");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="app-page">
      <form className="settings-card" onSubmit={save}>
        <div className="settings-header">
          <div>
            <div className="landing-eyebrow">Settings</div>
            <h1>Profile settings</h1>
            <p>Update your public profile and account details.</p>
          </div>

          <img className="settings-avatar" src={user.avatar_url} alt={user.display_name} />
        </div>

        <label>
          Username
          <input value={user.username} disabled />
        </label>

        <label>
          Display name
          <input value={displayName} onChange={(e) => setDisplayName(e.target.value)} />
        </label>

        <label>
          Email
          <input value={email} onChange={(e) => setEmail(e.target.value)} />
        </label>

        <div className="settings-divider" />

        <label>
          Current password
          <input
            type="password"
            value={currentPassword}
            onChange={(e) => setCurrentPassword(e.target.value)}
            placeholder="Required only when changing password"
          />
        </label>

        <label>
          New password
          <input
            type="password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            placeholder="Leave empty to keep current password"
          />
        </label>

        {error && <div className="form-error">{error}</div>}
        {success && <div className="form-success">{success}</div>}

        <div className="dashboard-actions">
          <button className="primary-button" type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Saving..." : "Save changes"}
          </button>

          <a className="ghost-link" href="/app/">
            Back
          </a>
        </div>
      </form>
    </div>
  );
}
