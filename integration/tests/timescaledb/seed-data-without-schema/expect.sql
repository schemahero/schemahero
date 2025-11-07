insert into users (id, login, name) values (1, 'test', 'test2') on conflict ("id") do update set (id, login, name) = (excluded.id, excluded.login, excluded.name);
insert into users (id, login, name) values (2, 'admin', 'Administrator') on conflict ("id") do update set (id, login, name) = (excluded.id, excluded.login, excluded.name);
