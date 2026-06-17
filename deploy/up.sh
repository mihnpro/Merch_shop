#!/usr/bin/env bash
set -euo pipefail

DEPLOY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$DEPLOY_DIR/.." && pwd)"
cd "$ROOT_DIR"

echo "==> 1/5 Внешние зависимости (postgres+init / redis / kafka / minio + миграции)"
docker compose -f deploy/docker-compose.deps.yml up -d
echo "    ждём миграции..."
docker compose -f deploy/docker-compose.deps.yml logs -f \
  migrate-user migrate-product migrate-inventory migrate-cart migrate-order || true

echo "==> 2/5 minikube"
minikube status >/dev/null 2>&1 || minikube start

echo "==> 3/5 Сборка образов в docker-демон minikube"
eval "$(minikube docker-env)"
docker build -t user:latest      services/user
docker build -t gateway:latest   services/gateway
docker build -t product:latest   services/products
docker build -t media:latest     services/media
docker build -t inventory:latest services/invetory
docker build -t cart:latest      -f services/cart/Dockerfile  services
docker build -t order:latest     -f services/order/Dockerfile services
docker build -t frontend:latest --build-arg VITE_API_URL= services/frontend
eval "$(minikube docker-env -u)"

echo "==> 4/5 Применение манифестов"
kubectl apply -f deploy/namespace.yaml
for svc in user gateway product media inventory cart order frontend; do
  kubectl apply -f "deploy/$svc/"
done

echo "==> 5/5 Ожидание готовности подов"
for svc in user gateway product media inventory cart order frontend; do
  kubectl -n merchshop rollout status "deploy/$svc" --timeout=180s
done

echo
echo "Готово. Поды:"
kubectl -n merchshop get pods
echo
echo "UI наружу (docker-драйвер под macOS) — открой в отдельном терминале:"
echo "  kubectl -n merchshop port-forward svc/frontend 8080:80"
echo "затем открой http://localhost:8080 в браузере"
echo
echo "Прямой доступ к API (для curl):"
echo "  kubectl -n merchshop port-forward svc/gateway 8080:8080"
echo
echo "Проверка сквозного маршрута:  bash deploy/verify.sh"
