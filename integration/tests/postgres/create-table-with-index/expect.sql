create table "other" ("id" integer, "something" character varying (255) not null, primary key ("id"));
create index idx_other_something on other (something);
create table "users" ("id" integer, "email" character varying (255) not null, "phone" character varying (10) not null default '', primary key ("id"), constraint "idx_users_email" unique ("email"));
