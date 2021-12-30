create table "users" ("id" integer, "login" character varying (255), "name" character varying (255) not null, primary key ("id"));
insert into users (id, login, name) values (1, 'test', 'test2') on conflict ("id") do update set (id, login, name) = (excluded.id, excluded.login, excluded.name);
insert into users (id, login, name) values (2, 'other', 'test2') on conflict ("id") do update set (id, login, name) = (excluded.id, excluded.login, excluded.name);
insert into users (id, login, name) values (3, 'yet', 'someone') on conflict ("id") do update set (id, login, name) = (excluded.id, excluded.login, excluded.name);
insert into users (id, login, name) values (4, 'another', 'more') on conflict ("id") do update set (id, login, name) = (excluded.id, excluded.login, excluded.name);
