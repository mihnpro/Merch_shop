import { useCallback, useEffect, useState } from "react";
import { getAnalytics, getUsersStats } from "../../api/admin";
import type { AnalyticsPeriod, AnalyticsView } from "../../api/types";

const PERIODS: { value: AnalyticsPeriod; label: string }[] = [
  { value: "day", label: "День" },
  { value: "week", label: "Неделя" },
  { value: "month", label: "Месяц" },
];

export default function AnalyticsPanel() {
  const [period, setPeriod] = useState<AnalyticsPeriod>("week");
  const [data, setData] = useState<AnalyticsView | null>(null);
  const [newUsers, setNewUsers] = useState<number | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  const load = useCallback(async (p: AnalyticsPeriod) => {
    setError("");
    setLoading(true);
    try {
      const [analytics, stats] = await Promise.all([getAnalytics(p), getUsersStats(p)]);
      setData(analytics);
      setNewUsers(stats.new_users);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load(period);
  }, [period, load]);

  return (
    <div>
      <div className="tabs" style={{ marginBottom: 16 }}>
        {PERIODS.map((p) => (
          <button
            key={p.value}
            type="button"
            className={period === p.value ? "tab active" : "tab"}
            onClick={() => setPeriod(p.value)}
          >
            {p.label}
          </button>
        ))}
      </div>

      {error && <p className="error">{error}</p>}
      {loading ? (
        <p className="info">Загрузка…</p>
      ) : data ? (
        <>
          <div className="stat-cards">
            <StatCard label="Заказов" value={data.orders_count} />
            <StatCard label="Списано баллов" value={data.points_spent} />
            <StatCard label="Средний чек" value={data.average_order_value} />
            <StatCard label="Новых сотрудников" value={newUsers ?? 0} />
          </div>

          <h3>Топ товаров</h3>
          {data.top_products.length === 0 ? (
            <p className="muted">Нет данных за период</p>
          ) : (
            <div className="table-wrap">
              <table className="table">
                <thead>
                  <tr>
                    <th>Товар</th>
                    <th>Продано, шт</th>
                  </tr>
                </thead>
                <tbody>
                  {data.top_products.map((p) => (
                    <tr key={p.product_id}>
                      <td>{p.product_name}</td>
                      <td>{p.quantity}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
      ) : null}
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <div className="stat-card">
      <span className="stat-value">{value}</span>
      <span className="stat-label">{label}</span>
    </div>
  );
}
