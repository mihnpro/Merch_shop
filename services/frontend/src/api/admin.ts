import { request, uploadFile } from "./client";
import type {
  Category,
  CreateCategoryInput,
  InventoryAdjustInput,
  InventoryAdjustResult,
  ListStockResponse,
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
