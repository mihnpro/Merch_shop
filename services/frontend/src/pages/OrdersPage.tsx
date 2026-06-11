import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import NavBar from "../components/NavBar";
import { listMyOrders } from "../api/orders";
import type { OrderStatus, OrderView } from "../api/types";

const STATUS_LABEL: Record<OrderStatus, string> = {
  pending: "Ожидает резерва",
  confirmed: "Подтверждён",
  ready_to_pickup: "Готов к выдаче",
  delivered: "Выдан",
  cancelled: "Отменён",
};

export default function OrdersPage() {
  const [orders, setOrders] = useState<OrderView[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    listMyOrders({ page_size: 50 })
      .then((res) => {
        if (!cancelled) setOrders(res.orders ?? []);
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof Error ? e.message : "Ошибка");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="page">
      <NavBar />
      <h1>Мои заказы</h1>

      {error && <p className="error">{error}</p>}

      {loading ? (
        <p className="info">Загрузка…</p>
      ) : orders.length === 0 ? (
        <div className="cart-empty">
          <p className="muted">Заказов пока нет</p>
          <Link to="/catalog" className="back-link">
            ← Перейти в каталог
          </Link>
        </div>
      ) : (
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>Заказ</th>
                <th>Статус</th>
                <th>Сумма, баллы</th>
                <th>Создан</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {orders.map((o) => (
                <tr key={o.id}>
                  <td>{o.id.slice(0, 8)}…</td>
                  <td>
                    <span className={`status-badge status-${o.status}`}>
                      {STATUS_LABEL[o.status]}
                    </span>
                  </td>
                  <td>{o.total_points}</td>
                  <td>{new Date(o.created_at).toLocaleString("ru-RU")}</td>
                  <td className="actions">
                    <Link to={`/order/${o.id}`} className="btn-secondary">
                      Подробнее
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
