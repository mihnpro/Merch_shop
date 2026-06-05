import { request } from "./client";
import type {
  Category,
  CreateCategoryInput,
  Product,
  ProductInput,
  UpdateCategoryInput,
  UpdateProductInput,
} from "./types";

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
