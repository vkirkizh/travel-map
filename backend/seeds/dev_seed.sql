DELETE FROM places
WHERE user_id = (
  SELECT id FROM users WHERE username = 'valery'
);

INSERT INTO countries (code, name)
VALUES
  ('DE', 'Germany'),
  ('HR', 'Croatia'),
  ('GB', 'United Kingdom'),
  ('AT', 'Austria'),
  ('CZ', 'Czech Republic'),
  ('FR', 'France'),
  ('IT', 'Italy')
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
    'valery',
    'valery@kirkizh.com',
    '$2a$10$uk9CdFZbGiIm4RpQdXoHHeEjU5JiKC9tuP5TNuZs5RPYwLVuNHQsq',
    'Valery Kirkizh',
    NULL
  )
  ON CONFLICT (username) DO UPDATE SET
  display_name = EXCLUDED.display_name,
  avatar_url = EXCLUDED.avatar_url,
  updated_at = now();

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
WHERE u.username = 'valery'
  ON CONFLICT DO NOTHING;
