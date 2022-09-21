create table "users" ("id" integer not null, "login" text, "name" text not null, primary key ("id"));
replace into users (id, login, name) values (1, 'test', 'test2');
