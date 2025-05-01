FROM postgres:11.6

ENV POSTGRES_USER=schemahero
ENV POSTGRES_PASSWORD=password
ENV POSTGRES_DB=schemahero

COPY ./fixtures.sql /docker-entrypoint-initdb.d/
