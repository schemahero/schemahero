create table "orders" ("id" integer, "user_id" integer not null, "total" numeric (10, 2), "created_at" timestamp, primary key ("id"));
create table "products" ("id" integer, "name" character varying (255) not null, "price" numeric (10, 2), "sku" character varying (50) not null, primary key ("id"));
create table "users" ("id" integer, "email" character varying (255) not null, "name" character varying (255), primary key ("id"));
