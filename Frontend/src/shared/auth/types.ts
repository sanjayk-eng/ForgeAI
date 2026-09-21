export type AuthTokens = {
  access_token: string;
  refresh_token: string;
};

export type UserProfile = {
  id: string;
  email: string;
  name: string;
};

export type Credentials = {
  email: string;
  password: string;
};

export type Registration = Credentials & {
  name: string;
};
