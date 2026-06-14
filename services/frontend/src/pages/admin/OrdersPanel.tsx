import { useCallback, useEffect, useState } from "react";
import {
  adminListOrders,
  adminUpdateOrderStatus,
  getUser,
} from "../../api/admin";
import type { OrderStatus, OrderView } from "../../api/types";

const STATUS_LABEL: Record<OrderStatus, string> = {
  pending: "Ожидает резерва",
  confirmed: "Подтверждён",
  ready_to_pickup: "Готов к выдаче",
  delivered: "Выдан",
  cancelled: "Отменён",
};

const TRANSITIONS: Record<OrderStatus, OrderStatus[]> = {
  pending: ["confirmed", "cancelled"],
  confirmed: ["ready_to_pickup", "cancelled"],
  ready_to_pickup: ["delivered"],
  delivered: [],
  cancelled: [],
};

const STATUS_FILTERS: { value: string; label: string }[] = [
  { value: "", label: "Все статусы" },
  { value: "pending", label: STATUS_LABEL.pending },
  { value: "confirmed", label: STATUS_LABEL.confirmed },
  { value: "ready_to_pickup", label: STATUS_LABEL.ready_to_pickup },
  { value: "delivered", label: STATUS_LABEL.delivered },
  { value: "cancelled", label: STATUS_LABEL.cancelled },
];

export default function OrdersPanel() {
  const [orders, setOrders] = useState<OrderView[]>([]);
  const [status, setStatus] = useState("");
  const [userId, setUserId] = useState("");
  const [nextToken, setNextToken] = useState<string | undefined>(undefined);
  const [error, setError] = useState("");
  const [info, setInfo] = useState("");
  const [loading, setLoading] = useState(true);
  const [openId, setOpenId] = useState<string | null>(null);
  const [names, setNames] = useState<Record<string, string>>({});

  const load = useCallback(
    async (token?: string, override?: { status?: string; userId?: string }) => {
      setError("");
      setLoading(true);
      try {
        const res = await adminListOrders({
          status: (override?.status ?? status) || undefined,
          user_id: (override?.userId ?? userId).trim() || undefined,
          page_size: 25,
          page_token: token,
        });
        setOrders((prev) => (token ? [...prev, ...(res.orders ?? [])] : res.orders ?? []));
        setNextToken(res.next_page_token || undefined);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Ошибка");
      } finally {
        setLoading(false);
      }
    },
    [status, userId],
  );

  useEffect(() => {
    load();
  }, []);

  const resolveName = useCallback(
    async (uid: string) => {
      if (names[uid]) return;
      try {
        const u = await getUser(uid);
        setNames((prev) => ({ ...prev, [uid]: `${u.last_name} ${u.first_name} (${u.login})` }));
      } catch {
        setNames((prev) => ({ ...prev, [uid]: uid }));
      }
    },
    [names],
  );

  function handleFilter(event: React.FormEvent) {
    event.preventDefault();
    setOpenId(null);
    load();
  }

  function replaceOrder(updated: OrderView) {
    setOrders((prev) => prev.map((o) => (o.id === updated.id ? updated : o)));
  }

  return (
    <div>
      <form className="panel-form" onSubmit={handleFilter}>
        <h3>Все заказы</h3>
        <div className="row">
          <label>
            Статус
            <select value={status} onChange={(e) => setStatus(e.target.value)}>
              {STATUS_FILTERS.map((s) => (
                <option key={s.value} value={s.value}>
                  {s.label}
                </option>
              ))}
            </select>
          </label>
          <label>
            ID пользователя
            <input
              value={userId}
              onChange={(e) => setUserId(e.target.value)}
              placeholder="uuid (необязательно)"
            />
          </label>
        </div>
        <button type="submit">Применить</button>
      </form>

      {error && <p className="error">{error}</p>}
      {info && <p className="info">{info}</p>}

      {loading && orders.length === 0 ? (
        <p className="info">Загрузка…</p>
      ) : orders.length === 0 ? (
        <p className="muted">Заказов нет</p>
      ) : (
        <>
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
                  <OrderRow
                    key={o.id}
                    order={o}
                    open={openId === o.id}
                    userName={names[o.user_id]}
                    onToggle={() => {
                      const willOpen = openId !== o.id;
                      setOpenId(willOpen ? o.id : null);
                      if (willOpen) resolveName(o.user_id);
                    }}
                    onUpdated={(u) => {
                      replaceOrder(u);
                      setInfo(`Заказ ${u.id.slice(0, 8)}… → «${STATUS_LABEL[u.status]}».`);
                    }}
                    onError={setError}
                  />
                ))}
              </tbody>
            </table>
          </div>
          {nextToken && (
            <button type="button" className="btn-secondary" onClick={() => load(nextToken)}>
              Показать ещё
            </button>
          )}
        </>
      )}
    </div>
  );
}

function OrderRow({
  order,
  open,
  userName,
  onToggle,
  onUpdated,
  onError,
}: {
  order: OrderView;
  open: boolean;
  userName?: string;
  onToggle: () => void;
  onUpdated: (order: OrderView) => void;
  onError: (msg: string) => void;
}) {
  const [reason, setReason] = useState("");
  const [busy, setBusy] = useState(false);
  const allowed = TRANSITIONS[order.status];

  function changeStatus(next: OrderStatus) {
    onError("");
    setBusy(true);
    adminUpdateOrderStatus(order.id, {
      status: next,
      reason: next === "cancelled" ? reason.trim() || "отменён администратором" : undefined,
    })
      .then(onUpdated)
      .catch((err) => onError(err instanceof Error ? err.message : "Ошибка"))
      .finally(() => setBusy(false));
  }

  return (
    <>
      <tr>
        <td>{order.id.slice(0, 8)}…</td>
        <td>
          <span className={`status-badge status-${order.status}`}>
            {STATUS_LABEL[order.status]}
          </span>
        </td>
        <td>{order.total_points}</td>
        <td>{new Date(order.created_at).toLocaleString("ru-RU")}</td>
        <td className="actions">
          <button type="button" className="btn-secondary" onClick={onToggle}>
            {open ? "Свернуть" : "Подробнее"}
          </button>
        </td>
      </tr>
      {open && (
        <tr>
          <td colSpan={5}>
            <p>
              <strong>Заказчик:</strong> {userName ?? "загрузка…"}
            </p>
            <p>
              <strong>Адрес доставки:</strong> {order.delivery_address}
            </p>
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

            {allowed.length === 0 ? (
              <p className="muted">Заказ в конечном статусе — изменения недоступны.</p>
            ) : (
              <div style={{ marginTop: 12 }}>
                {allowed.includes("cancelled") && (
                  <div className="row">
                    <label>
                      Причина отмены
                      <input
                        value={reason}
                        onChange={(e) => setReason(e.target.value)}
                        placeholder="нет в наличии"
                      />
                    </label>
                  </div>
                )}
                <div className="actions" style={{ flexWrap: "wrap" }}>
                  {allowed.map((next) => (
                    <button
                      key={next}
                      type="button"
                      className="btn-secondary"
                      disabled={busy}
                      onClick={() => changeStatus(next)}
                    >
                      → {STATUS_LABEL[next]}
                    </button>
                  ))}
                </div>
              </div>
            )}
          </td>
        </tr>
      )}
    </>
  );
}
