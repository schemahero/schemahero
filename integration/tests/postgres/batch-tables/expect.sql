create table "orders" ("id" integer, "user_id" integer not null, "total" decimal (10, 2), "created_at" timestamp, primary key ("id"));
create table "products" ("id" integer, "name" varchar (255) not null, "price" decimal (10, 2), "sku" varchar (50) not null, primary key ("id"));
create table "users" ("id" integer, "email" varchar (255) not null, "name" varchar (255), primary key ("id"))
