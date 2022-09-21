begin transaction;
alter table "user_projects" rename to "user_projects_e9b9cedb5139e31de07edaaf8e769c995e6ad4d33e44dc1bc6e647823ff82692";
create table "user_projects" ("user_id" integer not null, "project_id" integer not null);
insert into user_projects (user_id, project_id) select user_id, project_id from user_projects_e9b9cedb5139e31de07edaaf8e769c995e6ad4d33e44dc1bc6e647823ff82692;
drop table user_projects_e9b9cedb5139e31de07edaaf8e769c995e6ad4d33e44dc1bc6e647823ff82692;
commit;
