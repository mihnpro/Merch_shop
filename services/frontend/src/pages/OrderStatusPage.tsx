import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import NavBar from "../components/NavBar";
import { getOrder } from "../api/orders";
import type { OrderStatus, OrderView } from "../api/types";

const POLL_INTERVAL_MS = 500;
const TERMINAL: OrderStatus[] = ["delivered", "cancelled"];

const STATUS_LABEL: Record<OrderStatus, string> = {
  pending: "Оформлен, ожидает резерва",
  confirmed: "Подтверждён, товар зарезервирован",
  ready_to_pickup: "Готов к выдаче",
  delivered: "Выдан",
  cancelled: "Отменён",
};

const STATUS_HINT: Record<OrderStatus, string> = {
  pending: "Резервируем товары на складе — статус обновится автоматически в течение нескольких секунд.",
  confirmed: "Баллы списаны, товары зарезервированы.",
  ready_to_pickup: "Можно забирать заказ.",
  delivered: "Заказ выдан. Спасибо за покупку!",
  cancelled: "Заказ отменён, баллы возвращены на баланс.",
};

export default function OrderStatusPage() {
  const { id } = useParams<{ id: string }>();
  const [order, setOrder] = useState<OrderView | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!id) return;
    let cancelled = false;
    let timer: ReturnType<typeof setInterval> | undefined;

    async function poll() {
      try {
        const view = await getOrder(id!);
        if (cancelled) return;
        setOrder(view);
        if (TERMINAL.includes(view.status) && timer) {
          clearInterval(timer);
          timer = undefined;
        }
      } catch (e) {
        if (cancelled) return;
        setError(e instanceof Error ? e.message : "Ошибка");
      }
    }

    void poll();
    timer = setInterval(poll, POLL_INTERVAL_MS);

    return () => {
      cancelled = true;
      if (timer) clearInterval(timer);
    };
  }, [id]);

  return (
    <div className="page">
      <NavBar />
      <Link to="/orders" className="back-link">
        ← Мои заказы
      </Link>
      <h1>Заказ</h1>

      {error && <p className="error">{error}</p>}

      {!order ? (
        <p className="info">Загрузка…</p>
      ) : (
        <>
          <div className="order-status-head">
            <span className={`status-badge status-${order.status}`}>
              {STATUS_LABEL[order.status]}
            </span>
            <span className="muted">№ {order.id}</span>
          </div>
          <p className="muted">{STATUS_HINT[order.status]}</p>

          <div className="table-wrap">
            <table className="table">
              <thead>
                <tr>
                  <th>Товар</th>
                  <th>Кол-во</th>
                  <th>Цена, баллы</th>
                </tr>
              </thead>
              <tbody>
                {order.items.map((it) => (
                  <tr key={it.id}>
                    <td>{it.product_name}</td>
                    <td>{it.quantity}</td>
                    <td>{it.price_points}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <p>
            <strong>Адрес доставки:</strong> {order.delivery_address}
          </p>
          {order.note && (
            <p>
              <strong>Ваши заметки:</strong>
              <div style={{ whiteSpace: "pre-wrap", marginTop: 4, padding: 8, backgroundColor: "#f5f5f5", borderRadius: 4 }}>
                {order.note}
              </div>
            </p>
          )}
          {order.status === "cancelled" && order.cancel_reason && (
            <p>
              <strong>Причина отмены:</strong>{" "}
              <span style={{ whiteSpace: "pre-wrap" }}>{order.cancel_reason}</span>
            </p>
          )}
          <div className="cart-total">
            Итого: <strong>{order.total_points} баллов</strong>
          </div>
        </>
      )}
    </div>
  );
}
