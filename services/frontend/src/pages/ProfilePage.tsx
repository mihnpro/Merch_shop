import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { getMe, logout } from "../api/auth";
import type { MeResponse } from "../api/types";
import { useAuth } from "../lib/AuthContext";

export default function ProfilePage() {
  const [me, setMe] = useState<MeResponse | null>(null);
  const [error, setError] = useState("");

  const navigate = useNavigate();
  const { setAnonymous } = useAuth();

  useEffect(() => {
    getMe()
      .then(setMe)
      .catch((err) => setError(err instanceof Error ? err.message : "Ошибка"));
  }, []);

  async function handleLogout() {
    try {
      await logout();
    } finally {
      setAnonymous();
      navigate("/login");
    }
  }

  return (
    <div className="card">
      <h1>Профиль</h1>

      {error && <p className="error">{error}</p>}

      {me ? (
        <ul className="profile">
          <li>
            <span>ID пользователя:</span> {me.user_id}
          </li>
          <li>
            <span>Роль:</span> {me.role}
          </li>
        </ul>
      ) : (
        !error && <p>Загрузка…</p>
      )}

      <button onClick={handleLogout}>Выйти</button>
    </div>
  );
}
