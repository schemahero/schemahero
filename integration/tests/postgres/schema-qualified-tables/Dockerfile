FROM postgres

ENV POSTGRES_USER=schemahero
ENV POSTGRES_DB=schemahero

COPY ./fixtures.sql /docker-entrypoint-initdb.d/
