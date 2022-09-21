alter table "users" rename to "users_e0b541947f3a981629f9b7b96ff2f2ce47b11d6cf5d2770fd705855a5f9b26f3";
create table "users" ("id" integer not null, "email" text not null, "account_type" text, "num_seats" integer, primary key ("id"));
insert into users (id, email, account_type, num_seats) select id, email, account_type, num_seats from users_e0b541947f3a981629f9b7b96ff2f2ce47b11d6cf5d2770fd705855a5f9b26f3;
drop table users_e0b541947f3a981629f9b7b96ff2f2ce47b11d6cf5d2770fd705855a5f9b26f3;
