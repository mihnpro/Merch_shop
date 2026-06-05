# MerchShop — фронтенд (регистрация и вход)

Простой учебный фронтенд на **React + TypeScript + Vite**. Только регистрация, вход и страница профиля.
Никаких UI-библиотек и стейт-менеджеров — обычный `fetch`, `useState`/`useEffect` и `react-router`.

## Запуск

1. Подними бэкенд (из каталога `services/`): сервис `user_customer` (gRPC `:50051`, нужны его Redis/Postgres)
   и `gateway` (HTTP `:8080`). Проверь, что gateway отвечает:
   ```bash
   curl -i -X POST localhost:8080/api/v1/auth/login
   ```
   Ждём какой-нибудь HTTP-ответ (4xx — это нормально), а не «connection refused».

2. Запусти фронт:
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
   Открой http://localhost:5173

## Как это работает

- Фронт ходит на gateway **напрямую** по адресу из переменной `VITE_API_URL`
  (по умолчанию `http://localhost:8080`, см. `src/api/client.ts`). Прокси не используется —
  CORS включён на самом gateway, поэтому браузер может обращаться к другому origin.
- Переменная `VITE_*` подставляется в код **на этапе сборки** (`npm run build` / `npm run dev`),
  а не в рантайме. Для docker это build-arg (см. `Dockerfile` и `docker-compose.yml`).
- После входа `access_token` и `refresh_token` сохраняются в `localStorage` (см. `src/lib/tokens.ts`).
  Каждый запрос автоматически получает заголовок `Authorization: Bearer <access_token>` (см. `src/api/client.ts`).

## Структура

```
src/
  main.tsx              точка входа, подключает роутер
  App.tsx               маршруты: /login, /register, /profile
  index.css             все стили
  api/
    types.ts            типы данных API
    client.ts           обёртка над fetch (базовый путь, токен, разбор ошибок)
    auth.ts             функции register / login / logout / getMe
  lib/
    tokens.ts           чтение/запись токенов в localStorage
  components/
    ProtectedRoute.tsx  пускает в /profile только при наличии токена
  pages/
    LoginPage.tsx       форма входа
    RegisterPage.tsx    форма регистрации
    ProfilePage.tsx     профиль (данные из GET /me) + кнопка «Выйти»
```

## Поток пользователя

1. `/register` → заполнил форму → бэкенд создаёт пользователя (токенов пока нет) → редирект на `/login`.
2. `/login` → ввёл логин и пароль → токены сохраняются → попадаем на `/profile`.
3. `/profile` → видим `user_id` и роль → «Выйти» гасит токен и возвращает на `/login`.

## Заметка про безопасность

Токены лежат в `localStorage` ради простоты. Это удобно для обучения, но уязвимо к XSS.
В продакшене обычно используют httpOnly-cookie. См. комментарий в `src/lib/tokens.ts`.
