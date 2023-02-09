insert into other (id, something) values (1, 'one') on conflict ("id") do update set (id, something) = (excluded.id, excluded.something);
insert into other (id, something) values (2, 'two') on conflict ("id") do update set (id, something) = (excluded.id, excluded.something);
create table "users" ("created_at" timestamp without time zone not null, "id" integer, "login" character varying (255), "name" character varying (255) not null default 'ethan', "tz_1" timestamp, "tz_2" timestamp with time zone, "tz_3" timestamp without time zone, primary key ("created_at", "id"));
select create_hypertable('users', 'created_at');
