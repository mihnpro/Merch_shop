import { useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { login } from "../api/auth";
import { saveTokens } from "../lib/tokens";

export default function LoginPage() {
  const location = useLocation();
  const justRegistered = (location.state as { registered?: boolean } | null)?.registered;

  const [loginValue, setLoginValue] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const navigate = useNavigate();

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    setLoading(true);
    try {
      const data = await login({ login: loginValue, password });
      saveTokens(data.tokens.access_token, data.tokens.refresh_token);
      navigate("/catalog");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось войти");
    } finally {
      setLoading(false);
    }
  }

  return (
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
  );
}
