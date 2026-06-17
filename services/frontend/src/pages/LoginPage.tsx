import { useState } from "react";
import { Link, Navigate, useLocation, useNavigate } from "react-router-dom";
import { login } from "../api/auth";
import { useAuth } from "../lib/AuthContext";

export default function LoginPage() {
  const location = useLocation();
  const { status, refresh } = useAuth();
  const justRegistered = (location.state as { registered?: boolean } | null)?.registered;

  const [loginValue, setLoginValue] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const navigate = useNavigate();

  if (status === "authenticated") {
    return <Navigate to="/catalog" replace />;
  }

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    setLoading(true);
    try {
      await login({ login: loginValue, password });
      await refresh();
      navigate("/catalog");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось войти");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="auth-screen">
      <aside className="auth-hero">
        <div className="auth-hero-logo">MerchShop</div>
        <p className="auth-hero-tagline">
          Магазин фирменного мерча. Стиль, который говорит за тебя.
        </p>
        <ul className="auth-hero-points">
          <li>Свежие дропы и лимитированные коллекции</li>
          <li>Баллы за покупки и бонусы</li>
          <li>Отслеживание заказа в реальном времени</li>
        </ul>
      </aside>
      <div className="auth-form-side">
        <div className="card">
          <h1>Вход</h1>
          {justRegistered && (
            <p className="info">Регистрация прошла успешно — теперь войдите.</p>
          )}
          <form onSubmit={handleSubmit}>
        <label>
          Логин
          <input
            value={loginValue}
            onChange={(e) => setLoginValue(e.target.value)}
            required
          />
        </label>
        <label>
          Пароль
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </label>

        {error && <p className="error">{error}</p>}

            <button type="submit" disabled={loading}>
              {loading ? "Входим…" : "Войти"}
            </button>
          </form>

          <p className="hint">
            Нет аккаунта? <Link to="/register">Зарегистрироваться</Link>
          </p>
        </div>
      </div>
    </div>
  );
}
