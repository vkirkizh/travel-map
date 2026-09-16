import {useEffect, useState} from "react";
import {CircleMarker, MapContainer, Popup, TileLayer, AttributionControl, useMap} from "react-leaflet";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import "./App.css";
import {AuthScreen, type User} from "./components/AuthScreen";
import {Dashboard} from "./components/Dashboard";
import {ProfileSettings} from "./components/ProfileSettings";

type Place = {
  id: string;
  title: string;
  country_code: string;
  lat: number;
  lng: number;
};

type PublicUser = {
  username: string;
  display_name: string;
  avatar_url: string;
};

type Stats = {
  countries_visited: number;
  places_visited: number;
};

type MapResponse = {
  user: PublicUser;
  places: Place[];
  stats: Stats;
};

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";

function App() {
  const pathname = window.location.pathname;

  if (pathname === "/" || pathname === "") {
    return <LandingPage />;
  }

  if (pathname === "/app" || pathname === "/app/" || pathname === "/app/settings" || pathname === "/app/settings/") {
    return <PrivateAppPage pathname={pathname} />;
  }

  const username = pathname.replace(/^\/+|\/+$/g, "");

  return <PublicMapPage username={username} />;
}

function LandingPage() {
  useEffect(() => {
    setPageTitle("Travel Map — Share your travels");
  }, []);

  return (
    <div className="landing-page">
      <div className="landing-card">
        <div className="landing-eyebrow">Travel Map</div>
        <h1>Share the places you have visited.</h1>
        <p>
          A personal travel map with visited cities, landmarks and travel statistics.
        </p>
        <p>
          Created by Valery Kirkizh: <a href="mailto:valery@kirkizh.com">Email</a> &bull;&nbsp;<a href="https://www.linkedin.com/in/vkirkizh/" rel="me">LinkedIn</a> &bull;&nbsp;<a href="https://github.com/vkirkizh" rel="me">GitHub</a>
        </p>
        <div className="landing-actions">
          <a href="/valery/" className="landing-link">
            View demo map
          </a>
          <a href="/login" className="landing-link landing-link-secondary">
            Log In / Sign Up
          </a>
        </div>
      </div>
    </div>
  );
}

function PublicMapPage({ username }: { username: string }) {
  const [data, setData] = useState<MapResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setPageTitle(`Travel Map — @${username}`);

    fetch(`${apiBaseUrl}/api/public/users/${username}/map`)
      .then((response) => {
        if (!response.ok) {
          throw new Error("Failed to load map data");
        }

        return response.json();
      })
      .then(setData)
      .catch((err: unknown) => {
        setError(err instanceof Error ? err.message : "Unknown error");
      });
  }, [username]);

  if (error) {
    setPageTitle('Travel Map — Error');

    return <div className="error">Failed to load Travel Map: {error}</div>;
  }

  if (!data) {
    return <div className="loading">Loading Travel Map...</div>;
  }

  return (
    <div className="page">
      <MapContainer
        center={[50.5, 10.5]}
        zoom={5}
        minZoom={2}
        maxZoom={18}
        scrollWheelZoom
        className="map"
        attributionControl={false}
      >
        <AttributionControl
          prefix='<a href="https://github.com/vkirkizh/travel-map" target="_blank">Travel Map</a> | <a href="https://leafletjs.com" target="_blank" rel="nofollow">Leaflet</a>'
        />

        <TileLayer
          attribution='<a href="https://www.openstreetmap.org/copyright" target="_blank" rel="nofollow">OpenStreetMap</a>'
          url="https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png"
        />

        {data.places.map((place) => (
          <CircleMarker
            key={place.id}
            center={[place.lat, place.lng]}
            radius={8}
            pathOptions={{
              opacity: 1,
              fillOpacity: 0.85,
              weight: 2,
            }}
          >
            <Popup>{place.title}</Popup>
          </CircleMarker>
        ))}

        <FitMapBounds data={data} />
      </MapContainer>

      <div className="profile-card">
        <div className="avatar">
          {data.user.avatar_url ? (
            <img src={data.user.avatar_url} alt={data.user.display_name} />
          ) : (
            data.user.display_name.charAt(0)
          )}
        </div>
        <div>
          <div className="display-name">{data.user.display_name}</div>
          <div className="username">@{data.user.username}</div>
        </div>
      </div>

      <div className="stats-card">
        <div>
          <strong>{data.stats.countries_visited}</strong>
          <span>countries</span>
        </div>
        <div>
          <strong>{data.stats.places_visited}</strong>
          <span>places</span>
        </div>
      </div>
    </div>
  );
}

function PrivateAppPage({ pathname }: { pathname: string }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    setPageTitle(
      pathname.startsWith("/app/settings")
        ? "Travel Map — Settings"
        : "Travel Map — Dashboard",
    );

    fetch(`${apiBaseUrl}/api/me`, {
      credentials: "include",
    })
      .then(async (response) => {
        if (response.status === 401) {
          setUser(null);
          return;
        }

        if (!response.ok) {
          throw new Error("Failed to load current user");
        }

        const payload = await response.json();
        setUser(payload.user);
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, [pathname]);

  if (isLoading) {
    return <div className="loading">Loading Travel Map App...</div>;
  }

  if (!user) {
    return (
      <AuthScreen apiBaseUrl={apiBaseUrl} onAuthenticated={(user) => setUser(user)} />
    );
  }

  if (pathname.startsWith("/app/settings")) {
    return (
      <ProfileSettings
        apiBaseUrl={apiBaseUrl}
        user={user}
        onUserUpdated={(user) => setUser(user)}
      />
    );
  }

  return (
    <Dashboard
      apiBaseUrl={apiBaseUrl}
      user={user}
      onLogout={() => setUser(null)}
    />
  );
}

function FitMapBounds({ data }: { data: MapResponse }) {
  const map = useMap();

  useEffect(() => {
    const points: [number, number][] = [];

    data.places.forEach((place) => {
      points.push([place.lat, place.lng]);
    });

    if (points.length === 0) {
      return;
    }

    if (points.length === 1) {
      map.setView(points[0], 8);
      return;
    }

    const bounds = L.latLngBounds(points);
    map.fitBounds(bounds, {
      padding: [80, 80],
      maxZoom: 8,
    });
  }, [data, map]);

  return null;
}

function setPageTitle(title: string) {
  document.title = title;
}

export default App;
