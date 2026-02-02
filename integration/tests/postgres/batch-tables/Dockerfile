ARG PG_VERSION
FROM postgres:${PG_VERSION}
COPY ./fixtures.sql /docker-entrypoint-initdb.d/
