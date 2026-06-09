#!/bin/sh
set -e

chown -R lector:lector /app/data /app/uploads /app/exports /app/plugins

exec su-exec lector "$@"
