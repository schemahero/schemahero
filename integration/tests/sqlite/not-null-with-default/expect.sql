begin transaction;
alter table "projects" rename to "projects_af391bf9bfeed3d6608216e29e9b14f7ad57b637cb9de9598ed4135294b7fe99";
create table "projects" ("id" integer not null, "name" text not null default 'unnamed', "icon_uri" text, primary key ("id"));
insert into projects (id, name, icon_uri) select id, name, icon_uri from projects_af391bf9bfeed3d6608216e29e9b14f7ad57b637cb9de9598ed4135294b7fe99;
drop table projects_af391bf9bfeed3d6608216e29e9b14f7ad57b637cb9de9598ed4135294b7fe99;
commit;
