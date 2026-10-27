#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p .local/e2e
runtime_dir=$(mktemp -d "$PWD/.local/e2e/run.XXXXXX")
export DATA_DIR="$runtime_dir/data"
export PLUGIN_DIR="$runtime_dir/plugins"
export LISTEN_ADDR=127.0.0.1:4300
export ALLOWED_HOSTS=127.0.0.1,localhost
export PUBLIC_DIR="$PWD/frontend/dist"
unset AUTH_USER AUTH_PASSWORD DATABASE_PATH
exec .local/lector
