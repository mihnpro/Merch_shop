#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"
mkdir -p AccountInternal
rm -f AccountInternal/*.pb.go
protoc \
    --proto_path=AccountInternal \
    --go_out=AccountInternal \
    --go_opt=paths=source_relative \
    --go-grpc_out=AccountInternal \
    --go-grpc_opt=paths=source_relative \
    AccountInternal/*.proto
echo "OK: generated in $SCRIPT_DIR/AccountInternal"
ls -1 AccountInternal/*.pb.go
