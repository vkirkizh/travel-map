import {useEffect, useState} from "react";

type Place = {
  id: string;
  title: string;
  query: string;
  country_code: string;
  lat: number;
  lng: number;
};

type Props = {
  apiBaseUrl: string;
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

export function PlacesManager({ apiBaseUrl }: Props) {
  const [places, setPlaces] = useState<Place[]>([]);
  const [query, setQuery] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function loadPlaces() {
    setError(null);

    const response = await fetch(`${apiBaseUrl}/api/places`, {
      credentials: "include",
    });

    const payload = await response.json();

    if (!response.ok) {
      setError(getPlaceErrorMessage(payload.error));
      return;
    }

    setPlaces(payload.places);
  }

  useEffect(() => {
    loadPlaces()
      .catch(() => setError("Unable to load places."))
      .finally(() => setIsLoading(false));
  }, []);

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
      <div className="section-header">
        <div>
          <h2>Visited places</h2>
          <p>Add cities or landmarks to your public travel map.</p>
        </div>
      </div>

      <form className="place-form" onSubmit={addPlace}>
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

      {error && <div className="form-error">{error}</div>}

      {isLoading ? (
        <div className="empty-state">Loading places...</div>
      ) : places.length === 0 ? (
        <div className="empty-state">No places yet.</div>
      ) : (
        <div className="places-list">
          {places.map((place) => (
            <div className="place-row" key={place.id}>
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
