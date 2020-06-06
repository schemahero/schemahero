# Airline

The manifests in this repo support the tutorial at https://schemahero.io/learn/tutorial/introduction/.

The files in this directory contain a very simple, fictitious airline database schema to manage reservations and flight schedules.

## Installing

To install this example, you'll need a Kubernetes cluster. The YAML in this example has been hard coded to run in a namespace named "schemahero-tutorial". The manifests here contain a postgresql 11.8.0 engine, and the SchemaHero database and table components. This directory does not contain the SchemaHero Operator; you'll have to install this yourself or already have it in the cluster.

### Database

To install the PostgresQL 11.8.0 instance, run the following commands from this (examples/airline) directory:

```shell
kubectl apply -f ./postgresql/postgresql-11.8.0.yaml
```

### Schema

After the database engine has been deployed, the schema directory can be deployed and SchemaHero will create the tables.

#### Database Connection

To start, deploy the database connection:

```
kubectl apply -f ./schema/airlinedb.yaml
```

#### Tables

Deplioy the tables:

```
kubectl apply -f ./schema
```

### Verifying

You can now connect to the database and explore the schema that was created:

```shell
kubectl schemahero shell --namespace schemahero-tutorial airlinedb
```

The above command will start a new pod and connect your terminal to it. You'll be authenticated to the database, and the `airlinedb` will already be selected.

Try viewing the schema of the reservation table:

```
\d reservation
                          Table "public.reservation"
    Column     |            Type             | Collation | Nullable | Default
---------------+-----------------------------+-----------+----------+---------
 id            | character(8)                |           | not null |
 created_at    | timestamp without time zone |           | not null |
 first_name    | character varying(40)       |           |          |
 last_name     | character varying(40)       |           |          |
 flight_number | character(4)                |           |          |
 flight_date   | timestamp without time zone |           | not null |
Indexes:
    "reservation_pkey" PRIMARY KEY, btree (id)
```
