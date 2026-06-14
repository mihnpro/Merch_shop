import { useEffect, useState } from "react";
import NavBar from "../components/NavBar";
import {
  changeMyPassword,
  getMyBalance,
  getMyProfile,
  getMyTransactions,
  updateMyProfile,
} from "../api/auth";
import type { Transaction, UserProfile } from "../api/types";

const PAGE_SIZE = 25;

export default function ProfilePage() {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [balance, setBalance] = useState<number | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    getMyProfile()
      .then(setProfile)
      .catch((err) => setError(err instanceof Error ? err.message : "Ошибка"));
    getMyBalance()
      .then((b) => setBalance(b.points))
      .catch(() => {});
  }, []);

  return (
    <div className="page">
      <NavBar />
      <h1>Профиль</h1>

      {error && <p className="error">{error}</p>}

      {!profile ? (
        !error && <p className="info">Загрузка…</p>
      ) : (
        <>
          <div className="row" style={{ alignItems: "stretch", flexWrap: "wrap" }}>
            <ProfileForm profile={profile} onSaved={setProfile} />
            <PasswordForm />
          </div>

          <div className="panel-form">
            <h3>Баланс</h3>
            <p>
              Текущий баланс:{" "}
              <strong>{balance === null ? "…" : `${balance} баллов`}</strong>
            </p>
            <p className="muted">Логин: {profile.login}</p>
            <p className="muted">Роль: {profile.role}</p>
          </div>

          <TransactionsSection />
        </>
      )}
    </div>
  );
}

function ProfileForm({
  profile,
  onSaved,
}: {
  profile: UserProfile;
  onSaved: (p: UserProfile) => void;
}) {
  const [firstName, setFirstName] = useState(profile.first_name);
  const [lastName, setLastName] = useState(profile.last_name);
  const [patronymic, setPatronymic] = useState(profile.patronymic ?? "");
  const [email, setEmail] = useState(profile.email);
  const [phone, setPhone] = useState(profile.phone_number ?? "");
  const [error, setError] = useState("");
  const [info, setInfo] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    setInfo("");
    setBusy(true);
    try {
      const updated = await updateMyProfile({
        first_name: firstName.trim(),
        last_name: lastName.trim(),
        patronymic: patronymic.trim() || undefined,
        email: email.trim(),
        phone_number: phone.trim() || undefined,
      });
      onSaved(updated);
      setInfo("Данные сохранены.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form className="panel-form" style={{ flex: "1 1 320px" }} onSubmit={submit}>
      <h3>Личные данные</h3>
      <div className="row">
        <label>
          Фамилия
          <input value={lastName} onChange={(e) => setLastName(e.target.value)} required />
        </label>
        <label>
          Имя
          <input value={firstName} onChange={(e) => setFirstName(e.target.value)} required />
        </label>
      </div>
      <div className="row">
        <label>
          Отчество
          <input value={patronymic} onChange={(e) => setPatronymic(e.target.value)} />
        </label>
      </div>
      <div className="row">
        <label>
          Email
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </label>
        <label>
          Телефон
          <input value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="+7..." />
        </label>
      </div>
      {error && <p className="error">{error}</p>}
      {info && <p className="info">{info}</p>}
      <button type="submit" disabled={busy}>
        {busy ? "Сохранение…" : "Сохранить"}
      </button>
    </form>
  );
}

function PasswordForm() {
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState("");
  const [info, setInfo] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    setInfo("");
    if (newPassword.length < 8) {
      setError("Новый пароль должен быть не короче 8 символов.");
      return;
    }
    if (newPassword !== confirm) {
      setError("Пароли не совпадают.");
      return;
    }
    setBusy(true);
    try {
      await changeMyPassword({ old_password: oldPassword, new_password: newPassword });
      setInfo("Пароль изменён.");
      setOldPassword("");
      setNewPassword("");
      setConfirm("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form className="panel-form" style={{ flex: "1 1 320px" }} onSubmit={submit}>
      <h3>Смена пароля</h3>
      <div className="row">
        <label>
          Текущий пароль
          <input
            type="password"
            value={oldPassword}
            onChange={(e) => setOldPassword(e.target.value)}
            required
          />
        </label>
      </div>
      <div className="row">
        <label>
          Новый пароль
          <input
            type="password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            minLength={8}
            required
          />
        </label>
        <label>
          Повторите пароль
          <input
            type="password"
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
            minLength={8}
            required
          />
        </label>
      </div>
      {error && <p className="error">{error}</p>}
      {info && <p className="info">{info}</p>}
      <button type="submit" disabled={busy}>
        {busy ? "Сохранение…" : "Сменить пароль"}
      </button>
    </form>
  );
}

function TransactionsSection() {
  const [txs, setTxs] = useState<Transaction[]>([]);
  const [nextToken, setNextToken] = useState<string | undefined>(undefined);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  async function load(token?: string) {
    setError("");
    try {
      const res = await getMyTransactions({ page_size: PAGE_SIZE, page_token: token });
      setTxs((prev) => (token ? [...prev, ...(res.transactions ?? [])] : res.transactions ?? []));
      setNextToken(res.next_page_token || undefined);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div>
      <h3>История операций по баллам</h3>
      {error && <p className="error">{error}</p>}
      {loading ? (
        <p className="info">Загрузка…</p>
      ) : txs.length === 0 ? (
        <p className="muted">Операций пока нет</p>
      ) : (
        <>
          <div className="table-wrap">
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
