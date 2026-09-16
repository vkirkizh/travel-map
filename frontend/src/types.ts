export type User = {
  id: string;
  username: string;
  email: string;
  display_name: string;
  avatar_url: string;
};

export type PrivatePlace = {
  id: string;
  title: string;
  query: string;
  country_code: string;
  lat: number;
  lng: number;
};

export type PublicPlace = {
  id: string;
  title: string;
  country_code: string;
  lat: number;
  lng: number;
};

type PublicUser = {
  username: string;
  display_name: string;
  avatar_url: string;
};

type MapStats = {
  countries_visited: number;
  places_visited: number;
};

export type PublicMapResponse = {
  user: PublicUser;
  places: PublicPlace[];
  stats: MapStats;
};
