#!/usr/bin/env bash
set -Eeo pipefail

if [ "${1-}" != "start" ]; then
    if [ "${1-}" = "shell" ]; then
        shift
        exec /bin/sh "$@"
    else
        exec /cockroach/cockroach "$@"
    fi
    return
fi

/cockroach/cockroach "$@" &
pid="$!"

# TODO: alternate port `--listen-addr <addr/host>[:<port>]`
csql=( /cockroach/cockroach sql --insecure --user=root --host=localhost )

# wait for cockroach to be available
for i in {30..0}; do
    if echo 'SELECT 1;' | "${csql[@]}" &> /dev/null; then
        break
    fi
    echo 'CockroachDB init process in progress...'
    sleep 1
done


"${csql[@]}" <<-EOF
CREATE USER $COCKROACH_USER;
GRANT SELECT ON DATABASE system TO $COCKROACH_USER;
CREATE DATABASE IF NOT EXISTS $COCKROACH_DATABASE;
GRANT ALL ON DATABASE $COCKROACH_DATABASE TO $COCKROACH_USER;
EOF

if [ -n "$COCKROACH_DATABASE" ]; then
    for f in /docker-entrypoint-initdb.d/*; do
        case "$f" in
            *.sh)
                # https://github.com/docker-library/postgres/issues/450#issuecomment-393167936
                # https://github.com/docker-library/postgres/pull/452
                if [ -x "$f" ]; then
                    echo "$0: running $f"
                    "$f"
                else
                    echo "$0: sourcing $f"
                    . "$f"
                fi
                ;;
            *.sql)    echo "$0: running $f"; "${csql[@]}" < "$f"; echo ;;
            *.sql.gz) echo "$0: running $f"; gunzip -c "$f" | "${csql[@]}"; echo ;;
            *)        echo "$0: ignoring $f" ;;
        esac
        echo
    done
fi

if ! kill -s TERM "$pid" || ! wait "$pid"; then
    echo >&2 'CockroachDB init process failed.'
    exit 1
fi

exec /cockroach/cockroach "$@"
