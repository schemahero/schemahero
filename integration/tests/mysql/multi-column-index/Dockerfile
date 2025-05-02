ARG MYSQL_VERSION=8.0
FROM mysql:${MYSQL_VERSION}

ENV MYSQL_USER=schemahero
ENV MYSQL_PASSWORD=password
ENV MYSQL_DATABASE=schemahero
ENV MYSQL_RANDOM_ROOT_PASSWORD=1

## Insert fixtures
COPY ./fixtures.sql /docker-entrypoint-initdb.d/
