import { request } from "./client";
import type {
  BalanceResponse,
  ChangePasswordBody,
  LoginBody,
  LoginResponse,
  MeResponse,
  RegisterBody,
  TransactionsParams,
  TransactionsResponse,
  UpdateProfileBody,
  User,
  UserProfile,
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

export function logout(): Promise<void> {
  return request<void>("/auth/logout", { method: "POST" });
}

export function getMe(): Promise<MeResponse> {
  return request<MeResponse>("/me", { method: "GET" });
}

export function getMyBalance(): Promise<BalanceResponse> {
  return request<BalanceResponse>("/me/balance", { method: "GET" });
}

export function getMyProfile(): Promise<UserProfile> {
  return request<UserProfile>("/me/profile", { method: "GET" });
}

export function updateMyProfile(body: UpdateProfileBody): Promise<UserProfile> {
  return request<UserProfile>("/me/profile", {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function changeMyPassword(body: ChangePasswordBody): Promise<void> {
  return request<void>("/me/password", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function getMyTransactions(
  params: TransactionsParams = {},
): Promise<TransactionsResponse> {
  const query = new URLSearchParams();
  if (params.page_size) query.set("page_size", String(params.page_size));
  if (params.page_token) query.set("page_token", params.page_token);
  const qs = query.toString();
  return request<TransactionsResponse>(`/me/transactions${qs ? `?${qs}` : ""}`, {
    method: "GET",
  });
}