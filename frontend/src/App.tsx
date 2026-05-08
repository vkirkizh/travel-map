import {useEffect, useState} from "react";
import {CircleMarker, MapContainer, Polyline, Popup, TileLayer, useMap} from "react-leaflet";
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

type Flight = {
  id: string;
  from: string;
  to: string;
  from_point: {
    lat: number;
    lng: number;
  };
  to_point: {
    lat: number;
    lng: number;
  };
};

type Stats = {
  countries_visited: number;
  places_visited: number;
  flights_taken: number;
  flight_distance_km: number;
  flight_hours: number;
};

type MapResponse = {
  user: User;
  places: Place[];
  flights: Flight[];
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
          A personal travel map with visited cities, landmarks, flights and
          travel statistics.
        </p>
        <p>
          Created by Valery Kirkizh: <a href="mailto:valery@kirkizh.com">Email</a> &bull;&nbsp;<a href="https://www.linkedin.com/in/vkirkizh/" rel="me">LinkedIn</a> &bull;&nbsp;<a href="https://github.com/vkirkizh" rel="me">GitHub</a>
        </p>
        <a href="/vkirkizh/" className="landing-link">
          View demo map
        </a>
      </div>
    </div>
  );
}

function PublicMapPage({ username }: { username: string }) {
  const [data, setData] = useState<MapResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setData(null);
    setError(null);

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
        scrollWheelZoom
        className="map"
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
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

        {data.flights.map((flight) => (
          <Polyline
            key={flight.id}
            positions={[
              [flight.from_point.lat, flight.from_point.lng],
              [flight.to_point.lat, flight.to_point.lng],
            ]}
            pathOptions={{
              opacity: 0.4,
              weight: 2,
            }}
          />
        ))}

        {data.flights.map((flight) => (
          <CircleMarker
            key={`${flight.id}-from`}
            center={[flight.from_point.lat, flight.from_point.lng]}
            radius={4}
            pathOptions={{
              opacity: 0.75,
              fillOpacity: 0.9,
              weight: 1,
            }}
          >
            <Popup>{flight.from}</Popup>
          </CircleMarker>
        ))}

        {data.flights.map((flight) => (
          <CircleMarker
            key={`${flight.id}-to`}
            center={[flight.to_point.lat, flight.to_point.lng]}
            radius={4}
            pathOptions={{
              opacity: 0.75,
              fillOpacity: 0.9,
              weight: 1,
            }}
          >
            <Popup>{flight.to}</Popup>
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
        <div>
          <strong>{data.stats.flights_taken}</strong>
          <span>flights</span>
        </div>
        <div>
          <strong>{data.stats.flight_distance_km}</strong>
          <span>km flown</span>
        </div>
        <div>
          <strong>{data.stats.flight_hours}</strong>
          <span>hours flown</span>
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
  }, []);

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

    data.flights.forEach((flight) => {
      points.push([flight.from_point.lat, flight.from_point.lng]);
      points.push([flight.to_point.lat, flight.to_point.lng]);
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
