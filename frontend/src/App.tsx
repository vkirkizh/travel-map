import {useEffect, useState} from "react";
import {MapContainer, Marker, Polyline, Popup, TileLayer} from "react-leaflet";
import "leaflet/dist/leaflet.css";
import "./App.css";

type User = {
  username: string;
  display_name: string;
  avatar_url: string | null;
};

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

  if (pathname === "/app" || pathname === "/app/") {
    return <AppPlaceholder />;
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

function AppPlaceholder() {
  useEffect(() => {
    setPageTitle("Travel Map — Dashboard");
  }, []);

  return (
    <div className="landing-page">
      <div className="landing-card">
        <div className="landing-eyebrow">Travel Map App</div>
        <h1>Private dashboard is coming soon.</h1>
        <p>
          This page will contain profile settings, places management and flight
          management.
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
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />

        {data.places.map((place) => (
          <Marker key={place.id} position={[place.lat, place.lng]}>
            <Popup>{place.title}</Popup>
          </Marker>
        ))}

        {data.flights.map((flight) => (
          <Polyline
            key={flight.id}
            positions={[
              [flight.from_point.lat, flight.from_point.lng],
              [flight.to_point.lat, flight.to_point.lng],
            ]}
          />
        ))}
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
      </div>
    </div>
  );
}

function setPageTitle(title: string) {
  document.title = title;
}

export default App;
