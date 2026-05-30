#!/bin/sh
set -e

chown -R appuser:appgroup /app/data /app/uploads /app/exports /app/plugins

exec su-exec appuser "$@"
