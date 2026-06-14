import { useEffect, useState } from "react";
import {
  blockUser,
  changeUserRole,
  getUserTransactions,
  grantPoints,
  listUsers,
  resetUserPassword,
} from "../../api/admin";
import type { AdminUser, Transaction } from "../../api/types";

export default function UsersPanel() {
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [search, setSearch] = useState("");
  const [error, setError] = useState("");
  const [info, setInfo] = useState("");
  const [openId, setOpenId] = useState<string | null>(null);

  async function reload(searchValue = "") {
    setError("");
    try {
      const res = await listUsers({ search: searchValue || undefined, page_size: 50 });
      setUsers(res.users ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка");
    }
  }

  useEffect(() => {
    reload();
  }, []);

  async function handleSearch(event: React.FormEvent) {
    event.preventDefault();
    await reload(search.trim());
  }

  function replaceUser(updated: AdminUser) {
    setUsers((prev) => prev.map((u) => (u.id === updated.id ? updated : u)));
  }

  return (
    <div>
      <form className="panel-form" onSubmit={handleSearch}>
        <h3>Пользователи</h3>
        <div className="row">
          <label>
            Поиск (логин / имя / email)
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="ivan"
            />
          </label>
        </div>
        <button type="submit">Найти</button>
      </form>

      {error && <p className="error">{error}</p>}
      {info && <p className="info">{info}</p>}

      <div className="table-wrap">
        <table className="table">
          <thead>
            <tr>
              <th>Логин</th>
              <th>Имя</th>
              <th>Email</th>
              <th>Роль</th>
              <th>Статус</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {users.map((u) => (
              <UserRow
                key={u.id}
                user={u}
                open={openId === u.id}
                onToggle={() => setOpenId(openId === u.id ? null : u.id)}
                onUpdated={replaceUser}
                onInfo={setInfo}
                onError={setError}
              />
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function UserRow({
  user,
  open,
  onToggle,
  onUpdated,
  onInfo,
  onError,
}: {
  user: AdminUser;
  open: boolean;
  onToggle: () => void;
  onUpdated: (user: AdminUser) => void;
  onInfo: (msg: string) => void;
  onError: (msg: string) => void;
}) {
  const [amount, setAmount] = useState("");
  const [reason, setReason] = useState("");
  const [busy, setBusy] = useState(false);
  const [txs, setTxs] = useState<Transaction[] | null>(null);

  const blocked = user.status === "blocked";

  function notify(fn: () => Promise<void>) {
    onError("");
    onInfo("");
    setBusy(true);
    fn()
      .catch((err) => onError(err instanceof Error ? err.message : "Ошибка"))
      .finally(() => setBusy(false));
  }

  function handleGrant(event: React.FormEvent) {
    event.preventDefault();
    const value = Number(amount);
    if (!Number.isInteger(value) || value <= 0) return;
    notify(async () => {
      const bal = await grantPoints(user.id, {
        amount: value,
        reason: reason.trim() || "начисление администратором",
        operation_id: crypto.randomUUID(),
      });
      onInfo(`Начислено ${value} баллов пользователю ${user.login}. Новый баланс: ${bal.points}.`);
      setAmount("");
      setReason("");
    });
  }

  function handleBlock() {
    notify(async () => {
      const updated = await blockUser(user.id, !blocked);
      onUpdated(updated);
      onInfo(`Пользователь ${user.login} ${updated.status === "blocked" ? "заблокирован" : "разблокирован"}.`);
    });
  }

  function handleRole() {
    const nextRole = user.role === "admin" ? "user" : "admin";
    notify(async () => {
      const updated = await changeUserRole(user.id, nextRole);
      onUpdated(updated);
      onInfo(`Роль пользователя ${user.login} изменена на «${updated.role}».`);
    });
  }

  function handleReset() {
    notify(async () => {
      const res = await resetUserPassword(user.id);
      onInfo(`Новый пароль для ${user.login}: ${res.new_password} (покажите пользователю и не сохраняйте).`);
    });
  }

  function handleLoadTxs() {
    notify(async () => {
      const res = await getUserTransactions(user.id, { page_size: 25 });
      setTxs(res.transactions ?? []);
    });
  }

  return (
    <>
      <tr>
        <td>{user.login}</td>
        <td>
          {user.last_name} {user.first_name}
        </td>
        <td>{user.email}</td>
        <td>{user.role}</td>
        <td>
          <span className={blocked ? "status-badge status-cancelled" : "status-badge status-delivered"}>
            {blocked ? "заблокирован" : "активен"}
          </span>
        </td>
        <td className="actions">
          <button type="button" className="btn-secondary" onClick={onToggle}>
            {open ? "Свернуть" : "Управление"}
          </button>
        </td>
      </tr>
      {open && (
        <tr>
          <td colSpan={6}>
            <div className="actions" style={{ marginBottom: 12, flexWrap: "wrap" }}>
              <button type="button" className="btn-secondary" disabled={busy} onClick={handleBlock}>
                {blocked ? "Разблокировать" : "Заблокировать"}
              </button>
              <button type="button" className="btn-secondary" disabled={busy} onClick={handleRole}>
                {user.role === "admin" ? "Снять админа" : "Сделать админом"}
              </button>
              <button type="button" className="btn-secondary" disabled={busy} onClick={handleReset}>
                Сбросить пароль
              </button>
              <button type="button" className="btn-secondary" disabled={busy} onClick={handleLoadTxs}>
                История операций
              </button>
            </div>

            <form className="grant-form" onSubmit={handleGrant}>
              <label>
                Сумма баллов
                <input
                  type="number"
                  min={1}
                  step={1}
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  required
                />
              </label>
              <label>
                Причина
                <input
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  placeholder="бонус за активность"
                />
              </label>
              <button type="submit" disabled={busy}>
                Начислить
              </button>
            </form>

            {txs && (
              <div className="table-wrap" style={{ marginTop: 12 }}>
                {txs.length === 0 ? (
                  <p className="muted">Операций нет</p>
                ) : (
                  <table className="table">
                    <thead>
                      <tr>
                        <th>Дата</th>
                        <th>Сумма</th>
                        <th>Причина</th>
                      </tr>
                    </thead>
                    <tbody>
                      {txs.map((t) => (
                        <tr key={t.id}>
                          <td>{new Date(t.created_at).toLocaleString("ru-RU")}</td>
                          <td style={{ color: t.amount >= 0 ? "#1a7f37" : "#cf222e" }}>
                            {t.amount > 0 ? `+${t.amount}` : t.amount}
                          </td>
                          <td>{t.reason}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}
              </div>
            )}
          </td>
        </tr>
      )}
    </>
  );
}
