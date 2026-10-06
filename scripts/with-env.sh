#!/bin/sh
# .env uses shell KEY=value syntax. Quote values containing shell metacharacters.
set -eu
if [ -f .env ]; then
    set -a
    . ./.env
    set +a
fi
exec "$@"
