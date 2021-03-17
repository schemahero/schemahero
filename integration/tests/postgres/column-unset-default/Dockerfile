FROM postgres:10.7

ENV POSTGRES_USER=schemahero
ENV POSTGRES_DB=schemahero

## Insert fixtures
COPY ./fixtures.sql /docker-entrypoint-initdb.d/
