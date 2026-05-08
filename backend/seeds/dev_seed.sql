DELETE FROM flights
WHERE user_id = (
  SELECT id FROM users WHERE username = 'vkirkizh'
);

DELETE FROM places
WHERE user_id = (
  SELECT id FROM users WHERE username = 'vkirkizh'
);

INSERT INTO countries (code, name)
VALUES
  ('DE', 'Germany'),
  ('HR', 'Croatia'),
  ('GB', 'United Kingdom')
  ON CONFLICT (code) DO UPDATE SET
  name = EXCLUDED.name;

INSERT INTO users (
  username,
  email,
  password_hash,
  display_name,
  avatar_url
)
VALUES (
    'vkirkizh',
    'valery@kirkizh.com',
    'dev-password-hash',
    'Valery Kirkizh',
    NULL
  )
  ON CONFLICT (username) DO UPDATE SET
  display_name = EXCLUDED.display_name,
  avatar_url = EXCLUDED.avatar_url,
  updated_at = now();

INSERT INTO airports (
  iata_code,
  name,
  city,
  country_code,
  lat,
  lng
)
VALUES
  ('BER', 'Berlin Brandenburg Airport', 'Berlin', 'DE', 52.3667, 13.5033),
  ('ZAG', 'Zagreb Airport', 'Zagreb', 'HR', 45.7429, 16.0688)
  ON CONFLICT (iata_code) DO UPDATE SET
  icao_code = EXCLUDED.icao_code,
  name = EXCLUDED.name,
  city = EXCLUDED.city,
  country_code = EXCLUDED.country_code,
  lat = EXCLUDED.lat,
  lng = EXCLUDED.lng;

INSERT INTO places (
  user_id,
  country_code,
  title,
  query,
  lat,
  lng
)
SELECT
  u.id,
  p.country_code,
  p.title,
  p.query,
  p.lat,
  p.lng
FROM users u
CROSS JOIN (
VALUES
  ('DE', 'Berlin, Germany', 'berlin germany', 52.5200, 13.4050),
  ('HR', 'Zagreb, Croatia', 'zagreb croatia', 45.8150, 15.9819),
  ('GB', 'Stonehenge, United Kingdom', 'stonehenge united kingdom', 51.1789, -1.8262)
) AS p(country_code, title, query, lat, lng)
WHERE u.username = 'vkirkizh'
  ON CONFLICT DO NOTHING;

INSERT INTO flights (
  user_id,
  from_airport_iata,
  to_airport_iata,
  departure_time,
  arrival_time,
  flight_number,
  distance_km
)
SELECT
  u.id,
  'BER',
  'ZAG',
  '2024-05-01 10:00:00+00',
  '2024-05-01 11:30:00+00',
  'OU4401',
  770
FROM users u
WHERE u.username = 'vkirkizh'
  ON CONFLICT DO NOTHING;
