#!/usr/bin/env bash
set -euo pipefail

LOCAL_PORT="${LOCAL_PORT:-8080}"
URL="http://localhost:${LOCAL_PORT}"
COOKIES="$(mktemp)"
PASS="Passw0rd!123"
LOGIN="verify_user"

echo "==> Поды в namespace merchshop"
kubectl -n merchshop get pods

echo
echo "==> port-forward svc/gateway ${LOCAL_PORT} -> 8080 (в фоне)"
kubectl -n merchshop port-forward svc/gateway "${LOCAL_PORT}:8080" >/tmp/merchshop-pf.log 2>&1 &
PF_PID=$!
trap 'kill "$PF_PID" 2>/dev/null || true; rm -f "$COOKIES"' EXIT
for i in $(seq 1 20); do
  curl -sS -o /dev/null "$URL/" 2>/dev/null && break
  sleep 0.5
done

check() {
  local name="$1" path="$2"
  local code body
  body="$(curl -sS -b "$COOKIES" -w '\n%{http_code}' "$URL$path" 2>/dev/null || true)"
  code="$(printf '%s' "$body" | tail -n1)"
  body="$(printf '%s' "$body" | sed '$d' | head -c 200)"
  printf '  %-28s %-3s  %s\n' "$path -> $name" "$code" "$body"
}

echo
echo "==> Регистрация (идемпотентно; 409 если уже есть)"
curl -sS -o /dev/null -w '  register -> %{http_code}\n' -X POST "$URL/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d "{\"login\":\"$LOGIN\",\"password\":\"$PASS\",\"first_name\":\"Verify\",\"last_name\":\"User\",\"email\":\"verify_user@example.com\"}" || true

echo "==> Логин (сохраняем cookie access_token)"
curl -sS -o /dev/null -c "$COOKIES" -w '  login    -> %{http_code}\n' -X POST "$URL/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"login\":\"$LOGIN\",\"password\":\"$PASS\"}" || true

echo
echo "==> Проверка маршрутов gateway -> сервисы (ожидаем 2xx, НЕ 502/504):"
check "user"      "/api/v1/me"
check "product"   "/api/v1/products"
check "product"   "/api/v1/categories"
check "inventory" "/api/v1/stock"
check "cart"      "/api/v1/cart"
check "order"     "/api/v1/orders"

echo
echo "Как читать:"
echo "  2xx/4xx (JSON от сервиса) — маршрут и сам сервис живы."
echo "  502/504 Bad Gateway      — gateway не достучался до сервиса (проверь его Service/pod)."
echo "  401                      — cookie/JWT не приняты (проверь единый JWT_ACCESS_SECRET)."
