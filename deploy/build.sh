#!/usr/bin/env bash
set -euo pipefail

# Build sequentially to avoid exhausting host memory during `go build`.
# Exporting the limit explicitly guarantees it applies regardless of which
# directory Compose is invoked from.
export COMPOSE_BUILD_PARALLEL_LIMIT=1

cd "$(dirname "$0")"

docker compose up --build -d
