import { request } from "./client";
import type { CartItem, CartView } from "./types";

export function getCart(): Promise<CartView> {
  return request<CartView>("/cart");
}

export function addItem(product_id: string, quantity: number): Promise<CartItem> {
  return request<CartItem>("/cart/items", {
    method: "POST",
    body: JSON.stringify({ product_id, quantity }),
  });
}

export function updateItem(id: string, quantity: number): Promise<CartItem | void> {
  return request<CartItem | void>(`/cart/items/${id}`, {
    method: "PATCH",
    body: JSON.stringify({ quantity }),
  });
}

export function removeItem(id: string): Promise<void> {
  return request<void>(`/cart/items/${id}`, { method: "DELETE" });
}

export function clearCart(): Promise<void> {
  return request<void>("/cart", { method: "DELETE" });
}
