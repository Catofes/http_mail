#!/usr/bin/env bash
set -euo pipefail

# Only a disposable local cluster is created. No production configuration is read.
if [[ -n "${HTTP_MAIL_TEST_DSN:-}" ]]; then
    go test -race -count=1 -run TestPostgresCompatibility -v ./...
    exit
fi

for tool in initdb pg_ctl go; do
    command -v "$tool" >/dev/null || { echo "Missing test dependency: $tool" >&2; exit 1; }
done

# Resolve symlinks so PostgreSQL can find its installation's share directory.
pg_bindir=$(dirname "$(readlink -f "$(command -v initdb)")")

pg_tmp=$(mktemp -d /tmp/http-mail-pg.XXXXXXXX)
cleanup() {
    "$pg_bindir/pg_ctl" -D "$pg_tmp/data" -m immediate -w stop >/dev/null 2>&1 || true
    rm -rf "$pg_tmp"
}
trap cleanup EXIT

if ! "$pg_bindir/initdb" -D "$pg_tmp/data" -A trust -U http_mail_test --no-locale -E UTF8 >"$pg_tmp/initdb.log" 2>&1; then
    cat "$pg_tmp/initdb.log" >&2
    exit 1
fi
if ! "$pg_bindir/pg_ctl" -D "$pg_tmp/data" -l "$pg_tmp/server.log" \
    -o "-k $pg_tmp -c listen_addresses='' -c fsync=off" -w start >"$pg_tmp/start.log" 2>&1; then
    cat "$pg_tmp/start.log" "$pg_tmp/server.log" >&2
    exit 1
fi
export HTTP_MAIL_TEST_DSN="host=$pg_tmp dbname=postgres user=http_mail_test sslmode=disable"
go test -race -count=1 -run TestPostgresCompatibility -v ./...
