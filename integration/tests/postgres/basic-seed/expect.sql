create table "users" ("id" integer, "login" character varying (255), "name" character varying (255) not null, primary key ("id"));
insert into users (id, login, name) values (1, 'test', 'test2') on conflict do nothing;
