#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

PROTO_DIR="server/AccountInternal"
WKT_INCLUDE="$(dirname "$(command -v protoc 2>/dev/null || echo /usr/bin/protoc)")/../include"


command -v protoc >/dev/null 2>&1 || {
    echo "ERROR: protoc не найден в PATH."
    echo "  macOS:  brew install protobuf"
    echo "  Linux:  apt-get install -y protobuf-compiler  (или пакет дистрибутива)"
    exit 1
}

command -v protoc-gen-go >/dev/null 2>&1 || {
    echo "ERROR: protoc-gen-go не найден в PATH."
    echo "  Установка: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    echo "  Убедитесь, что \$GOPATH/bin или \$GOBIN в PATH."
    exit 1
}

command -v protoc-gen-go-grpc >/dev/null 2>&1 || {
    echo "ERROR: protoc-gen-go-grpc не найден в PATH."
    echo "  Установка: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    exit 1
}


protoc \
    --proto_path="$PROTO_DIR" \
    --proto_path="$WKT_INCLUDE" \
    --go_out="$PROTO_DIR" \
    --go_opt=paths=source_relative \
    --go-grpc_out="$PROTO_DIR" \
    --go-grpc_opt=paths=source_relative \
    "$PROTO_DIR"/*.proto

echo "OK: сгенерировано в $SCRIPT_DIR/$PROTO_DIR"
ls -1 "$PROTO_DIR"
