CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username TEXT NOT NULL UNIQUE,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  display_name TEXT NOT NULL,
  avatar_url TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX sessions_user_id_idx ON sessions(user_id);
CREATE INDEX sessions_expires_at_idx ON sessions(expires_at);

CREATE TABLE countries (
  code CHAR(2) PRIMARY KEY,
  name TEXT NOT NULL UNIQUE
);

CREATE TABLE places (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  country_code CHAR(2) NOT NULL REFERENCES countries(code),
  title TEXT NOT NULL,
  query TEXT NOT NULL,
  lat DOUBLE PRECISION NOT NULL,
  lng DOUBLE PRECISION NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX places_user_id_idx ON places(user_id);
CREATE INDEX places_country_code_idx ON places(country_code);

CREATE TABLE airports (
  iata_code CHAR(3) PRIMARY KEY,
  name TEXT NOT NULL,
  city TEXT,
  country_code CHAR(2) REFERENCES countries(code),
  lat DOUBLE PRECISION NOT NULL,
  lng DOUBLE PRECISION NOT NULL
);

CREATE INDEX airports_country_code_idx ON airports(country_code);

CREATE TABLE flights (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  from_airport_iata CHAR(3) NOT NULL REFERENCES airports(iata_code),
  to_airport_iata CHAR(3) NOT NULL REFERENCES airports(iata_code),
  departure_time TIMESTAMPTZ,
  arrival_time TIMESTAMPTZ,
  flight_number TEXT,
  distance_km INTEGER,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX flights_user_id_idx ON flights(user_id);
CREATE INDEX flights_from_airport_iata_idx ON flights(from_airport_iata);
CREATE INDEX flights_to_airport_iata_idx ON flights(to_airport_iata);

CREATE TABLE geocoding_cache (
  query_normalized TEXT PRIMARY KEY,
  provider TEXT NOT NULL,
  result_json JSONB NOT NULL,
  country_code CHAR(2) REFERENCES countries(code),
  lat DOUBLE PRECISION NOT NULL,
  lng DOUBLE PRECISION NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX geocoding_cache_country_code_idx ON geocoding_cache(country_code);
