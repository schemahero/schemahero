begin transaction;
alter table "projects" rename to "projects_c1372ea07c32b6c87d5e43d073ec5c3852c079e383666cf9c203fcb13dda0d39";
create table "projects" ("id" integer not null, "name" text not null, "icon_uri" text, primary key ("id"));
insert into projects (id, name, icon_uri) select id, name, icon_uri from projects_c1372ea07c32b6c87d5e43d073ec5c3852c079e383666cf9c203fcb13dda0d39;
drop table projects_c1372ea07c32b6c87d5e43d073ec5c3852c079e383666cf9c203fcb13dda0d39;
commit;
