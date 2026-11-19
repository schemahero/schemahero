FROM mysql:8.0

ENV MYSQL_USER=schemahero
ENV MYSQL_PASSWORD=password
ENV MYSQL_DATABASE=schemahero
ENV MYSQL_RANDOM_ROOT_PASSWORD=1

## Insert fixtures
COPY ./fixtures.sql /docker-entrypoint-initdb.d/
