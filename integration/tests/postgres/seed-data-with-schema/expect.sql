create table "table1" ("id" integer, "col1" character varying (255) not null, "col2" timestamp, primary key ("id"));
insert into table1 (id, col1, col2) values (1, 'seed-value', '2024-01-01T00:00:00Z') on conflict ("id") do update set (id, col1, col2) = (excluded.id, excluded.col1, excluded.col2);
