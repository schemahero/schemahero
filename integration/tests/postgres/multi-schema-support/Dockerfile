FROM postgres:11.6

# Using environment variables that will be provided at runtime
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=postgres
ENV POSTGRES_DB=postgres

COPY ./fixtures.sql /docker-entrypoint-initdb.d/
