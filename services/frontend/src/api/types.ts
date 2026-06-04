export interface User {
  id: string;
  login: string;
  first_name: string;
  last_name: string;
  patronymic?: string;
  email: string;
  phone_number?: string;
}

export interface Tokens {
  access_token: string;
  refresh_token: string;
}

export interface RegisterBody {
  login: string;
  password: string;
  first_name: string;
  last_name: string;
  email: string;
  patronymic?: string;
  phone_number?: string;
}

export interface LoginBody {
  login: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  tokens: Tokens;
}

export interface MeResponse {
  user_id: string;
  role: string;
}
