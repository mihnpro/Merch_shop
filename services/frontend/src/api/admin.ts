import { request, uploadFile } from "./client";
import type {
  BalanceResponse,
  Category,
  CreateCategoryInput,
  GrantPointsBody,
  InventoryAdjustInput,
  InventoryAdjustResult,
  ListStockResponse,
  ListUsersParams,
  ListUsersResponse,
  Product,
  ProductInput,
  UpdateCategoryInput,
  UpdateProductInput,
  UploadPhotoResponse,
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
