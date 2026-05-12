FROM timescale/timescaledb:2.25.0-pg17

ENV POSTGRES_USER=schemahero
ENV POSTGRES_DB=schemahero

## Insert fixtures
COPY ./fixtures.sql /docker-entrypoint-initdb.d/
