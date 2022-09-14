alter table "org" rename to "org_e14ee3e9b88284bec448aaf9d2272daff9d0057110f42e1a83ac5bc1ca0d30b0";
create table "org" ("id" integer not null, "project_id" integer, primary key ("id"));
insert into org (id, project_id) select id, project_id from org_e14ee3e9b88284bec448aaf9d2272daff9d0057110f42e1a83ac5bc1ca0d30b0;
drop table org_e14ee3e9b88284bec448aaf9d2272daff9d0057110f42e1a83ac5bc1ca0d30b0;
