alter table "users" rename to "users_662b1fbe4ce8b8e73cc961d3400d5dbcce5105681d2d5857568bfab4fdafb6b6";
create table "users" ("id" integer not null, primary key ("id"));
insert into users (id) select id from users_662b1fbe4ce8b8e73cc961d3400d5dbcce5105681d2d5857568bfab4fdafb6b6;
drop table users_662b1fbe4ce8b8e73cc961d3400d5dbcce5105681d2d5857568bfab4fdafb6b6;
