import { useEffect, useState } from "react";
import { grantPoints, listUsers } from "../../api/admin";
import type { AdminUser } from "../../api/types";

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

  async function handleGrant(user: AdminUser, amount: number, reason: string) {
    setError("");
    setInfo("");
    try {
      const bal = await grantPoints(user.id, {
        amount,
        reason,
        operation_id: crypto.randomUUID(),
      });
      setInfo(`Начислено ${amount} баллов пользователю ${user.login}. Новый баланс: ${bal.points}.`);
      setOpenId(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка");
    }
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
                onGrant={handleGrant}
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
  onGrant,
}: {
  user: AdminUser;
  open: boolean;
  onToggle: () => void;
  onGrant: (user: AdminUser, amount: number, reason: string) => void;
}) {
  const [amount, setAmount] = useState("");
  const [reason, setReason] = useState("");

  function submit(event: React.FormEvent) {
    event.preventDefault();
    const value = Number(amount);
    if (!Number.isInteger(value) || value <= 0) return;
    onGrant(user, value, reason.trim() || "начисление администратором");
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
        <td>{user.status}</td>
        <td className="actions">
          <button type="button" className="btn-secondary" onClick={onToggle}>
            {open ? "Отмена" : "Начислить баллы"}
          </button>
        </td>
      </tr>
      {open && (
        <tr>
          <td colSpan={6}>
            <form className="grant-form" onSubmit={submit}>
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
              <button type="submit">Начислить</button>
            </form>
          </td>
        </tr>
      )}
    </>
  );
}
