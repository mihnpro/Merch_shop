import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { useAuth } from "./AuthContext";
import * as cartApi from "../api/cart";
import type { CartView } from "../api/types";

const EMPTY_CART: CartView = { items: [], total: 0, item_count: 0 };

interface CartContextValue {
  cart: CartView | null;
  refresh: () => Promise<void>;
  addItem: (productId: string, qty: number) => Promise<void>;
  updateItem: (itemId: string, qty: number) => Promise<void>;
  removeItem: (itemId: string) => Promise<void>;
  clearCart: () => Promise<void>;
}

const CartContext = createContext<CartContextValue | null>(null);

export function CartProvider({ children }: { children: ReactNode }) {
  const { status } = useAuth();
  const [cart, setCart] = useState<CartView | null>(null);

  const refresh = useCallback(async () => {
    try {
      const data = await cartApi.getCart();
      setCart(data);
    } catch {
      setCart(EMPTY_CART);
    }
  }, []);

  useEffect(() => {
    if (status === "authenticated") {
      void refresh();
    } else if (status === "anonymous") {
      setCart(null);
    }
  }, [status, refresh]);

  const addItem = useCallback(
    async (productId: string, qty: number) => {
      await cartApi.addItem(productId, qty);
      await refresh();
    },
    [refresh],
  );

  const updateItem = useCallback(
    async (itemId: string, qty: number) => {
      await cartApi.updateItem(itemId, qty);
      await refresh();
    },
    [refresh],
  );

  const removeItem = useCallback(
    async (itemId: string) => {
      await cartApi.removeItem(itemId);
      await refresh();
    },
    [refresh],
  );

  const clearCart = useCallback(async () => {
    await cartApi.clearCart();
    await refresh();
  }, [refresh]);

  const value: CartContextValue = { cart, refresh, addItem, updateItem, removeItem, clearCart };

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
}

export function useCart(): CartContextValue {
  const ctx = useContext(CartContext);
  if (!ctx) throw new Error("useCart must be used within CartProvider");
  return ctx;
}
