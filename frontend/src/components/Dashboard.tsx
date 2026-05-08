import type {User} from "./AuthScreen";
import {PlacesManager} from "./PlacesManager";
import {FlightsManager} from "./FlightsManager";

type Props = {
  apiBaseUrl: string;
  user: User;
  onLogout: () => void;
};

export function Dashboard({ apiBaseUrl, user, onLogout }: Props) {
  async function logout() {
    await fetch(`${apiBaseUrl}/api/auth/logout`, {
      method: "POST",
      credentials: "include",
    });

    onLogout();
  }

  return (
    <div className="app-page">
      <div className="dashboard-card">
        <div>
          <div className="landing-eyebrow">Travel Map</div>
          <h1>Hello, {user.display_name}</h1>
          <p>
            Manage your profile, visited places and flights.
          </p>
        </div>

        <div className="dashboard-actions">
          <a className="landing-link" href={`/${user.username}/`} target="_blank">
            View public map
          </a>
          <a className="ghost-link" href="/app/settings/">
            Settings
          </a>
          <button className="ghost-button" type="button" onClick={logout}>
            Logout
          </button>
        </div>

        <PlacesManager apiBaseUrl={apiBaseUrl} />

        <FlightsManager apiBaseUrl={apiBaseUrl} />
      </div>
    </div>
  );
}
