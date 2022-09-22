begin transaction;
alter table "users" rename to "users_b2c1a24f693cd984c516dc5a970c88b6ea2b08b29dd73c45ea98111c7ea82eee";
create table "users" ("id" integer not null, "name" text, "age" integer, primary key ("id"));
insert into users (id, name, age) select id, name, age from users_b2c1a24f693cd984c516dc5a970c88b6ea2b08b29dd73c45ea98111c7ea82eee;
drop table users_b2c1a24f693cd984c516dc5a970c88b6ea2b08b29dd73c45ea98111c7ea82eee;
commit;
