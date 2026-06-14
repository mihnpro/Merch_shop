import { request, uploadFile } from "./client";
import type {
  AdminListOrdersParams,
  AdminUser,
  AnalyticsPeriod,
  AnalyticsView,
  BalanceResponse,
  Category,
  CreateCategoryInput,
  GrantPointsBody,
  InventoryAdjustInput,
  InventoryAdjustResult,
  ListOrdersResponse,
  ListStockResponse,
  ListUsersParams,
  ListUsersResponse,
  OrderView,
  Product,
  ProductInput,
  ResetPasswordResponse,
  TransactionsParams,
  TransactionsResponse,
  UpdateCategoryInput,
  UpdateOrderStatusBody,
  UpdateProductInput,
  UploadPhotoResponse,
  UsersStats,
} from "./types";

export function uploadPhoto(file: File): Promise<UploadPhotoResponse> {
  return uploadFile<UploadPhotoResponse>("/admin/media/photos", file);
}

export function createProduct(body: ProductInput): Promise<Product> {
  return request<Product>("/admin/products", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function updateProduct(id: string, body: UpdateProductInput): Promise<Product> {
  return request<Product>(`/admin/products/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function deactivateProduct(id: string): Promise<void> {
  return request<void>(`/admin/products/${id}`, { method: "DELETE" });
}

export function createCategory(body: CreateCategoryInput): Promise<Category> {
  return request<Category>("/admin/categories", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function updateCategory(id: string, body: UpdateCategoryInput): Promise<Category> {
  return request<Category>(`/admin/categories/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function listInventory(): Promise<ListStockResponse> {
  return request<ListStockResponse>("/admin/inventory", { method: "GET" });
}

export function adjustInventory(body: InventoryAdjustInput): Promise<InventoryAdjustResult> {
  return request<InventoryAdjustResult>("/admin/inventory/adjust", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function listUsers(params: ListUsersParams = {}): Promise<ListUsersResponse> {
  const query = new URLSearchParams();
  if (params.search) query.set("search", params.search);
  if (params.role) query.set("role", params.role);
  if (params.status) query.set("status", params.status);
  if (params.page_size) query.set("page_size", String(params.page_size));
  if (params.page_token) query.set("page_token", params.page_token);
  const qs = query.toString();
  return request<ListUsersResponse>(`/admin/users${qs ? `?${qs}` : ""}`, { method: "GET" });
}

export function grantPoints(userId: string, body: GrantPointsBody): Promise<BalanceResponse> {
  return request<BalanceResponse>(`/admin/users/${userId}/grant-points`, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function getUser(userId: string): Promise<AdminUser> {
  return request<AdminUser>(`/admin/users/${userId}`, { method: "GET" });
}

export function resetUserPassword(userId: string): Promise<ResetPasswordResponse> {
  return request<ResetPasswordResponse>(`/admin/users/${userId}/reset-password`, {
    method: "POST",
  });
}

export function blockUser(
  userId: string,
  blocked: boolean,
  reason?: string,
): Promise<AdminUser> {
  return request<AdminUser>(`/admin/users/${userId}/status`, {
    method: "PUT",
    body: JSON.stringify({ blocked, reason }),
  });
}

export function changeUserRole(userId: string, role: "user" | "admin"): Promise<AdminUser> {
  return request<AdminUser>(`/admin/users/${userId}/role`, {
    method: "PUT",
    body: JSON.stringify({ role }),
  });
}

export function getUserTransactions(
  userId: string,
  params: TransactionsParams = {},
): Promise<TransactionsResponse> {
  const query = new URLSearchParams();
  if (params.page_size) query.set("page_size", String(params.page_size));
  if (params.page_token) query.set("page_token", params.page_token);
  const qs = query.toString();
  return request<TransactionsResponse>(
    `/admin/users/${userId}/transactions${qs ? `?${qs}` : ""}`,
    { method: "GET" },
  );
}

export function getUsersStats(period: AnalyticsPeriod): Promise<UsersStats> {
  return request<UsersStats>(`/admin/users/stats?period=${period}`, { method: "GET" });
}

export function adminListOrders(
  params: AdminListOrdersParams = {},
): Promise<ListOrdersResponse> {
  const query = new URLSearchParams();
  if (params.user_id) query.set("user_id", params.user_id);
  if (params.status) query.set("status", params.status);
  if (params.page_size) query.set("page_size", String(params.page_size));
  if (params.page_token) query.set("page_token", params.page_token);
  const qs = query.toString();
  return request<ListOrdersResponse>(`/admin/orders${qs ? `?${qs}` : ""}`, {
    method: "GET",
  });
}

export function adminUpdateOrderStatus(
  orderId: string,
  body: UpdateOrderStatusBody,
): Promise<OrderView> {
  return request<OrderView>(`/admin/orders/${orderId}/status`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function getAnalytics(period: AnalyticsPeriod): Promise<AnalyticsView> {
  return request<AnalyticsView>(`/admin/analytics?period=${period}`, { method: "GET" });
}
