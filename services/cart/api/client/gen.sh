#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

PROTO_SRC_DIR="./product"
GEN_OUT_DIR="./product"


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


mkdir -p "$GEN_OUT_DIR"
rm -f "${GEN_OUT_DIR:?}"/*.pb.go

protoc \
    --proto_path="$PROTO_SRC_DIR" \
    --go_out="$GEN_OUT_DIR" \
    --go_opt=paths=source_relative \
    --go-grpc_out="$GEN_OUT_DIR" \
    --go-grpc_opt=paths=source_relative \
    "$PROTO_SRC_DIR"/*.proto

echo "OK: сгенерировано в $SCRIPT_DIR/$GEN_OUT_DIR"
ls -1 "$GEN_OUT_DIR"
