import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import NavBar from "../components/NavBar";
import { useCart } from "../lib/CartContext";
import { getProduct } from "../api/catalog";
import { getStock } from "../api/stock";
import { createOrder } from "../api/orders";
import { getMyBalance } from "../api/auth";
import { photoUrl } from "../lib/media";

export default function CartPage() {
  const navigate = useNavigate();
  const { cart, updateItem, removeItem, clearCart, refresh } = useCart();
  const [busy, setBusy] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [address, setAddress] = useState("");
  const [note, setNote] = useState("");
  const [balance, setBalance] = useState<number | null>(null);

  useEffect(() => {
    let cancelled = false;
    getMyBalance()
      .then((b) => {
        if (!cancelled) setBalance(b.points);
      })
      .catch(() => {
        if (!cancelled) setBalance(null);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const [covers, setCovers] = useState<Record<string, string | undefined>>({});
  const [stock, setStock] = useState<Record<string, number>>({});

  async function handleQty(itemId: string, newQty: number) {
    if (busy) return;
    setBusy(itemId);
    setError("");
    try {
      if (newQty <= 0) {
        await removeItem(itemId);
      } else {
        await updateItem(itemId, newQty);
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "Ошибка");
    } finally {
      setBusy(null);
    }
  }

  async function handleRemove(itemId: string) {
    if (busy) return;
    setBusy(itemId);
    setError("");
    try {
      await removeItem(itemId);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Ошибка");
    } finally {
      setBusy(null);
    }
  }

  async function handleClear() {
    if (busy) return;
    setBusy("clear");
    setError("");
    try {
      await clearCart();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Ошибка");
    } finally {
      setBusy(null);
    }
  }

  async function handleCheckout() {
    if (busy) return;
    const delivery = address.trim();
    if (!delivery) {
      setError("Укажите адрес доставки");
      return;
    }
    setBusy("checkout");
    setError("");
    try {
      const order = await createOrder(delivery, note.trim() || undefined);
      await refresh();
      navigate(`/order/${order.id}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Не удалось оформить заказ");
    } finally {
      setBusy(null);
    }
  }

  const items = cart?.items ?? [];

  const hasShortage = items.some(
    (i) => i.product_id in stock && i.quantity > stock[i.product_id],
  );

  const productKey = Array.from(new Set(items.map((i) => i.product_id))).sort().join(",");
  useEffect(() => {
    const ids = productKey ? productKey.split(",") : [];
    if (ids.length === 0) {
      setCovers({});
      setStock({});
      return;
    }
    let cancelled = false;
    Promise.all(
      ids.map((id) =>
        getProduct(id)
          .then((p) => [id, photoUrl(p.photo_keys[0])] as const)
          .catch(() => [id, undefined] as const),
      ),
    ).then((entries) => {
      if (cancelled) return;
      setCovers(Object.fromEntries(entries));
    });
    getStock(ids)
      .then((map) => {
        if (!cancelled) setStock(map);
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [productKey]);

  return (
    <div className="page">
      <NavBar />
      <h1>Корзина</h1>

      {error && <p className="error">{error}</p>}

      {items.length === 0 ? (
        <div className="cart-empty">
          <p className="muted">Корзина пуста</p>
          <Link to="/catalog" className="back-link">
            ← Перейти в каталог
          </Link>
        </div>
      ) : (
        <>
          <ul className="cart-list">
            {items.map((item) => {
              const avail = stock[item.product_id];
              const known = avail !== undefined;
              const over = known && item.quantity > avail;
              return (
              <li key={item.id} className={`cart-item${over ? " cart-item-over" : ""}`}>
                <Link to={`/catalog/${item.product_id}`} className="cart-item-photo">
                  {covers[item.product_id] ? (
                    <img src={covers[item.product_id]} alt={item.product_name} loading="lazy" />
                  ) : null}
                </Link>

                <div className="cart-item-info">
                  <Link to={`/catalog/${item.product_id}`} className="cart-item-name">
                    {item.product_name}
                  </Link>
                  {known && (
                    <span className={over ? "error" : "muted"}>
                      {over ? `В наличии только ${avail} шт` : `осталось ${avail} шт`}
                    </span>
                  )}
                </div>

                <div className="qty-controls">
                  <button
                    type="button"
                    onClick={() => handleQty(item.id, item.quantity - 1)}
                    disabled={!!busy}
                    aria-label="Уменьшить"
                  >
                    −
                  </button>
                  <span>{item.quantity}</span>
                  <button
                    type="button"
                    onClick={() => handleQty(item.id, item.quantity + 1)}
                    disabled={!!busy || (known && item.quantity >= avail)}
                    aria-label="Увеличить"
                  >
                    +
                  </button>
                </div>

                <div className="cart-item-price">
                  {item.price_per_unit} × {item.quantity} = {item.line_total} баллов
                </div>

                <button
                  type="button"
                  className="cart-item-remove"
                  onClick={() => handleRemove(item.id)}
                  disabled={!!busy}
                  aria-label="Удалить"
                >
                  ✕
                </button>
              </li>
              );
            })}
          </ul>

          <div className="cart-total">
            Итого: <strong>{cart!.total} баллов</strong>
          </div>

          {balance !== null && (
            <p className={`cart-balance${balance < cart!.total ? " cart-balance-low" : ""}`}>
              Доступно баллов: <strong>{balance}</strong>
              {balance < cart!.total && " — недостаточно для заказа"}
            </p>
          )}

          {hasShortage && (
            <p className="cart-balance cart-balance-low">
              Некоторых товаров недостаточно на складе — уменьшите количество
            </p>
          )}

          <div className="checkout-box">
            <label>
              Адрес доставки
              <input
                value={address}
                onChange={(e) => setAddress(e.target.value)}
                placeholder="Город, улица, дом, квартира"
                disabled={!!busy}
              />
            </label>
            <label>
              Заметки к заказу (не обязательно)
              <textarea
                value={note}
                onChange={(e) => setNote(e.target.value)}
                placeholder="Добавьте пожелания, инструкции или комментарии..."
                disabled={!!busy}
                rows={3}
              />
            </label>
          </div>

          <div className="cart-actions">
            <button
              type="button"
              className="btn-secondary"
              onClick={handleClear}
              disabled={!!busy}
            >
              Очистить корзину
            </button>
            <button
              type="button"
              onClick={handleCheckout}
              disabled={
                !!busy || hasShortage || (balance !== null && balance < cart!.total)
              }
            >
              {busy === "checkout" ? "Оформляем…" : "Заказать"}
            </button>
          </div>
        </>
      )}
    </div>
  );
}
