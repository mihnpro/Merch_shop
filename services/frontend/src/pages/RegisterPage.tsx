import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { login, register } from "../api/auth";
import type { RegisterBody } from "../api/types";
import { useAuth } from "../lib/AuthContext";

const EMPTY_FORM: RegisterBody = {
  login: "",
  password: "",
  first_name: "",
  last_name: "",
  email: "",
  patronymic: "",
  phone_number: "",
};

export default function RegisterPage() {
  const [form, setForm] = useState<RegisterBody>(EMPTY_FORM);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const navigate = useNavigate();
  const { refresh } = useAuth();

  function handleChange(event: React.ChangeEvent<HTMLInputElement>) {
    const { name, value } = event.target;
    setForm((prev) => ({ ...prev, [name]: value }));
  }

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    setLoading(true);

    try {
      await register(form);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось зарегистрироваться");
      setLoading(false);
      return;
    }

    try {
      await login({ login: form.login, password: form.password });
      await refresh();
      navigate("/catalog");
    } catch {

      navigate("/login", { state: { registered: true } });
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="card">
      <h1>Регистрация</h1>
      <form onSubmit={handleSubmit}>
        <label>
          Логин
          <input name="login" value={form.login} onChange={handleChange} required />
        </label>
        <label>
          Пароль (минимум 8 символов)
          <input
            type="password"
            name="password"
            value={form.password}
            onChange={handleChange}
            minLength={8}
            required
          />
        </label>
        <label>
          Имя
          <input name="first_name" value={form.first_name} onChange={handleChange} required />
        </label>
        <label>
          Фамилия
          <input name="last_name" value={form.last_name} onChange={handleChange} required />
        </label>
        <label>
          Email
          <input
            type="email"
            name="email"
            value={form.email}
            onChange={handleChange}
            required
          />
        </label>
        <label>
          Отчество (необязательно)
          <input name="patronymic" value={form.patronymic} onChange={handleChange} />
        </label>
        <label>
          Телефон (необязательно)
          <input
            name="phone_number"
            value={form.phone_number}
            onChange={handleChange}
            placeholder="+71234567890"
          />
        </label>

        {error && <p className="error">{error}</p>}

        <button type="submit" disabled={loading}>
          {loading ? "Регистрируем…" : "Зарегистрироваться"}
        </button>
      </form>

      <p className="hint">
        Уже есть аккаунт? <Link to="/login">Войти</Link>
      </p>
    </div>
  );
}
