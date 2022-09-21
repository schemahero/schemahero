alter table "projects" rename to "projects_8c15d51ea0c81cfa97744f5daa3e4e60238741787765a30e610359e566dbfa1c";
create table "projects" ("id" integer not null, "name" text not null, primary key ("id"));
insert into projects (id, name) select id, name from projects_8c15d51ea0c81cfa97744f5daa3e4e60238741787765a30e610359e566dbfa1c;
drop table projects_8c15d51ea0c81cfa97744f5daa3e4e60238741787765a30e610359e566dbfa1c;
