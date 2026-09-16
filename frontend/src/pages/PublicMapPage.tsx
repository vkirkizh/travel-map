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
            radius={7}
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
