FROM pgvector/pgvector:pg17

ENV POSTGRES_USER=schemahero
ENV POSTGRES_DB=schemahero

# Copy fixtures
COPY ./fixtures.sql /docker-entrypoint-initdb.d/
