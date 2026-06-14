import { request } from "./client";
import type { ListStockResponse } from "./types";


export async function getStock(productIds: string[]): Promise<Record<string, number>> {
  const ids = Array.from(new Set(productIds.filter(Boolean)));
  if (ids.length === 0) return {};
  const res = await request<ListStockResponse>(
    `/stock?product_ids=${encodeURIComponent(ids.join(","))}`,
    { method: "GET" },
  );
  const map: Record<string, number> = {};
  for (const item of res.items ?? []) {
    map[item.product_id] = item.available;
  }
  for (const id of ids) {
    if (!(id in map)) map[id] = 0;
  }
  return map;
}
