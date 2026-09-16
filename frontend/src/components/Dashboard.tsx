import {useState} from "react";
import type {User} from "../types";
import {PlacesManager} from "./PlacesManager";
import "./Dashboard.css";

type Props = {
  apiBaseUrl: string;
  user: User;
  onLogout: () => void;
  onUnauthorized: () => void;
};

export function Dashboard({apiBaseUrl, user, onLogout, onUnauthorized}: Props) {
  const [logoutError, setLogoutError] = useState<string | null>(null);

  async function logout() {
    setLogoutError(null);

    try {
      const response = await fetch(`${apiBaseUrl}/api/auth/logout`, {
        method: "POST",
        credentials: "include",
      });

      if (!response.ok) {
        setLogoutError("Unable to log out. Please try again.");
        return;
      }

      onLogout();
    } catch {
      setLogoutError("Unable to log out. Please try again.");
    }
  }

  return (
    <div className="dashboard-page">
      <div className="dashboard-card">
        <div>
          <div className="dashboard-eyebrow">Travel Map</div>
          <h1>Hello, {user.display_name}</h1>
          <p>
            Manage your profile and visited places.
          </p>
        </div>

        <div className="dashboard-actions">
          <a className="dashboard-public-link" href={`/${user.username}/`} target="_blank">
            View public map
          </a>
          <a className="dashboard-settings-link" href="/app/settings/">
            Settings
          </a>
          <button className="dashboard-logout-button" type="button" onClick={logout}>
            Logout
          </button>
        </div>

        {logoutError && <div className="dashboard-error">{logoutError}</div>}

        <PlacesManager
          apiBaseUrl={apiBaseUrl}
          onUnauthorized={onUnauthorized}
        />
      </div>
    </div>
  );
}
