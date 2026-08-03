export interface UserResponse {
  id: string;
  username: string;
  role: "admin" | "user";
  created_at: string;
}

export interface LoginResponse {
  access_token: string;
  expires_in: number;
  user: UserResponse;
}

export interface RefreshResponse {
  access_token: string;
  expires_in: number;
}