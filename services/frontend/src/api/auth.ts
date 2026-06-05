import { request } from "./client";
import type {
  LoginBody,
  LoginResponse,
  MeResponse,
  RegisterBody,
  User,
} from "./types";

export function register(body: RegisterBody): Promise<User> {
  return request<User>("/auth/register", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function login(body: LoginBody): Promise<LoginResponse> {
  return request<LoginResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function logout(refreshToken: string): Promise<void> {
  return request<void>("/auth/logout", {
    method: "POST",
    body: JSON.stringify({ refresh_token: refreshToken }),
  });
}

export function getMe(): Promise<MeResponse> {
  return request<MeResponse>("/me", { method: "GET" });
}