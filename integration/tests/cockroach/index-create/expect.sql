create table "users" ("id" integer, "email" character varying (255) not null, "phone" character varying (10) not null default '', primary key ("id"), constraint "idx_users_email" unique ("email"));
