import {useEffect, useState} from "react";

type Flight = {
  id: string;
  from_airport_iata: string;
  to_airport_iata: string;
  departure_time: string | null;
  arrival_time: string | null;
  flight_number: string | null;
  distance_km: number;
};

type Props = {
  apiBaseUrl: string;
};

function getFlightErrorMessage(error: string): string {
  switch (error) {
    case "airport not found":
      return "Airport was not found. Try valid IATA codes like BER, ZAG, VIE, LHR.";

    case "validation failed":
      return "Please check flight details.";

    default:
      return "Something went wrong. Please try again.";
  }
}

export function FlightsManager({ apiBaseUrl }: Props) {
  const [flights, setFlights] = useState<Flight[]>([]);
  const [fromAirport, setFromAirport] = useState("");
  const [toAirport, setToAirport] = useState("");
  const [flightNumber, setFlightNumber] = useState("");
  const [departureTime, setDepartureTime] = useState("");
  const [arrivalTime, setArrivalTime] = useState("");
  const [isLoading, setIsLoading] = useState(true)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function loadFlights() {
    const response = await fetch(`${apiBaseUrl}/api/flights`, {
      credentials: "include",
    })

    const payload = await response.json()

    if (!response.ok) {
      setError(getFlightErrorMessage(payload.error))
      return
    }

    setFlights(payload.flights)
  }

  useEffect(() => {
    loadFlights()
      .catch(() => setError("Unable to load flights."))
      .finally(() => setIsLoading(false))
  }, [])

  async function addFlight(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()

    setError(null)
    setIsSubmitting(true)

    try {
      const departureTimeISO = departureTime
        ? new Date(`${departureTime}`).toISOString()
        : null
      const arrivalTimeISO = arrivalTime
        ? new Date(`${arrivalTime}`).toISOString()
        : null

      const response = await fetch(`${apiBaseUrl}/api/flights`, {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          from_airport_iata: fromAirport,
          to_airport_iata: toAirport,
          departure_time: departureTimeISO,
          arrival_time: arrivalTimeISO,
          flight_number: flightNumber || null,
        }),
      })

      const payload = await response.json()

      if (!response.ok) {
        setError(getFlightErrorMessage(payload.error))
        return
      }

      setFlights((current) => [...current, payload.flight])

      setFromAirport("")
      setToAirport("")
      setFlightNumber("")
      setDepartureTime("")
      setArrivalTime("")
    } catch {
      setError("Unable to create flight.")
    } finally {
      setIsSubmitting(false)
    }
  }

  async function deleteFlight(id: string) {
    const previousFlights = flights
    setFlights((current) => current.filter((flight) => flight.id !== id))

    try {
      const response = await fetch(`${apiBaseUrl}/api/flights/${id}`, {
        method: "DELETE",
        credentials: "include",
      })

      if (!response.ok) {
        setFlights(previousFlights)
        setError("Unable to delete flight.")
      }
    } catch {
      setFlights(previousFlights)
      setError("Unable to delete flight.")
    }
  }

  return (
    <section className="flights-section">
      <div className="section-header">
        <h2>Flights</h2>
        <p>Add your flights to visualize routes on your public map.</p>
      </div>

      <form className="flight-form" onSubmit={addFlight}>
        <input
          value={fromAirport}
          onChange={(e) => setFromAirport(e.target.value.toUpperCase())}
          placeholder="From"
          maxLength={3}
          disabled={isSubmitting}
        />

        <input
          value={toAirport}
          onChange={(e) => setToAirport(e.target.value.toUpperCase())}
          placeholder="To"
          maxLength={3}
          disabled={isSubmitting}
        />

        <input
          value={flightNumber}
          onChange={(e) => setFlightNumber(e.target.value)}
          placeholder="Flight number"
          disabled={isSubmitting}
        />

        <input
          type="datetime-local"
          value={departureTime}
          onChange={(e) => setDepartureTime(e.target.value)}
          disabled={isSubmitting}
        />

        <input
          type="datetime-local"
          value={arrivalTime}
          onChange={(e) => setArrivalTime(e.target.value)}
          disabled={isSubmitting}
        />

        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Adding..." : "Add flight"}
        </button>
      </form>

      {error && <div className="form-error">{error}</div>}

      {isLoading ? (
        <div className="empty-state">Loading flights...</div>
      ) : flights.length === 0 ? (
        <div className="empty-state">Add your first flight to see routes on your public map.</div>
      ) : (
        <div className="flights-list">
          {flights.map((flight) => (
            <div className="flight-row" key={flight.id}>
              <div>
                <strong>
                  {flight.from_airport_iata} → {flight.to_airport_iata}
                </strong>
                <span>
                  {formatFlightDate(flight.departure_time)} ·{" "}
                  {formatFlightNumber(flight.flight_number)} ·{" "}
                  {formatDuration(flight.departure_time, flight.arrival_time)} ·{" "}
                  {flight.distance_km} km
                </span>
              </div>

              <button
                type="button"
                onClick={() => void deleteFlight(flight.id)}
              >
                Delete
              </button>
            </div>
          ))}
        </div>
      )}
    </section>
  )
}

function formatFlightDate(value: string | null): string {
  if (!value) {
    return "N/A";
  }

  return new Intl.DateTimeFormat("de-DE", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(new Date(value));
}

function formatDuration(departure: string | null, arrival: string | null): string {
  if (!departure || !arrival) {
    return "N/A";
  }

  const durationMs = new Date(arrival).getTime() - new Date(departure).getTime();
  if (durationMs <= 0) {
    return "N/A";
  }

  const totalMinutes = Math.round(durationMs / 1000 / 60);
  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;
  if (hours === 0) {
    return `${minutes}m`;
  }
  if (minutes === 0) {
    return `${hours}h`;
  }

  return `${hours}h ${minutes}m`;
}

function formatFlightNumber(value: string | null): string {
  return value && value.trim() !== "" ? value : "N/A";
}
