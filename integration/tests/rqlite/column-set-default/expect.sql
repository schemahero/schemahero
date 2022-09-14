alter table "users" rename to "users_177853ca2be414e548038166010e689fe31ede3391cd16bcd1baf7dbbc719bb7";
create table "users" ("id" integer not null, "email" text not null, "account_type" text default 'trial', "num_seats" integer default '5', primary key ("id"));
insert into users (id, email, account_type, num_seats) select id, email, account_type, num_seats from users_177853ca2be414e548038166010e689fe31ede3391cd16bcd1baf7dbbc719bb7;
drop table users_177853ca2be414e548038166010e689fe31ede3391cd16bcd1baf7dbbc719bb7;
