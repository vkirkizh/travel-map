import {useCallback, useEffect, useState} from "react";
import {AuthScreen} from "../components/AuthScreen";
import {Dashboard} from "../components/Dashboard";
import {ProfileSettings} from "../components/ProfileSettings";
import type {User} from "../types";

type Props = {
  apiBaseUrl: string;
  pathname: string;
};

export function PrivateAppPage({apiBaseUrl, pathname}: Props) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  const handleUnauthorized = useCallback(() => {
    setUser(null);
  }, []);

  useEffect(() => {
    document.title = pathname.startsWith("/app/settings")
      ? "Travel Map — Settings"
      : "Travel Map — Dashboard";

    async function loadCurrentUser() {
      setIsLoading(true);
      setLoadError(null);

      try {
        const response = await fetch(`${apiBaseUrl}/api/me`, {
          credentials: "include",
        });

        if (response.status === 401) {
          setUser(null);
          return;
        }

        if (!response.ok) {
          throw new Error("Failed to load current user");
        }

        const payload = await response.json();
        setUser(payload.user);
      } catch {
        setLoadError("Unable to load Travel Map right now.");
      } finally {
        setIsLoading(false);
      }
    }

    void loadCurrentUser();
  }, [apiBaseUrl, pathname]);

  if (isLoading) {
    return <div className="loading">Loading Travel Map App...</div>;
  }

  if (loadError) {
    return <div className="error">{loadError}</div>;
  }

  if (!user) {
    return (
      <AuthScreen
        apiBaseUrl={apiBaseUrl}
        onAuthenticated={(authenticatedUser) => setUser(authenticatedUser)}
      />
    );
  }

  if (pathname.startsWith("/app/settings")) {
    return (
      <ProfileSettings
        apiBaseUrl={apiBaseUrl}
        user={user}
        onUserUpdated={(updatedUser) => setUser(updatedUser)}
        onUnauthorized={handleUnauthorized}
      />
    );
  }

  return (
    <Dashboard
      apiBaseUrl={apiBaseUrl}
      user={user}
      onLogout={() => setUser(null)}
      onUnauthorized={handleUnauthorized}
    />
  );
}
