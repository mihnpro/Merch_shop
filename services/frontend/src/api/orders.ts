import { request } from "./client";
import type { ListOrdersParams, ListOrdersResponse, OrderView } from "./types";

export function createOrder(delivery_address: string, note?: string): Promise<OrderView> {
  return request<OrderView>("/orders", {
    method: "POST",
    body: JSON.stringify({ delivery_address, note }),
  });
}

export function getOrder(id: string): Promise<OrderView> {
  return request<OrderView>(`/orders/${id}`);
}

export function listMyOrders(params: ListOrdersParams = {}): Promise<ListOrdersResponse> {
  const query = new URLSearchParams();
  if (params.status) query.set("status", params.status);
  if (params.page_size) query.set("page_size", String(params.page_size));
  if (params.page_token) query.set("page_token", params.page_token);
  const qs = query.toString();
  return request<ListOrdersResponse>(`/orders${qs ? `?${qs}` : ""}`);
}
