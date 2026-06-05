import { request } from "./client";
import type {
  CategoriesResponse,
  ListProductsParams,
  ListProductsResponse,
  Product,
} from "./types";

function buildQuery(params: Record<string, string | number | boolean | undefined>): string {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== "") {
      search.set(key, String(value));
    }
  }
  const query = search.toString();
  return query ? `?${query}` : "";
}

export function listCategories(activeOnly = true): Promise<CategoriesResponse> {
  return request<CategoriesResponse>(`/categories${buildQuery({ active_only: activeOnly })}`, {
    method: "GET",
  });
}

export function listProducts(params: ListProductsParams = {}): Promise<ListProductsResponse> {
  return request<ListProductsResponse>(`/products${buildQuery({ ...params })}`, {
    method: "GET",
  });
}

export function getProduct(id: string): Promise<Product> {
  return request<Product>(`/products/${id}`, { method: "GET" });
}
