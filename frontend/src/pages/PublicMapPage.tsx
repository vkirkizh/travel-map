import {useEffect, useState} from "react";
import L from "leaflet";
import {
  AttributionControl,
  CircleMarker,
  MapContainer,
  Popup,
  TileLayer,
  useMap,
} from "react-leaflet";
import type {PublicMapResponse} from "../types";
import "./PublicMapPage.css";

const cartoApiKey = import.meta.env.VITE_CARTO_API_KEY?.trim() ?? "";

type Props = {
  apiBaseUrl: string;
  username: string;
};

export function PublicMapPage({apiBaseUrl, username}: Props) {
  const [data, setData] = useState<PublicMapResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    document.title = error
      ? "Travel Map — Error"
      : `Travel Map — @${username}`;
  }, [error, username]);

  useEffect(() => {
    fetch(`${apiBaseUrl}/api/public/users/${encodeURIComponent(username)}/map`)
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
  }, [apiBaseUrl, username]);

  if (error) {
    return (
      <div className="public-map-status public-map-status-error">
        Failed to load Travel Map: {error}
      </div>
    );
  }

  if (!data) {
    return <div className="public-map-status">Loading Travel Map...</div>;
  }

  return (
    <div className="public-map-page">
      <MapContainer
        center={[50.5, 10.5]}
        zoom={5}
        minZoom={2}
        maxZoom={18}
        scrollWheelZoom
        className="public-map-canvas"
        attributionControl={false}
      >
        <AttributionControl
          prefix='<a href="https://map.kirkizh.com/">Travel Map</a> | <a href="https://github.com/vkirkizh/travel-map" target="_blank" rel="noopener">GitHub Project</a>'
        />

        <TileLayer
          attribution='<a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noopener nofollow">OpenStreetMap</a> | <a href="https://carto.com/attributions" target="_blank" rel="noopener nofollow">CARTO</a>'
          url={`https://basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png?key=${encodeURIComponent(cartoApiKey)}`}
        />

        {data.places.map((place) => (
          <CircleMarker
            key={place.id}
            center={[place.lat, place.lng]}
            radius={6}
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

      <div className="public-profile-card">
        <div className="public-profile-avatar">
          {data.user.avatar_url ? (
            <img src={data.user.avatar_url} alt={data.user.display_name} />
          ) : (
            data.user.display_name.charAt(0)
          )}
        </div>
        <div>
          <div className="public-profile-display-name">{data.user.display_name}</div>
          <div className="public-profile-username">@{data.user.username}</div>
        </div>
      </div>

      <div className="public-stats-card">
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

function FitMapBounds({data}: {data: PublicMapResponse}) {
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
