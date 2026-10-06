#!/bin/sh
set -eu
# With no external URL, use the local Compose DB/migration service.
if [ "${DBMATE_TARGET:-local}" = "local" ]; then
    exec docker compose run --rm dbmate "$@"
fi
# Pass the database URI through environment, not command-line arguments.
export DATABASE_URL="${MIGRATION_DATABASE_URL:-${DATABASE_URL:-}}"
export DBMATE_SCHEMA_FILE=/data/schema.sql
export DBMATE_MIGRATIONS_DIR=/data/migrations
# DBMATE_NETWORK is intentionally word-split to support Docker network options.
exec docker run --rm ${DBMATE_NETWORK:---network=host} \
    --user "$(id -u):$(id -g)" \
    -e DATABASE_URL -e DBMATE_SCHEMA_FILE -e DBMATE_MIGRATIONS_DIR \
    -v "$(pwd)/db:/data" \
    "ghcr.io/amacneil/dbmate:${DBMATE_VERSION:-2.16.0}" "$@"
