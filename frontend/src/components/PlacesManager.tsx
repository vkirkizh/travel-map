import {useEffect, useState} from "react";
import type {PrivatePlace} from "../types";
import "./PlacesManager.css";

type Props = {
  apiBaseUrl: string;
  onUnauthorized: () => void;
};

function getPlaceErrorMessage(error: string): string {
  switch (error) {
    case "place not found":
      return "Place was not found. Please try again.";
    case "validation failed":
      return "Please enter a place.";
    case "unauthorized":
      return "Please login again.";
    default:
      return "Something went wrong. Please try again.";
  }
}

export function PlacesManager({apiBaseUrl, onUnauthorized}: Props) {
  const [places, setPlaces] = useState<PrivatePlace[]>([]);
  const [query, setQuery] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadPlaces() {
      const response = await fetch(`${apiBaseUrl}/api/places`, {
        credentials: "include",
      });

      if (response.status === 401) {
        onUnauthorized();
        return;
      }

      const payload = await response.json();

      if (!response.ok) {
        setError(getPlaceErrorMessage(payload.error));
        return;
      }

      setPlaces(payload.places);
    }

    loadPlaces()
      .catch(() => setError("Unable to load places."))
      .finally(() => setIsLoading(false));
  }, [apiBaseUrl, onUnauthorized]);

  async function addPlace(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    setError(null);
    setIsSubmitting(true);

    if (query.trim() === "") {
      setError("Please enter a place.");
      setIsSubmitting(false);
      return;
    }

    try {
      const response = await fetch(`${apiBaseUrl}/api/places`, {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ query }),
      });

      if (response.status === 401) {
        onUnauthorized();
        return;
      }

      const payload = await response.json();

      if (!response.ok) {
        setError(getPlaceErrorMessage(payload.error));
        return;
      }

      setPlaces((current) => [...current, payload.place]);
      setQuery("");
    } catch {
      setError("Unable to add place.");
    } finally {
      setIsSubmitting(false);
    }
  }

  async function deletePlace(id: string) {
    setError(null);

    const previousPlaces = places;
    setPlaces((current) => current.filter((place) => place.id !== id));

    try {
      const response = await fetch(`${apiBaseUrl}/api/places/${id}`, {
        method: "DELETE",
        credentials: "include",
      });

      if (response.status === 401) {
        onUnauthorized();
        return;
      }

      if (!response.ok) {
        const payload = await response.json();
        setError(getPlaceErrorMessage(payload.error));
        setPlaces(previousPlaces);
      }
    } catch {
      setError("Unable to delete place.");
      setPlaces(previousPlaces);
    }
  }

  return (
    <section className="places-section">
      <div className="places-header">
        <div>
          <h2>Visited places</h2>
          <p>Add cities or landmarks to your public travel map.</p>
        </div>
      </div>

      <form className="places-form" onSubmit={addPlace}>
        <input
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Berlin, Germany"
          disabled={isSubmitting}
        />
        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Adding..." : "Add place"}
        </button>
      </form>

      {error && <div className="places-error">{error}</div>}

      {isLoading ? (
        <div className="places-empty">Loading places...</div>
      ) : places.length === 0 ? (
        <div className="places-empty">Add your first visited place to see it on your public map.</div>
      ) : (
        <div className="places-list">
          {places.map((place) => (
            <div className="places-row" key={place.id}>
              <div>
                <strong>{place.title}</strong>
                <span>
                  {place.country_code} · {place.lat.toFixed(4)},{" "}
                  {place.lng.toFixed(4)}
                </span>
              </div>

              <button type="button" onClick={() => void deletePlace(place.id)}>
                Delete
              </button>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
